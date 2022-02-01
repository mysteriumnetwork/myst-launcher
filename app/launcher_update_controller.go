package app

import (
	"context"
	"fmt"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"strings"
	"time"
)

func (s *AppState) CheckLauncherUpdates(gitHubOrg, gitHubRepo string) {
	ctx := context.TODO()

	for {
		func() {
			release, err := updates.FetchLatestLauncherRelease(ctx, gitHubOrg, gitHubRepo)
			if err != nil {
				fmt.Println(err)
				return
			}

			// version has no postfix
			if release.Version.String() != release.Version.FinalizeVersion() {
				return
			}
			s.model.ProductVersionLatest = release.Version.String()

			pvCurrent := s.model.ProductVersion
			launcherHasUpdate := strings.Compare(release.Version.String(), pvCurrent) > 0

			if launcherHasUpdate != s.model.LauncherHasUpdate {
				isNew := s.model.ProductVersionLatestUrl != release.Assets[0].URL

				s.model.LauncherHasUpdate = launcherHasUpdate
				s.model.ProductVersionLatestUrl = release.Assets[0].URL
				s.model.Update()

				if isNew {
					s.model.Publish("launcher-update")
				}
			}
		}()

		time.Sleep(time.Hour * 24)
	}
}
