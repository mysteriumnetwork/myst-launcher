package app

import (
	"context"
	"fmt"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"strings"
	"time"
)

func (s *AppState) CheckLauncherUpdates() {
	ctx := context.TODO()

	for {
		func() {
			release, err := updates.FetchLatestLauncherRelease(ctx)
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
			launcherHasUpdate := false
			if strings.Compare(release.Version.String(), pvCurrent) > 0 {
				launcherHasUpdate = true
			}

			if launcherHasUpdate != s.model.LauncherHasUpdate {
				s.model.LauncherHasUpdate = launcherHasUpdate
				s.model.ProductVersionLatestUrl = release.Assets[0].URL
				s.model.Update()
			}
		}()

		time.Sleep(time.Hour * 24)
	}
}
