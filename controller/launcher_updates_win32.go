//go:build windows
// +build windows

package controller

import (
	"fmt"
	"path"

	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	gitHubOrg  = "mysteriumnetwork"
	gitHubRepo = "myst-launcher"
)

func launcherHasUpdate(release *updates.Release, latest string, currentVer *string, model *model.UIModel) bool {

	hasUpdate, _ := utils.LauncherMSIHasUpdateOrPkgNI(latest, currentVer)
	return hasUpdate
}

func downloadAndInstall(release updates.Release, model *model.UIModel) error {
	url, name := "", ""
	for _, v := range release.Assets {
		if v.Name == "myst-launcher-x64.msi" {
			url, name = v.URL, v.Name
			break
		}
	}
	if url == "" {
		return nil
	}
	fmt.Println("Downloading update:", url)

	msiPath := path.Join(utils.GetTmpDir(), name)
	msiPath = utils.MakeCanonicalPath(msiPath)
	fmt.Println(msiPath)

	err := utils.DownloadFile(msiPath, url, func(progress int) {
		if model != nil {
			model.Publish("launcher-update-download", progress)
		}
		if progress%10 == 0 {
			fmt.Printf("%s - %d%%\n", name, progress)
		}
	})
	if err != nil {
		fmt.Println("Download error:", err)
		model.Publish("launcher-update-download", -1)
		return err
	}
    err = utils.RunMsi(msiPath)
	if err != nil {
		fmt.Println("RunMsi err>", err)
		model.Publish("launcher-update-download", -1)
		return err
	}
	fmt.Println("Update successfully completed!")
	return nil
}

// return: bool exit
func UpdateLauncherFromNewBinary(ui *gui_win32.Gui, p *ipc_.Handler) bool {
	if utils.LauncherUpgradeAvailable() {
		update := ui.YesNoModal("Mysterium launcher upgrade", "You are running a newer version of launcher.\r\nUpgrade launcher installation ?")
		if model.IDYES == update {
			if !p.OwnsPipe() {
				p.SendStopApp()
				p.OpenPipe()
			}
			utils.UpdateExe()
			return false
		}
	}

	if !p.OwnsPipe() {
		p.SendPopupApp()
		return true
	}
	return false
}

func PopupFirstInstance(ui *gui_win32.Gui, p *ipc_.Handler) bool {
	if !p.OwnsPipe() {
		p.SendPopupApp()
		return true
	}
	return false
}
