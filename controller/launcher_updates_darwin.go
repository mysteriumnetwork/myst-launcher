//go:build darwin
// +build darwin

package controller

import (
	"strings"
	
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
)

const (
	gitHubOrg  = "mysteriumnetwork"
	gitHubRepo = "myst-launcher-osx"
)

func launcherHasUpdate(release *updates.Release, latest string, currentVer *string, model *model.UIModel) bool {

	*currentVer = model.ProductVersion
	launcherHasUpdate := strings.Compare(release.Version.String(), *currentVer) > 0

	return launcherHasUpdate
}

func downloadAndInstall(release updates.Release, model *model.UIModel) error {
	return nil
}
