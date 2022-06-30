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

func (s *Controller) CheckAndUpgradeNodeExe(forceUpgrade bool) bool {
	cfg := &s.a.GetModel().Config
	mdl := s.a.GetModel()

	exename := getNodeProcessName()
	fullpath := path.Join(s.runner.binpath, exename)
	fullpath = utils.MakeCanonicalPath(fullpath)

	sha256, _ := checksum.SHA256sum(fullpath)
	if cfg.NodeExeDigest != sha256 || sha256 == "" {
		cfg.NodeExeDigest = sha256
		cfg.NodeExeVersion = ""
		cfg.Save()
	}

	if cfg.NodeExeVersion == "" || cfg.NeedToCheckUpgrade() || forceUpgrade {
		ctx := context.Background()
		release, _ := updates.FetchLatestRelease(ctx, org, repo)
		tagLatest := release.TagName

		mdl.ImageInfo.VersionLatest = tagLatest
		mdl.ImageInfo.VersionCurrent = cfg.NodeExeVersion
		mdl.ImageInfo.HasUpdate = tagLatest != cfg.NodeExeVersion
		defer func() {
			cfg.NodeLatestTag = tagLatest
			cfg.RefreshLastUpgradeCheck()
			cfg.Save()	
		}()

		if cfg.NodeExeVersion != tagLatest {

			fullpath := path.Join(s.runner.binpath, exename)
			fullpath = utils.MakeCanonicalPath(fullpath)
			p, err := utils.IsProcessRunningExt(exename, fullpath)
			if err != nil {
				log.Println("IsRunningOrTryStart >", err)
			}
			if p != 0 {
				utils.TerminateProcess(p, 0)
			}

			if s.a.GetModel().Config.AutoUpgrade {
				s.tryInstall()

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
func (s *Controller) tryInstall() bool {

	ctx := context.Background()
	release, _ := updates.FetchLatestRelease(ctx, org, repo)

	for _, v := range release.Assets {
		if v.Name == asset {
			log.Println("Downloading node: ", v.URL)

			fullPath := path.Join(utils.GetTmpDir(), asset)
			err := utils.DownloadFile(fullPath, v.URL, func(progress int) {
				if progress%10 == 0 {
					log.Println(fmt.Sprintf("%s - %d%%", v.Name, progress))
				}
			})
			if err != nil {
				log.Println("err>", err)
			}

			err = unzip.New(fullPath, s.runner.binpath).Extract()
			if err != nil {
				log.Println(err)
			}
			break
		}
	}

	return false
}
