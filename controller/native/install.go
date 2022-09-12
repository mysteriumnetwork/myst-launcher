package native

import (
	"context"
	"log"
	"path"
	"runtime"

	"github.com/codingsince1985/checksum"
	wapi "github.com/iamacarpet/go-win64api"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
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

	log.Println("CheckAndUpgradeNodeExe>", cfg, forceUpgrade)

	if cfg.NodeExeVersion == "" || cfg.NeedToCheckUpgrade() || forceUpgrade {
		ctx := context.Background()
		release, _ := updates.FetchLatestRelease(ctx, org, repo)
		tagLatest := release.TagName

		mdl.ImageInfo.VersionLatest = tagLatest
		mdl.ImageInfo.VersionCurrent = cfg.NodeExeVersion
		mdl.ImageInfo.HasUpdate = tagLatest != cfg.NodeExeVersion
		defer func() {
			cfg.NodeLatestTag = tagLatest
			cfg.NodeExeVersion = tagLatest

			cfg.RefreshLastUpgradeCheck()
			cfg.Save()
		}()

		if cfg.NodeExeVersion != tagLatest {
			fullpath := path.Join(c.runner.binpath, exename)
			fullpath = utils.MakeCanonicalPath(fullpath)
			p, err := utils.IsProcessRunningExt(exename, fullpath)
			if err != nil {
				// c.lg.Println("IsRunningOrTryStart >", err)
			}
			if p != 0 {
				utils.TerminateProcess(p, 0)
			}

			if c.a.GetModel().Config.AutoUpgrade || sha256 == "" {
				c.tryInstall()

				sha256, _ := checksum.SHA256sum(fullpath)
				cfg.NodeExeDigest = sha256
			}

			return false
		}
		return true
	}

	return false
}

// returns: will exit
func (c *Controller) tryInstall() bool {
	log.Println("tryInstall >")
	ctx := context.Background()
	release, _ := updates.FetchLatestRelease(ctx, org, repo)
	log.Println("tryInstall >", release)

	asset := getAssetName()
	for _, v := range release.Assets {
		if v.Name != asset {
			continue
		}

		c.lg.Println("Downloading node: ", v.URL)
		fullPath := path.Join(utils.GetTmpDir(), asset)
		err := utils.DownloadFile(fullPath, v.URL, func(progress int) {
			if progress%10 == 0 {
				c.lg.Printf("%s - %d%%", v.Name, progress)
			}
		})
		if err != nil {
			c.lg.Println("err>", err)
		}
		if err = extractNodeBinary(fullPath, c.runner.binpath); err != nil {
			c.lg.Println(err)
		}
		break
	}

	ui := c.a.GetUI()
	tryInstallFirewallRules(ui)

	return false
}

func tryInstallFirewallRules(ui model.Gui_) {

	// check firewall rules
	needFirewallSetup := false
	rule, err := wapi.FirewallRuleGet(fwRuleNameUDP)
	if err != nil || rule.Name == "" {
		needFirewallSetup = true
	}
	rule, err = wapi.FirewallRuleGet(fwRuleNameTCP)
	if err != nil || rule.Name == "" {
		needFirewallSetup = true
	}
	if needFirewallSetup {
		ret := ui.ConfirmModal("Installation", "Firewall rule missing, addition is required. Press OK to approve.")
		if ret == model.IDOK {
			utils.RunasWithArgsAndWait(_const.FlagInstallFirewall)
		}
	}
}
