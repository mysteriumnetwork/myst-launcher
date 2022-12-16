package native

import (
	"context"
	"log"
	"path"
	"runtime"
	"sync"

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
	switch runtime.GOOS {
	case "windows":
		return "myst_windows_amd64.zip"
	case "darwin":
		return "myst_darwin_amd64.tar.gz"
	}
	return ""
}

func (c *Controller) CheckAndUpgradeNodeExe(forceUpgrade bool) bool {
	cfg := &c.a.GetModel().Config
	mdl := c.a.GetModel()

	exename := getNodeProcessName()
	fullpath := path.Join(c.runner.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)
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

	doRefresh := (cfg.NodeLatestTag != cfg.NodeExeVersion && cfg.AutoUpgrade) ||
		cfg.NodeExeVersion == "" ||
		cfg.NeedToCheckUpgrade() ||
		forceUpgrade

	log.Println("CheckAndUpgradeNodeExe doRefresh>", doRefresh)
	if doRefresh {
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
func (c *Controller) tryInstall(release updates.Release) error {
	log.Println("tryInstall >")

	model := c.a.GetModel()
	model.SetStateContainer(model_.RunnableStateInstalling)

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

	ui := c.a.GetUI()
	tryInstallFirewallRules(ui)
	return nil
}

var once sync.Once

func tryInstallFirewallRules(ui model.Gui_) {
	once.Do(func() {

		// check firewall rules
		needFirewallSetup := checkFirewallRules()

		if needFirewallSetup {
			ret := ui.YesNoModal("Installation", "Firewall rule missing, addition is required. Press Yes to approve.")
			if ret == model.IDYES {
				utils.RunasWithArgsAndWait(_const.FlagInstallFirewall)
			}
		}
	})
}
