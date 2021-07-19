package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Tag struct {
}

func CheckUpdates() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	imageID := ""
	for _, container := range containers {
		if container.Image == "mysteriumnetwork/myst:latest" {
			fmt.Printf("%+v\n", container)
			//fmt.Printf("%+v\n", container.ImageID)
			imageID = container.ImageID
		}
	}
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}
	for _, image := range images {
		_ = image
		fmt.Printf("%+v\n", image)
	}

	return
	//////
	url := "https://registry.hub.docker.com/v2/repositories/mysteriumnetwork/myst/tags?page_size=10"
	resp, err := http.Get(url)
	fmt.Println(">", err)
	data, _ := ioutil.ReadAll(resp.Body)
	//val, _, _, err := jsonparser.Get(data, "results", "[0]")
	//fmt.Println(">", string(val), err)

	//results
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		//fmt.Println("r>", string(value), err)
		name, err := jsonparser.GetString(value, "name")
		//fmt.Println("r>", name, err)

		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			//fmt.Println("i>", string(value), err)

			digest, err := jsonparser.GetString(value, "digest")
			fmt.Println("i>>", digest, strings.ToLower(digest) == strings.ToLower(imageID), name)

		}, "images")
	}, "results")

	//hub, err := registry.New(url, username, password)
	//if err != nil {
	//	//log.Fatal(err)
	//	fmt.Println(">", err)
	//	return
	//}
	//
	//tags, err := hub.Tags("mysteriumnetwork/myst")
	//if err != nil {
	//	fmt.Println(">", err)
	//	return
	//}
	//fmt.Println(">", tags)
	//
	//manifest, err := hub.Manifest("heroku/cedar", "14")
	//if err != nil {
	//	fmt.Println(">", err)
	//	return
	//}
	//fmt.Println(">", manifest.Name)
}
