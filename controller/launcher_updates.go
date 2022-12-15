package controller

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
	"github.com/winlabs/gowin32"
)

const (
	gitHubOrg  = "mysteriumnetwork"
	gitHubRepo = "myst-launcher"
)

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

	gowin32.SetInstallerInternalUI(gowin32.InstallUILevelFull)
	if err := gowin32.InstallProduct(msiPath, `ACTION=INSTALL`); err != nil {
		fmt.Println("InstallProduct err>", err)
		model.Publish("launcher-update-download", -1)
		return err
	}
	fmt.Println("Update successfully completed!")
	return nil
}

// bool - exit
func checkLauncherUpdates(model *model.UIModel) bool {
	ctx := context.TODO()
	release, err := updates.FetchLatestRelease(ctx, gitHubOrg, gitHubRepo)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if model != nil {
		model.ProductVersionLatest = release.Version.String()
	}

	// version has postfix
	if release.Version.String() != release.Version.FinalizeVersion() {
		// return false
	}

	latest := release.Version.String()
	currentVer := ""
	hasUpdate, _ := utils.LauncherMSIHasUpdateOrPkgNI(latest, &currentVer)
	if !hasUpdate {
		fmt.Println("Mysterium Launcher is up to date")
		return false
	}
	fmt.Println("There's an update for Mysterium Launcher")
	fmt.Println("Current version:", currentVer)
	fmt.Println("Latest version:", latest)

	if model != nil {
		model.LauncherHasUpdate = true
		model.Update()
		model.Publish("launcher-update")

		dlgOK := make(chan int)
		model.UIBus.SubscribeOnce("launcher-update-ok", func(action int) {
			dlgOK <- action
		})

		if <-dlgOK == 2 {
			err := downloadAndInstall(release, model)
			return err == nil
		}
	} else {
		err := downloadAndInstall(release, model)
		return err == nil
	}

	return false
}

func CheckLauncherUpdates(model *model.UIModel) {
	for {
		if checkLauncherUpdates(model) {
			return
		}

		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Hour*12)
		model.UIBus.SubscribeOnce("launcher-trigger-update", func() {
			cancel()
		})
		<-ctxTimeout.Done()
	}
}

func CheckLauncherUpdatesCli() {
	checkLauncherUpdates(nil)
}
