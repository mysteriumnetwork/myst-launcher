/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package myst

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	errors2 "github.com/pkg/errors"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	containerName             = "myst"
	reportLauncherVersionFlag = "--launcher.ver"

	operationTimeout = 10 * time.Second
)

var (
	ErrCouldNotConnect     = errors.New("could not connect to docker client")
	ErrCouldNotList        = errors.New("could not list containers")
	ErrContainerNotFound   = errors.New("could not find myst container")
	ErrContainerStart      = errors.New("could not start myst container")
	ErrCouldNotPullImage   = errors.New("could not pull myst image")
	ErrCouldNotCreateImage = errors.New("could not create myst image")
	ErrCouldNotStop        = errors.New("could not stop myst container")
	ErrCouldNotRemoveImage = errors.New("could not remove myst image")
)

type Manager struct {
	dockerAPI *client.Client
	//launcherCfg *model.Config
	model   *model.UIModel
	dataDir string
}

func NewManager(model *model.UIModel) (*Manager, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors2.Wrap(err, ErrCouldNotConnect.Error())
	}
	dataDir := fmt.Sprintf("%s.mysterium-node", utils.GetUserProfileDir()+string(os.PathSeparator))

	if err := utils.MakeDirectoryIfNotExists(dataDir); err != nil {
		return nil, err
	}
	return &Manager{
		dockerAPI: dc,
		model:     model,
		dataDir:   dataDir,
	}, nil
}

func (m *Manager) GetDockerClient() *client.Client {
	return m.dockerAPI
}

// Returns: alreadyRunning, error
func (m *Manager) Start() (bool, error) {
	log.Println("Start >")

	mystContainer, err := m.findMystContainer()
	if errors.Is(err, ErrContainerNotFound) {

		if err := m.pullMystLatest(); err != nil {
			return false, errors2.Wrap(err, "pullMystLatest")
		}
		if err := m.pullMystLatestByDigestLatest(); err != nil {
			return false, errors2.Wrap(err, "pullMystLatestByDigestLatest")
		}
		if err := m.createMystContainer(); err != nil {
			return false, errors2.Wrap(err, "createMystContainer")
		}
		mystContainer, err = m.findMystContainer()
	}
	if err != nil {
		log.Println("err >", err)
		return false, err
	}

	// refresh config if image has support of a ReportVersion option
	if m.model.CurrentImgHasReportVersionOption &&
		!strings.Contains(mystContainer.Command, reportLauncherVersionFlag) ||
		m.launcherVersionChanged(mystContainer) {

		return true, m.Restart()
	}

	if mystContainer.isRunning() {
		return true, nil
	}
	return false, m.startMystContainer()
}

// stop, apply settings and start
func (m *Manager) Restart() error {
	log.Println("Restart >")
	mystContainer, err := m.findMystContainer()
	if err != nil && err != ErrContainerNotFound {
		return err
	}

	// if found
	if mystContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		defer cancel()
		err = m.dockerAPI.ContainerStop(ctx, mystContainer.ID, nil)
		if err != nil {
			return errors2.Wrap(err, ErrCouldNotStop.Error())
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), operationTimeout)
		defer cancel2()
		err = m.dockerAPI.ContainerRemove(ctx2, mystContainer.ID, types.ContainerRemoveOptions{})
		if err != nil {
			return errors2.Wrap(err, ErrCouldNotRemoveImage.Error())
		}
	}

	if err = m.createMystContainer(); err != nil {
		return err
	}
	return m.startMystContainer()
}

func (m *Manager) Update() error {
	log.Println("Update >")

	// pull image by tag and by digest
	// b/c docker client api returns additional digest (manifest) for mult-iarch images

	if err := m.pullMystLatest(); err != nil {
		return errors2.Wrap(err, ("pullMystLatest"))
	}
	if err := m.pullMystLatestByDigestLatest(); err != nil {
		return errors2.Wrap(err, ("pullMystLatestByDigestLatest"))
	}

	return m.Restart()
}

func extractRepoDigests(repoDigests []string) []string {
	a := make([]string, 0)
	for _, d := range repoDigests {
		a = append(a, strings.Split(d, "@")[1])
	}
	return a
}

func (m *Manager) launcherVersionChanged(mystContainer *Container) bool {
	launcherVer := getVersionFromCommand(mystContainer.Command)
	currentVersion := m.model.ProductVersion + "/" + runtime.GOOS

	return launcherVer != currentVersion && launcherVer != ""
}

func getVersionFromCommand(cmd string) string {
	// log.Println("getVersionFromCommand:", cmd)

	set := &flag.FlagSet{}
	env := set.String("launcher.ver", "", "")

	args := strings.Split(cmd, " ")
	if len(args) > 1 {
		err := set.Parse(args[1:])
		if err == nil {
			return *env
		}
	}
	return ""
}

