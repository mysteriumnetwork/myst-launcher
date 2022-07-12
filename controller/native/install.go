package native

import (
	"context"
	"fmt"
	"log"
	"path"

	"github.com/artdarek/go-unzip"
	"github.com/codingsince1985/checksum"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	asset = "myst_windows_amd64.zip"
	org   = "mysteriumnetwork"
	repo  = "node"
)

func (c *Controller) CheckAndUpgradeNodeExe(forceUpgrade bool) bool {
	cfg := &c.a.GetModel().Config
	mdl := c.a.GetModel()

	exename := getNodeProcessName()
	fullpath := path.Join(c.runner.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)
	log.Println("CheckAndUpgradeNodeExe>", fullpath)

	sha256, _ := checksum.SHA256sum(fullpath)
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
		log.Println("CheckAndUpgradeNodeExe>", release)


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

			if c.a.GetModel().Config.AutoUpgrade {
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

	for _, v := range release.Assets {
		if v.Name == asset {
			c.lg.Println("Downloading node: ", v.URL)

			fullPath := path.Join(utils.GetTmpDir(), asset)
			err := utils.DownloadFile(fullPath, v.URL, func(progress int) {
				if progress%10 == 0 {
					c.lg.Println(fmt.Sprintf("%s - %d%%", v.Name, progress))
				}
			})
			if err != nil {
				c.lg.Println("err>", err)
			}

			err = unzip.New(fullPath, c.runner.binpath).Extract()
			if err != nil {
				c.lg.Println(err)
			}
			break
		}
	}

	return false
}
