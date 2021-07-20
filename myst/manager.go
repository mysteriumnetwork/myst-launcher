package myst

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	imageName     = "mysteriumnetwork/myst:latest"
	containerName = "myst"
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
		DataDir:      fmt.Sprintf("%s\\.mysterium-node", os.Getenv("USERPROFILE")),
	}
)

type Manager struct {
	dockerAPI *client.Client
	cfg       ManagerConfig
}

type ManagerConfig struct {
	CTX          context.Context
	ActionTimout time.Duration
	DataDir      string
}

func NewManagerWithDefaults() (*Manager, error) {
	return NewManager(defaultConfig)
}

func NewManager(cfg ManagerConfig) (*Manager, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, wrap(err, ErrCouldNotConnect)
	}

	if err := utils.MakeDirectoryIfNotExists(cfg.DataDir); err != nil {
		return nil, err
	}
	return &Manager{
		dockerAPI: dc,
		cfg:       cfg,
	}, nil
}

func (m *Manager) CanPingDocker() bool {
	_, err := m.dockerAPI.Ping(m.ctx())
	if err != nil {
		return false
	}
	return true
}

func (m *Manager) Start() error {
	mystContainer, err := m.findMystContainer()
	if errors.Is(err, ErrContainerNotFound) {
		if err := m.pullMystLatest(); err != nil {
			return err
		}
		if err := m.createMystContainer(); err != nil {
			return err
		}
		mystContainer, err = m.findMystContainer()
	}
	if err != nil {
		return err
	}
	if mystContainer.IsRunning() {
		return nil
	}

	return m.startMystContainer()
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
	gui.UI.StateContainer = gui.RunnableStateUnknown
	gui.UI.Update()
	return nil
}

func (m *Manager) Update() error {
	mystContainer, err := m.findMystContainer()
	if err == nil {
		err = m.dockerAPI.ContainerRemove(m.ctx(), mystContainer.ID, types.ContainerRemoveOptions{})
		if err != nil {
			return wrap(err, ErrCouldNotRemoveImage)
		}
	}

	err = m.pullMystLatest()
	if err != nil {
		return err
	}

	err = m.createMystContainer()
	if err != nil {
		return err
	}

	err = m.startMystContainer()
	if err != nil {
		return err
	}

	return nil
}

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
	gui.UI.StateContainer = gui.RunnableStateStarting
	gui.UI.Update()
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
	out, err := m.dockerAPI.ImagePull(m.ctx(), imageName, types.ImagePullOptions{})
	if err != nil {
		return wrap(err, ErrCouldNotPullImage)
	}
	io.Copy(os.Stdout, out)
	return nil
}

func (m *Manager) createMystContainer() error {
	config := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"4449/tcp": struct{}{},
		},
		Cmd: strslice.StrSlice{
			"service",
			"--agreed-terms-and-conditions",
		},
	}
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

	_, err := m.dockerAPI.ContainerCreate(m.ctx(),
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
	c, _ := m.findMystContainer()

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

type Container struct {
	*types.Container
}

func (c *Container) IsRunning() bool {
	return c.State == "running"
}