func (m *Manager) Stop() error {
	mystContainer, err := m.findMystContainer()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	err = m.dockerAPI.ContainerStop(ctx, mystContainer.ID, m.timeout())
	if err != nil {
		return errors2.Wrap(err, ErrCouldNotStop.Error())
	}
	return nil
}

//////////////////////////////////////////////////////////////////////
func (m *Manager) startMystContainer() error {
	log.Println("startMystContainer >")
	mystContainer, err := m.findMystContainer()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	err = m.dockerAPI.ContainerStart(ctx, mystContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors2.Wrap(err, ErrContainerStart.Error())
	}
	return nil
}

func (m *Manager) findMystContainer() (*Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	list, err := m.dockerAPI.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, errors2.Wrap(err, ErrCouldNotList.Error())
	}
	for idx, ctr := range list {
		for _, ctrName := range ctr.Names {
			if ctrName == "/"+containerName {
				return &Container{&list[idx]}, nil
			}
		}
	}

	return nil, ErrContainerNotFound
}

func (m *Manager) pullMystImage(image string) error {
	log.Println("pullMystImage >", image)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	out, err := m.dockerAPI.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return errors2.Wrap(err, ErrCouldNotPullImage.Error())
	}
	defer out.Close()

	_, err = io.Copy(ioutil.Discard, out)
	return err
}

func (m *Manager) pullMystLatestByDigestLatest() error {
	if m.model.ImageInfo.DigestLatest == "" {
		log.Println("pullMystByDigest > no DigestLatest !")
		return nil
	}

	image := "docker.io/" + m.model.Config.GetImageNamePrefix() + "@" + m.model.ImageInfo.DigestLatest
	return m.pullMystImage(image)
}

func (m *Manager) pullMystLatest() error {
	image := m.model.Config.GetFullImageName()
	return m.pullMystImage(image)
}

func (m *Manager) createMystContainer() error {
	log.Println("createMystContainer >")

	portSpecs := []string{
		"4449/tcp",
	}
	cmdArgs := []string{
		"service", "--agreed-terms-and-conditions",
	}
	if m.model.CurrentImgHasReportVersionOption {
		versionArg := fmt.Sprintf("%s=%s/%s", reportLauncherVersionFlag, m.model.ProductVersion, runtime.GOOS)
		cmdArgs = append([]string{versionArg}, cmdArgs...)
	}

	c := m.model.Config
	if c.EnablePortForwarding {
		p := fmt.Sprintf("%d-%d:%d-%d/udp", c.PortRangeBegin, c.PortRangeEnd, c.PortRangeBegin, c.PortRangeEnd)
		portSpecs = append(portSpecs, p)

		portsArg := fmt.Sprintf("--udp.ports=%d:%d", c.PortRangeBegin, c.PortRangeEnd)
		// prepend
		cmdArgs = append([]string{portsArg}, cmdArgs...)
	}

	exposedPorts, _, err := nat.ParsePortSpecs(portSpecs)
	if err != nil {
		return err
	}

	image := c.GetFullImageName()
	containerConfig := &container.Config{
		Image:        image,
		ExposedPorts: nat.PortSet(exposedPorts),
		Cmd:          strslice.StrSlice(cmdArgs),
	}
	log.Println("createMystContainer >", containerConfig)

	portMap := make(nat.PortMap)
	portMap["4449/tcp"] = []nat.PortBinding{
		{
			HostIP:   "0.0.0.0",
			HostPort: "4449",
		},
	}
	if c.EnablePortForwarding {
		for pn := c.PortRangeBegin; pn <= c.PortRangeEnd; pn++ {
			portMap[nat.Port(strconv.Itoa(pn)+"/udp")] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(pn),
				},
			}
		}
	}

	hostConfig := &container.HostConfig{
		CapAdd: strslice.StrSlice{
			"NET_ADMIN",
		},
		PortBindings: portMap,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: m.dataDir,
				Target: "/var/lib/mysterium-node",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	_, err = m.dockerAPI.ContainerCreate(ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return errors2.Wrap(err, ErrCouldNotCreateImage.Error())
	}
	return nil
}

func (m *Manager) timeout() *time.Duration {
	t := operationTimeout
	return &t
}

func (m *Manager) CheckCurrentVersionAndUpgrades(refreshVersionCache bool) {
	m.getCurrentImageDigest()
	updates.CheckVersionAndUpgrades(m.model, refreshVersionCache)
}

func (m *Manager) getCurrentImageDigest() {
	mystContainer, err := m.findMystContainer()
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	images, err := m.dockerAPI.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return
	}

	for _, i := range images {
		if i.ID == mystContainer.ImageID {
			log.Println("getCurrentImageDigest >", i.RepoDigests)
			m.model.ImageInfo.CurrentImgDigests = extractRepoDigests(i.RepoDigests)
		}
	}
}

// extend Container with method
type Container struct {
	*types.Container
}

func (c *Container) isRunning() bool {
	return c.State == "running"
}
