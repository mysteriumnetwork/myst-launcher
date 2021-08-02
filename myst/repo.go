package myst

import (
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+.*$`)

const checkPeriod = 12 * time.Hour

func CheckVersionAndUpgrades(imageDigest string, c *model.Config) bool {
	var data []byte
	f := os.Getenv("TMP") + "/myst_docker_hub_cache.txt"
	i, err := os.Stat(f)
	if err == nil {
		if i.ModTime().Add(checkPeriod).After(time.Now()) {
			data, err = os.ReadFile(f)
		}
	}
	if len(data) == 0 {
		url := "https://registry.hub.docker.com/v2/repositories/mysteriumnetwork/myst/tags?page_size=30"
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		if resp.StatusCode != 200 {
			return false
		}
		data, _ := ioutil.ReadAll(resp.Body)
		os.WriteFile(f, data, 0777)

		c.RefreshLastUpgradeCheck()
	}

	// results
	latestDigest := ""
	latestVersion := ""
	currentVersion := ""

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		name, err := jsonparser.GetString(value, "name")
		if err != nil {
			return
		}

		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			digest, err := jsonparser.GetString(value, "digest")
			if err != nil {
				return
			}

			if name == imageTag {
				latestDigest = digest
			}
			match := versionRegex.MatchString(name)
			// a work-around for testnet3, b/c there's only 1 version of image
			if imageTag == "testnet3" {
				match = true
			}
			if match && latestDigest == digest {
				latestVersion = name
			}
			digestsMatch := strings.ToLower(digest) == strings.ToLower(imageDigest)
			if digestsMatch && match {
				currentVersion = name
			}
		}, "images")
	}, "results")

	gui.UI.VersionUpToDate = latestDigest == imageDigest
	gui.UI.VersionCurrent = currentVersion
	gui.UI.VersionLatest = latestVersion
	gui.UI.Update()

	return true
}
