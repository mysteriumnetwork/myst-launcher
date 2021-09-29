package myst

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/buger/jsonparser"
)

var versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+).*$`)
var minVersion = []int{0, 64, 5}

func checkVersionRequirement(v string, minVersion []int) bool {
	match := versionRegex.MatchString(v)
	if !match {
		return false
	}
	versionParts := versionRegex.FindAllStringSubmatch(v, -1)
	fmt.Println("versionParts", versionParts[0][1:])

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

const checkPeriod = 12 * time.Hour

func CheckVersionAndUpgrades(mod *model.UIModel) {
	var data []byte
	f := utils.GetTmpDir() + "/myst_docker_hub_cache.txt"
	i, err := os.Stat(f)
	if err == nil {
		if i.ModTime().Add(checkPeriod).After(time.Now()) {
			data, err = os.ReadFile(f)
		}
	}

	getFile := func() bool {
		url := "https://registry.hub.docker.com/v2/repositories/mysteriumnetwork/myst/tags?page_size=30"
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		if resp.StatusCode != 200 {
			return false
		}
		data, _ = ioutil.ReadAll(resp.Body)
		os.WriteFile(f, data, 0777)

		mod.Config.RefreshLastUpgradeCheck()
		mod.Config.Save()
		return true
	}

	if len(data) == 0 {
		ok := getFile()
		if !ok {
			return
		}
	}

	// results
	latestDigest := ""
	latestVersion := ""
	currentVersion := ""

	parseJson := func() {
		jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			name, err := jsonparser.GetString(value, "name")
			if err != nil {
				return
			}
			match := versionRegex.MatchString(name)

			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				digest, err := jsonparser.GetString(value, "digest")
				if err != nil {
					return
				}

				if name == _const.ImageTag {
					latestDigest = digest
				}
				// a work-around for testnet3, b/c there's only 1 version of image
				if _const.ImageTag == "testnet3" {
					match = true
				}
				if match && latestDigest == digest {
					latestVersion = name
				}
				digestsMatch := strings.ToLower(digest) == strings.ToLower(mod.ImgVer.CurrentImgDigest)
				if digestsMatch && match {
					currentVersion = name
				}
			}, "images")
		}, "results")
	}
	parseJson()

	mod.ImgVer.HasUpdate = false
	if (latestDigest != "" && mod.ImgVer.CurrentImgDigest != "") && latestDigest != mod.ImgVer.CurrentImgDigest {
		// re-try on fresh data
		ok := getFile()
		if !ok {
			return
		}
		parseJson()

		if (latestDigest != "" && mod.ImgVer.CurrentImgDigest != "") && latestDigest != mod.ImgVer.CurrentImgDigest {
			mod.ImgVer.HasUpdate = true
		}
	}
	mod.ImgVer.VersionCurrent = currentVersion
	mod.ImgVer.VersionLatest = latestVersion
	mod.Update()

	if checkVersionRequirement(currentVersion, minVersion) {
		mod.GetConfig().HasOptionReportVersion = true
	}
}
