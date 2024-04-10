package docker

import "github.com/mysteriumnetwork/myst-launcher/controller/docker/myst"

func UninstallMystContainer() error {
	mystManager, err := myst.NewManager(nil)
	if err != nil {
		return err
	}
	err = mystManager.Stop()
	if err != nil {
		return err
	}
	err = mystManager.Remove()
	return err
}
