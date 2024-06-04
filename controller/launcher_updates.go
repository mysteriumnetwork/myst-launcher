//go:build !linux
// +build !linux

package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
)

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

	hasUpdate := launcherHasUpdate(&release, latest, &currentVer, model)
	if !hasUpdate {
		log.Println("Launcher is up to date")
		return false
	}
	log.Println("Launcher - current version:", currentVer)
	log.Println("Launcher - latest version:", latest)

	if model != nil {
		model.ProductVersionLatestUrl = release.Assets[0].URL
		model.ProductVersionLatest = release.Version.String()
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
