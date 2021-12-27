//go:build windows
// +build windows

package updates

import (
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

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
