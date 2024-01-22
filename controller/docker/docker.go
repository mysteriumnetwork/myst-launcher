package docker

import (
	"log"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/controller/docker/myst"
	"github.com/mysteriumnetwork/myst-launcher/platform"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Docker_ struct {
	lg *log.Logger

	model       *model.UIModel
	ui          model.Gui_
	docker      *DockerRunner
	mystManager *myst.Manager
}

func NewSvc(m *model.UIModel, ui model.Gui_) *Docker_ {
	lg := log.New(log.Writer(), "[docker_] ", log.Ldate|log.Ltime)

	m.Caps = 1
	m.UIBus.Publish("model-change")

	mystManager, err := myst.NewManager(m)
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := NewDockerRunner(mystManager)

	return &Docker_{
		lg:          lg,
		model:       m,
		ui:          ui,
		docker:      docker,
		mystManager: mystManager,
	}
}

// docker
func (c *Docker_) TryStartRuntime() bool {

	if c.docker.IsRunning() {
		c.model.SetStateDocker(model.RunnableStateRunning)
		return true
	}

	ok := c.CheckSysRequirements()
	if ok {

		running, _ := c.docker.IsRunningOrTryStart()
		if running {
			c.model.SetStateContainer(model.RunnableStateRunning)
		} else {
			c.model.SetStateContainer(model.RunnableStateUnknown)
		}
		return true
	}
	return false
}

func (c *Docker_) CheckSysRequirements() bool {
	c.lg.Println("checkSysRequirements >")

	mdl := c.model
	ui := c.ui
	mgr, _ := platform.NewManager()

	// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
	isUnderVM, ok, err := false, true, error(nil)
	err = utils.Retry(3, time.Second, func() error {
		isUnderVM, err = mgr.SystemUnderVm()
		if err != nil {
			return err
		}

		featuresOk, err := mgr.Features()
		if err != nil {
			return err
		}
		if !featuresOk {
			ok = false
		}

		hasDocker, _ := utils.HasDocker()
		if !hasDocker {
			ok = false
		}
		return nil
	})
	if err != nil {
		c.lg.Println("error", err)
		ui.ErrorModal("Application error", err.Error())
		return false
	}

	if isUnderVM && !mdl.Config.CheckVMSettingsConfirm {
		ret := ui.YesNoModal("Requirements checker", "VM has been detected.\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
		if ret == model.IDNO {
			return false
		}
		mdl.Config.CheckVMSettingsConfirm = true
		mdl.Config.Save()
	}
	return ok
}

// check for image updates before starting container, offer upgrade interactively
func (c *Docker_) StartContainer() {
	mdl := c.model
	ui := c.ui

	c.mystManager.CheckCurrentVersionAndUpgrades(false)

	if !mdl.Config.Enabled {
		return
	}

	mdl.SetStateContainer(model.RunnableStateInstalling)
	containerAlreadyRunning, err := c.mystManager.Start()
	if err != nil {
		mdl.SetStateContainer(model.RunnableStateUnknown)
		c.lg.Println("startContainer", err)
		return
	}
	mdl.SetStateContainer(model.RunnableStateRunning)

	if !containerAlreadyRunning && mdl.Config.InitialState == model.InitialStateFirstRunAfterInstall {
		mdl.Config.InitialState = model.InitialStateNormalRun
		mdl.Config.Save()

		if ui != nil {
			ui.ShowNotificationInstalled()
			ui.OpenNodeUI()
		} else {
			c.lg.Println("node installed!")
		}
	}
}

func (c *Docker_) StopContainer() {
	mdl := c.model
	mdl.SetStateContainer(model.RunnableStateInstalling)

	err := c.mystManager.Stop()
	if err != nil {
		c.lg.Println("stop", err)
	}
}

func (c *Docker_) RestartContainer() {
	mdl := c.model

	mdl.SetStateContainer(model.RunnableStateInstalling)
	err := c.mystManager.Restart()
	if err != nil {
		c.lg.Println("restart", err)
	}
}

func (c *Docker_) UpgradeContainer(refreshVersionCache bool) {
	mdl := c.model

	if !mdl.ImageInfo.HasUpdate {
		return
	}

	if refreshVersionCache {
		c.mystManager.CheckCurrentVersionAndUpgrades(refreshVersionCache)
	}
	mdl.SetStateContainer(model.RunnableStateInstalling)
	err := c.mystManager.Update()
	if err != nil {
		c.lg.Println("upgrade", err)
	}

	c.mystManager.CheckCurrentVersionAndUpgrades(false)
}

func (c *Docker_) CheckCurrentVersionAndUpgrades(refreshVersionCache bool) {
	c.mystManager.CheckCurrentVersionAndUpgrades(refreshVersionCache)
}
