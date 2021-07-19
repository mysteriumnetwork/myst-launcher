package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/mysteriumnetwork/myst-launcher/gui"

	"github.com/buger/jsonparser"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Tag struct {
}

var cli *client.Client

func init() {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
}

func getCurrent() string {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		//panic(err)
		return ""
	}
	imageID := ""
	for _, container := range containers {
		if container.Image == "mysteriumnetwork/myst:latest" {
			fmt.Printf("%+v\n", container)
			//fmt.Printf("%+v\n", container.ImageID)
			imageID = container.ImageID
		}
	}

	imageDigest := ""
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		//panic(err)
		return ""
	}
	for _, image := range images {
		fmt.Printf("%+v\n", image.RepoDigests)

		if imageID == image.ID {
			fmt.Printf("%+v\n", image.RepoDigests)

			for _, rd := range image.RepoDigests {
				fmt.Printf("rd> %+v\n", rd)
				digestArr := strings.Split(rd, "@")
				imageDigest = digestArr[1]
			}
		}
	}
	fmt.Printf("rd> %+v\n", imageDigest)
	return imageDigest
}

func checkUpdates() {
	imageDigest := getCurrent()
	imageDigest = "sha256:ff530e6dbc2538aa92887833db24ba8d40f5d630b6ba32a34a258873aad9fac9"

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
		//fmt.Println("r>", string(value), err)
		name, err := jsonparser.GetString(value, "name")
		fmt.Println("r>", name, err)

		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			digest, err := jsonparser.GetString(value, "digest")
			//fmt.Println("i>", string(value), err)

			if name == "latest" {
				latestDigest = digest
			}
			//fmt.Println("i >>>", digest, name)

			match, _ := regexp.MatchString(`^\d+\.\d+\.\d+.*$`, name)
			if match && latestDigest == digest {
				latestVersion = name
			}

			digestsMatch := strings.ToLower(digest) == strings.ToLower(imageDigest)
			if digestsMatch && match {
				fmt.Println("i>>", digest, name, match)
				currentVersion = name
			}

		}, "images")
	}, "results")

	upToDate := (latestDigest == imageDigest)
	fmt.Println("latestDigest >", latestDigest)
	fmt.Println("latestVersion >", latestVersion)
	fmt.Println("currentVersion >", currentVersion)
	fmt.Println("upToDate >", upToDate)

	gui.UI.VersionCurrent = currentVersion
	gui.UI.VersionLatest = latestVersion
	gui.UI.Update()
}

func upgrade() {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		//panic(err)
		return
	}

	containerID := ""
	for _, container := range containers {
		if container.Image == "mysteriumnetwork/myst:latest" {
			containerID = container.ID
			//fmt.Printf("%+v\n", container.ImageID)
		}
	}
	fmt.Printf("%+v\n", containerID)
	//err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	//fmt.Printf("%+v\n", err)

	//cli.ContainerExecCreate()
}
