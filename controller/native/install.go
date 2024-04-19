package native

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/codingsince1985/checksum"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
	model_ "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	org  = "mysteriumnetwork"
	repo = "node"
)

func getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	f := "tar.gz"
	if os == "windows" {
		f = "zip"
	}
	return fmt.Sprintf("myst_%s_%s.%s", os, arch, f)
}

// return: has update
func (c *Native_) CheckAndUpgradeNodeExe_(refreshVersionCache, doUpgrade bool) bool {
	log.Println("CheckAndUpgradeNodeExe>")

	mdl := c.model
	cfg := &mdl.Config

	// set ui part
	setUi := func() {
		mdl.ImageInfo.VersionLatest = cfg.NodeExeLatestTag
		mdl.ImageInfo.VersionCurrent = cfg.NodeExeVersion
		mdl.ImageInfo.HasUpdate = cfg.NodeExeLatestTag != cfg.NodeExeVersion
		mdl.Update()
	}

	exename := getNodeProcessName()
	fullpath := utils.MakeCanonicalPath(path.Join(c.runner.binpath, exename))

    fileAbsent := false
	file, err := os.Stat(fullpath)
	if err != nil {
		cfg.NodeExeVersion = ""
		cfg.NodeExeTimestamp = time.Time{}
        fileAbsent = true
	} else {
		modTime := file.ModTime()
		if !modTime.Equal(cfg.NodeExeTimestamp) {

			sha256, _ := checksum.SHA256sum(fullpath)

			// if cfg.NodeExeDigest == sha256 {
			// }
			if cfg.NodeExeDigest != sha256 || sha256 == "" {
				cfg.NodeExeDigest = sha256
				cfg.NodeExeVersion = ""
			}

			cfg.NodeExeTimestamp = modTime
			cfg.Save()
		}
	}
	setUi()

	hasUpdate := func() bool {
		return (cfg.NodeExeVersion != cfg.NodeExeLatestTag) || cfg.NodeExeVersion == ""
	}

	doRefresh := cfg.NodeExeVersion == "" || cfg.NodeExeLatestTag == "" ||
		cfg.TimeToCheckUpgrade() ||
		refreshVersionCache ||
		fileAbsent ||
		doUpgrade

	if !doRefresh {
		return hasUpdate()
	}

	log.Println("CheckAndUpgradeNodeExe doRefresh>")
	release, err := updates.FetchLatestRelease(context.Background(), org, repo)
	if err != nil {
		log.Println("FetchLatestRelease>", err)
		return false
	}
	cfg.NodeExeLatestTag = release.TagName
	cfg.RefreshLastUpgradeCheck()
	defer func() {
		cfg.Save()
	}()

	setUi()

	doUpgrade_ := hasUpdate() && doUpgrade
	log.Println("CheckAndUpgradeNodeExe doUpgrade>", hasUpdate(), doUpgrade_)

	if doUpgrade_ || fileAbsent {
		p, _ := utils.IsProcessRunningExt(exename, fullpath)
		if p != 0 {
			utils.TerminateProcess(p, 0)
		}

		err := c.tryInstall(release)
		if err != nil {
			c.lg.Println("tryInstall >", err)
			return false
		}

		sha256, _ := checksum.SHA256sum(fullpath)
		cfg.NodeExeVersion = cfg.NodeExeLatestTag
		cfg.NodeExeDigest = sha256

		return true
	}
	return hasUpdate()
}

// returns: will exit
func (c *Native_) tryInstall(release updates.Release) error {
	log.Println("tryInstall >")

	c.model.SetStateContainer(model_.RunnableStateInstalling)

	asset := getAssetName()
	for _, v := range release.Assets {
		if v.Name != asset {
			continue
		}

		c.lg.Println("Downloading node: ", v.URL)
		fullPath := path.Join(utils.GetTmpDir(), asset)
		err := utils.DownloadFile(fullPath, v.URL, func(progress int) {
			if progress%10 == 0 {
				c.lg.Printf("%s - %d%%\n", v.Name, progress)
			}
		})
		if err != nil {
			c.lg.Println("download>", err)
			return err
		}
		if err = extractNodeBinary(fullPath, c.runner.binpath); err != nil {
			c.lg.Println("extractNodeBinary", err)
			return err
		}
		break
	}

	tryInstallFirewallRules(c.ui)
	return nil
}

var once sync.Once

func tryInstallFirewallRules(ui model.Gui_) {
	once.Do(func() {
		// check firewall rules
		needFirewallSetup := checkFirewallRules()

		if needFirewallSetup {
			ret := model.IDYES
			if ui != nil {
				ret = ui.YesNoModal("Installation", "Firewall rule missing, addition is required. Press Yes to approve.")
			}
			if ret == model.IDYES {
				if utils.IsAdmin() {
					CheckAndInstallFirewallRules()
				} else {
					utils.RunasWithArgsAndWait("-" + _const.FlagInstallFirewall)
				}
			}
		}
	})
}
