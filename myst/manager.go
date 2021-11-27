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
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	containerName = "myst"
	reportVerFlag = "--launcher.ver"
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

	defaultConfig = ManagerConfig{
		CTX:          context.Background(),
		ActionTimout: 10 * time.Second,
		DataDir:      fmt.Sprintf("%s.mysterium-node", utils.GetUserProfileDir()+string(os.PathSeparator)),
	}
)

type Manager struct {
	dockerAPI   *client.Client
	cfg         ManagerConfig
	launcherCfg *model.Config
}

type ManagerConfig struct {
	CTX          context.Context
	ActionTimout time.Duration
	DataDir      string
}

func NewManagerWithDefaults(launcherCfg *model.Config) (*Manager, error) {
	return NewManager(defaultConfig, launcherCfg)
}

func NewManager(cfg ManagerConfig, launcherCfg *model.Config) (*Manager, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, wrap(err, ErrCouldNotConnect)
	}

	if err := utils.MakeDirectoryIfNotExists(cfg.DataDir); err != nil {
		return nil, err
	}
	return &Manager{
		dockerAPI:   dc,
		cfg:         cfg,
		launcherCfg: launcherCfg,
	}, nil
}

func (m *Manager) CanPingDocker() bool {
	ctx, cancel := context.WithTimeout(m.cfg.CTX, 10*time.Second)
	defer cancel()

	_, err := m.dockerAPI.Ping(ctx)
	return err == nil
}

// Returns: alreadyRunning, error
func (m *Manager) Start(mm *model.UIModel) (bool, error) {

	mystContainer, err := m.findMystContainer()
	if errors.Is(err, ErrContainerNotFound) {
		if err := m.pullMystLatest(); err != nil {
			log.Println("pullMystLatest >", err)
			return false, err
		}
		if err := m.createMystContainer(mm); err != nil {
			fmt.Println("createMystContainer >", err)
			return false, err
		}
		mystContainer, err = m.findMystContainer()
	}
	if err != nil {
		return false, err
	}

	// container isn't running yet

	launcherVer := getVersionFromCommand(mystContainer.Command)
	currentVersion := mm.ProductVersion + "/" + runtime.GOOS
	launcherVersionChanged := launcherVer != currentVersion && launcherVer != ""

	// refresh config if image has support of a given option
	if mm.CurrentImgHasReportVersionAbility && !strings.Contains(mystContainer.Command, reportVerFlag) || launcherVersionChanged {
		return true, m.Restart(mm)
	}

	if mystContainer.IsRunning() {
		return true, nil
	}
	return false, m.startMystContainer()
}

func getVersionFromCommand(cmd string) string {
	fmt.Println(cmd)

	set := &flag.FlagSet{}
	env := set.String("launcher.ver", "", "")
	_ = env
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

	err = m.dockerAPI.ContainerStop(m.ctx(), mystContainer.ID, m.timeout())
	if err != nil {
		return wrap(err, ErrCouldNotStop)
	}
	return nil
}

// stop, apply settings and start
func (m *Manager) Restart(mm *model.UIModel) error {
	log.Println("Restart >")
	mystContainer, err := m.findMystContainer()
	if err != nil && err != ErrContainerNotFound {
		return err
	}

	err = m.dockerAPI.ContainerStop(m.ctx(), mystContainer.ID, nil)
	if err != nil {
		return wrap(err, ErrCouldNotStop)
	}
	err = m.dockerAPI.ContainerRemove(m.ctx(), mystContainer.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return wrap(err, ErrCouldNotRemoveImage)
	}

	err = m.createMystContainer(mm)
	if err != nil {
		return err
	}
	return m.startMystContainer()
}

func (m *Manager) Update(mm *model.UIModel) error {
	err := m.pullMystLatest()
	if err != nil {
		return err
	}

	mystContainer, err := m.findMystContainer()
	if err != nil && err != ErrContainerNotFound {
		return err
	}
	if !errors.Is(err, ErrContainerNotFound) {
		err = m.dockerAPI.ContainerStop(m.ctx(), mystContainer.ID, nil)
		if err != nil {
			return wrap(err, ErrCouldNotStop)
		}
		err = m.dockerAPI.ContainerRemove(m.ctx(), mystContainer.ID, types.ContainerRemoveOptions{})
		if err != nil {
			return wrap(err, ErrCouldNotRemoveImage)
		}
	}

	err = m.createMystContainer(mm)
	if err != nil {
		return err
	}
	err = m.startMystContainer()
	return err
}

//////////////////////////////////////////////////////////////////////
func wrap(external, internal error) error {
	return fmt.Errorf(external.Error()+": %w", internal)
}

func (m *Manager) startMystContainer() error {
	mystContainer, err := m.findMystContainer()
	if err != nil {
		return err
	}
	err = m.dockerAPI.ContainerStart(m.ctx(), mystContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		return wrap(err, ErrContainerStart)
	}
	return nil
}

func (m *Manager) findMystContainer() (*Container, error) {
	list, err := m.dockerAPI.ContainerList(m.ctx(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, wrap(err, ErrCouldNotList)
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

func (m *Manager) pullMystLatest() error {
	image := m.launcherCfg.GetFullImageName()
	out, err := m.dockerAPI.ImagePull(m.ctx(), image, types.ImagePullOptions{})
	if err != nil {
		return wrap(err, ErrCouldNotPullImage)
	}
	io.Copy(os.Stdout, out)
	return nil
}

func (m *Manager) createMystContainer(mm *model.UIModel) error {
	c := mm.Config
	log.Println("createMystContainer >")
	portSpecs := []string{
		"4449/tcp",
	}
	cmdArgs := []string{
		"service", "--agreed-terms-and-conditions",
	}
	if mm.CurrentImgHasReportVersionAbility {
		versionArg := fmt.Sprintf("%s=%s/%s", reportVerFlag, mm.ProductVersion, runtime.GOOS)
		cmdArgs = append([]string{versionArg}, cmdArgs...)
	}

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

	image := m.launcherCfg.GetFullImageName()
	config := &container.Config{
		Image:        image,
		ExposedPorts: nat.PortSet(exposedPorts),
		Cmd:          strslice.StrSlice(cmdArgs),
	}
	log.Println("createMystContainer >", config.Cmd)

	hostConfig := &container.HostConfig{
		CapAdd: strslice.StrSlice{
			"NET_ADMIN",
		},
		PortBindings: nat.PortMap{
			"4449/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "4449",
				},
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: m.cfg.DataDir,
				Target: "/var/lib/mysterium-node",
			},
		},
	}

	_, err = m.dockerAPI.ContainerCreate(m.ctx(),
		config,
		hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return wrap(err, ErrCouldNotCreateImage)
	}
	return nil
}

func (m *Manager) ctx() context.Context {
	return m.cfg.CTX
}

func (m *Manager) timeout() *time.Duration {
	return &m.cfg.ActionTimout
}

func (m *Manager) GetCurrentImageDigest() string {
	c, err := m.findMystContainer()
	if err != nil {
		return ""
	}

	images, err := m.dockerAPI.ImageList(m.ctx(), types.ImageListOptions{})
	if err != nil {
		return ""
	}

	imageDigest := ""
	for _, image := range images {
		if c.ImageID == image.ID {
			for _, rd := range image.RepoDigests {
				digestArr := strings.Split(rd, "@")
				imageDigest = digestArr[1]
			}
		}
	}
	return imageDigest
}

// extend Container with method
type Container struct {
	*types.Container
}

func (c *Container) IsRunning() bool {
	return c.State == "running"
}
