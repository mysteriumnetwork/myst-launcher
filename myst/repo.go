package myst

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/mysteriumnetwork/myst-launcher/gui"

	"github.com/buger/jsonparser"
)

func CheckUpdates(imageDigest string) {
	url := "https://registry.hub.docker.com/v2/repositories/mysteriumnetwork/myst/tags?page_size=10"
	resp, err := http.Get(url)
	fmt.Println(">", err)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	data, _ := ioutil.ReadAll(resp.Body)

	//results
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
			if name == "latest" {
				latestDigest = digest
			}
			match, _ := regexp.MatchString(`^\d+\.\d+\.\d+.*$`, name)
			if match && latestDigest == digest {
				latestVersion = name
			}
			digestsMatch := strings.ToLower(digest) == strings.ToLower(imageDigest)
			if digestsMatch && match {
				currentVersion = name
			}
		}, "images")
	}, "results")

	//upToDate := (latestDigest == imageDigest)
	//fmt.Println("latestDigest >", latestDigest)
	//fmt.Println("latestVersion >", latestVersion)
	//fmt.Println("currentVersion >", currentVersion)
	//fmt.Println("upToDate >", upToDate)

	gui.UI.VersionCurrent = currentVersion
	gui.UI.VersionLatest = latestVersion
	gui.UI.Update()
}
