/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package updates

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

var versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+).*$`)
var minVersion = []int{0, 66, 3}

func checkVersionRequirement(v string, minVersion []int) bool {
	log.Println("checkVersionRequirement>", v, minVersion)
	match := versionRegex.MatchString(v)
	if !match {
		return false
	}
	versionParts := versionRegex.FindAllStringSubmatch(v, -1)

	verMatches := false
	for k, v := range versionParts[0][1:] {
		i, _ := strconv.Atoi(v)

		// sufficient condition
		if i > minVersion[k] {
			verMatches = true
			break
		}

		// last part
		if k == len(minVersion)-1 {
			verMatches = i >= minVersion[k]
		}
	}
	return verMatches
}

func CheckVersionAndUpgrades(mod *model.UIModel, fastPath bool) {
	// log.Println("CheckVersionAndUpgrades 1>")

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

	imageTag := mod.Config.GetImageTag()
	parseJson := func() {
		jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err_ error) {
			name, err := jsonparser.GetString(value, "name")
			if err != nil {
				return
			}
			match := versionRegex.MatchString(name)

			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err_ error) {
				digest, err := jsonparser.GetString(value, "digest")
				if err != nil {
					return
				}

				if name == imageTag {
					latestDigest = digest
				}
				// a work-around for custom tags like testnet3, which have only 1 version of image
				if imageTag != "latest" {
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
	if checkVersionRequirement(currentVersion, minVersion) {
		mod.CurrentImgHasOptionReportVersion = true
	}

	// Reload image list if cache has no info about current version
	if !fastPath && len(data) == 0 || mod.Config.NeedToCheckUpgrade() || currentVersion == "" {
		ok := getFile()
		if ok {
			os.WriteFile(f, data, 0777)

			parseJson()
			updateUI()

			if checkVersionRequirement(currentVersion, minVersion) {
				mod.CurrentImgHasOptionReportVersion = true
			}
		}
	}

}
