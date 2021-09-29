package myst

import (
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+.*$`)

func CheckVersionAndUpgrades(mod *model.UIModel) {
	var data []byte
	getFile := func() bool {
		url := "https://registry.hub.docker.com/v2/repositories/mysteriumnetwork/myst/tags?page_size=30"
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return false
		}
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return false
		}

		mod.Config.RefreshLastUpgradeCheck()
		mod.Config.Save()
		return true
	}

	f := utils.GetTmpDir() + "/myst_docker_hub_cache.txt"
	data, _ = os.ReadFile(f)

	// results
	latestDigest := ""
	latestVersion := ""
	currentVersion := ""

	parseJson := func() {
		jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err_ error) {
			name, err := jsonparser.GetString(value, "name")
			if err != nil {
				return
			}

			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err_ error) {
				digest, err := jsonparser.GetString(value, "digest")
				if err != nil {
					return
				}

				if name == _const.ImageTag {
					latestDigest = digest
				}
				match := versionRegex.MatchString(name)
				// a work-around for testnet3, b/c there's only 1 version of image
				if _const.ImageTag == "testnet3" {
					match = true
				}

				if match {
					if latestDigest == digest {
						latestVersion = name
					}
					digestsMatch := strings.EqualFold(digest, mod.ImgVer.CurrentImgDigest)
					if digestsMatch {
						currentVersion = name
					}
				}

			}, "images")
		}, "results")
	}
	hasUpdate := func() bool {
		return (latestDigest != "" && mod.ImgVer.CurrentImgDigest != "") && latestDigest != mod.ImgVer.CurrentImgDigest
	}
	updateUI := func() {
		mod.ImgVer.VersionCurrent = currentVersion
		mod.ImgVer.VersionLatest = latestVersion

		mod.ImgVer.HasUpdate = hasUpdate()
		mod.Update()
	}

	if len(data) != 0 {
		parseJson()
		updateUI()
	}

	if len(data) == 0 || mod.Config.NeedToCheckUpgrade() {
		ok := getFile()
		if ok {
			os.WriteFile(f, data, 0777)
		} else {
			return
		}

		parseJson()
		updateUI()
	}

}
