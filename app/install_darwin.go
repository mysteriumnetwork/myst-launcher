// +build darwin

package app

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	s.model.SwitchState(model.UIStateInstallNeeded)
	if !s.ui.WaitDialogueComplete() {
		return true
	}

	s.model.SwitchState(model.UIStateInstallInProgress)

	features, err := utils.QueryFeatures()
	if err != nil {
		log.Println("Failed to query feature:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	if len(features) > 0 {
		log.Println("Virtualization is not supported !")
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckVTx": true})

	hasDocker, err := utils.HasDocker()
	if err != nil {
		log.Println("Failed to check Docker:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckDocker": true})
	if hasDocker {
		s.model.SwitchState(model.UIStateInstallFinished)
		ok := s.ui.WaitDialogueComplete()
		if !ok {
			return true
		}
		s.model.SwitchState(model.UIStateInitial)
		return false
	}

	name := "Docker.dmg"
	url, err := utils.GetDockerDesktopLink()
	if err != nil {
		log.Println("Couldn't get Docker Desktop link")
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	log.Println("Downloading Docker desktop: ", url)
	err = utils.DownloadFile(utils.GetTmpDir()+name, url, func(progress int) {
		if progress%10 == 0 {
			log.Println(fmt.Sprintf("%s - %d%%", name, progress))
		}
	})
	if err != nil {
		log.Println("Couldn't get Docker Desktop")
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"DownloadFiles": true})

	var buf bytes.Buffer
	_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
	if err != nil {
		log.Println("Failed to run command:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	buf.Reset()

	_, err = utils.CmdRun(&buf, "/usr/bin/hdiutil", "attach", utils.GetTmpDir()+name)
	if err != nil {
		log.Println("Failed to run command:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	buf.Reset()

	// cp -R /Volumes/Docker/Docker.app /Applications
	_, err = utils.CmdRun(&buf, "/bin/cp", "-pR", "/Volumes/Docker/Docker.app", "/Applications")
	if err != nil {
		log.Println("Failed to run command:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	buf.Reset()

	//  xattr -d -r com.apple.quarantine /Applications/Docker.app
	_, err = utils.CmdRun(&buf, "/usr/bin/xattr", "-d", "-r", "com.apple.quarantine", "/Applications/Docker.app")
	if err != nil {
		log.Println("Failed to run command:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	buf.Reset()

	_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
	if err != nil {
		log.Println("Failed to run command:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	buf.Reset()

	s.model.UpdateProperties(model.UIProps{"InstallDocker": true})
	log.Println("Installation succeeded")

	s.model.SwitchState(model.UIStateInstallFinished)
	if !s.ui.WaitDialogueComplete() {
		return true
	}
	s.model.SwitchState(model.UIStateInitial)
	return false
}
