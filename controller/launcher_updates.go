package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
)

const (
	gitHubOrg  = "mysteriumnetwork"
	gitHubRepo = "myst-launcher"
)

func CheckLauncherUpdates(model *model.UIModel) {
	ctx := context.TODO()

	for {
		func() {
			release, err := updates.FetchLatestRelease(ctx, gitHubOrg, gitHubRepo)
			if err != nil {
				fmt.Println(err)
				return
			}

			// version has no postfix
			if release.Version.String() != release.Version.FinalizeVersion() {
				return
			}
			model.ProductVersionLatest = release.Version.String()

			pvCurrent := model.ProductVersion
			launcherHasUpdate := strings.Compare(release.Version.String(), pvCurrent) > 0

			if launcherHasUpdate != model.LauncherHasUpdate {
				isNew := model.ProductVersionLatestUrl != release.Assets[0].URL

				model.LauncherHasUpdate = launcherHasUpdate
				model.ProductVersionLatestUrl = release.Assets[0].URL
				model.Update()

				if isNew {
					model.Publish("launcher-update")
				}
			}
		}()

		time.Sleep(time.Hour * 24)
	}
}
