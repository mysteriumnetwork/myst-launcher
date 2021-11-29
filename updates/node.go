/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package updates

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Tags struct {
	Detail   string      `json:"detail"`
	Count    int64       `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Results  []TagResult `json:"results"`
}

type TagResult struct {
	Creator             int         `json:"creator"`
	ID                  int         `json:"id"`
	Images              []ImageInfo `json:"images"`
	LastUpdated         string      `json:"last_updated"`
	LastUpdater         int         `json:"last_updater"`
	LastUpdaterUsername string      `json:"last_updater_username"`
	Name                string      `json:"name"`
	Repository          int         `json:"repository"`
	FullSize            int         `json:"full_size"`
	V2                  bool        `json:"v2"`
	TagStatus           string      `json:"tag_status"`
	TagLastPulled       string      `json:"tag_last_pulled"`
	TagLastPushed       string      `json:"tag_last_pushed"`
}

type ImageInfo struct {
	Architecture string `json:"architecture"`
	Features     string `json:"features"`
	Variant      string `json:"variant"`
	Digest       string `json:"digest"`
	OS           string `json:"os"`
	OSFeatures   string `json:"os_features"`
	OSVersion    string `json:"os_version"`
	Size         int64  `json:"size"`
	Status       string `json:"status"`
	LastPulled   string `json:"last_pulled"`
	LastPushed   string `json:"last_pushed"`
}

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
	log.Println("CheckCurrentVersionAndUpgrades>")

	var data []byte
	getFile := func() bool {
		url := "https://registry.hub.docker.com/v2/repositories/mysteriumnetwork/myst/tags?page_size=10"
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
	imageTag := mod.Config.GetLatestImageTag()

	parseJson := func() {
		found := false

		var queryResults Tags
		if err := json.Unmarshal(data, &queryResults); err != nil {
			return
		}

		for _, res := range queryResults.Results {
			matchVersionRegex := versionRegex.MatchString(res.Name)
			for _, im := range res.Images {

				if res.Name == imageTag && im.Architecture == "amd64" {
					latestDigest = im.Digest
				}
				// a work-around for custom tags like testnet3, which have only 1 version of image
				if imageTag != "latest" {
					matchVersionRegex = true
				}

				if matchVersionRegex {
					if latestDigest == im.Digest {
						latestVersion = res.Name
					}

					// multi-arch images have 2 digests: one for image itself, second - for manifest
					if mod.ImageInfo.HasDigest(im.Digest) {
						currentVersion = res.Name
						found = true
					}
				}
			}
		}

		if !found {
			currentVersion = ""
		}
	}

	hasUpdate := func() bool {
		if latestDigest != "" {
			// has a digest of the latest version
			return !mod.ImageInfo.HasDigest(latestDigest)
		}
		return false
	}

	updateUI := func() {
		mod.ImageInfo.DigestLatest = latestDigest
		mod.ImageInfo.VersionCurrent = currentVersion
		mod.ImageInfo.VersionLatest = latestVersion
		mod.ImageInfo.HasUpdate = hasUpdate()
		mod.Update()
	}

	if len(data) != 0 {
		parseJson()
		updateUI()
	}
	if checkVersionRequirement(currentVersion, minVersion) {
		mod.CurrentImgHasReportVersionOption = true
	}

	// Reload image list if cache has no info about current version
	if !fastPath && len(data) == 0 || mod.Config.NeedToCheckUpgrade() || currentVersion == "" {
		ok := getFile()
		if ok {
			os.WriteFile(f, data, 0777)

			parseJson()
			updateUI()

			if checkVersionRequirement(currentVersion, minVersion) {
				mod.CurrentImgHasReportVersionOption = true
			}
		}
	}

}
