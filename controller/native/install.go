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

func (c *Native_) CheckUpgrades(refreshVersionCache bool) bool {

}

func (c *Native_) CheckAndUpgradeNodeExe(forceUpgrade bool) bool {
	mdl := c.model
	cfg := &mdl.Config

	exename := getNodeProcessName()
	fullpath := path.Join(c.runner.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)

	file, err := os.Stat(fullpath)
	if err != nil {
		cfg.NodeExeVersion = ""
		cfg.NodeExeTimestamp = time.Time{}
	} else {
		modTime := file.ModTime()
		if !modTime.Equal(cfg.NodeExeTimestamp) {
			log.Println("CheckAndUpgradeNodeExe>", fullpath)

			sha256, _ := checksum.SHA256sum(fullpath)
			log.Println("CheckAndUpgradeNodeExe>", cfg.NodeExeDigest, sha256, cfg.NodeExeDigest == sha256)

			if cfg.NodeExeDigest == sha256 {
				mdl.ImageInfo.VersionCurrent = cfg.NodeExeVersion
				mdl.Update()
			}
			if cfg.NodeExeDigest != sha256 || sha256 == "" {
				cfg.NodeExeDigest = sha256
				cfg.NodeExeVersion = ""
				cfg.Save()
			}
			cfg.NodeExeTimestamp = modTime
		}
	}

	doRefresh := (cfg.NodeLatestTag != cfg.NodeExeVersion && cfg.AutoUpgrade) ||
		cfg.NodeExeVersion == "" ||
		cfg.NeedToCheckUpgrade() ||
		forceUpgrade

	if doRefresh {
		log.Println("CheckAndUpgradeNodeExe doRefresh>")

		ctx := context.Background()
		release, err := updates.FetchLatestRelease(ctx, org, repo)
		if err != nil {
			log.Println("FetchLatestRelease>", err)
			return false
		}
		tagLatest := release.TagName

		mdl.ImageInfo.VersionLatest = tagLatest
		mdl.ImageInfo.VersionCurrent = cfg.NodeExeVersion
		mdl.ImageInfo.HasUpdate = tagLatest != cfg.NodeExeVersion
		mdl.Update()

		cfg.NodeLatestTag = tagLatest

		defer func() {
			cfg.RefreshLastUpgradeCheck()
			cfg.Save()
		}()

		// log.Println("cfg.NodeExeVersion != tagLatest >", cfg.NodeExeVersion, tagLatest, cfg.AutoUpgrade)
		doUpgrade := (cfg.NodeExeVersion != tagLatest && cfg.AutoUpgrade) || cfg.NodeExeVersion == ""
		log.Println("CheckAndUpgradeNodeExe doUpgrade>", doUpgrade)
		if doUpgrade {
			fullpath := path.Join(c.runner.binpath, exename)
			fullpath = utils.MakeCanonicalPath(fullpath)
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
			cfg.NodeExeVersion = tagLatest
			cfg.NodeExeDigest = sha256

			return true
		}

		return false
	}

	return false
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
