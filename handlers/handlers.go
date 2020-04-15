package handlers

import (
	"context"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"fmt"

	Types "../types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var ClientImageList string

func UpdateClientImageList(registryUrl *string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		// With address
		// fmt.Println(image.RepoTags)
		matchList := []string{}
		tags := image.RepoTags
		// re := regexp.MustCompile(`(([0–9]|[1–9][0–9]|1[0–9]{2}|2[0–4][0–9]|25[0–5])\.){3}([0–9]|[1–9][0–9]|1[0–9]{2}|2[0–4][0–9]|25[0–5])\:5000/([a-zA-Z0-9_\-])/g`)
		// submatch := re.FindSubmatch([]byte(tag[0]))
		for _, tag := range tags {
			if !strings.Contains(tag, *registryUrl) {
				matchList = append(matchList, tag)
			}
		}
		fmt.Println(matchList)
		// for _, match := range submatch {
		// 	if
		// 	matchList = append(matchList, string(match))
		// }
		// image.RepoTags
	}
	return
}

// Status of the client
// func Status(req *Types.RequestStruct) map[string]interface{} {
// 	retu
// }

func HandleRequest(req *Types.RequestStruct, regsitryUrl *string) map[string]interface{} {
	result := make(map[string]interface{})
	resultList := []string{}

	switch req.Request {
	case "Status":
		UpdateClientImageList(regsitryUrl)
		result["List"] = ClientImageList
		break
	case "Image.Pull":
		var listArg []string
		log.Printf("The image list to process ==> %s")
		log.Println(req.Arguments["List"])
		s := reflect.ValueOf(req.Arguments["List"])
		if s.Kind() != reflect.Slice {
			panic("INterface mismatches!")
		}

		for i := 0; i < s.Len(); i++ {
			listArg = append(listArg, fmt.Sprintf("%s", s.Index(i)))
		}
		log.Printf("The listArg = %v", listArg)

		for i, image := range listArg {
			log.Printf("Image.Pull :: Iteration # %i", i)
			imageTagSource := fmt.Sprintf("%v/%v", *regsitryUrl, image)
			log.Printf("HandleRequest :: imageTagSource  %s", imageTagSource)
			err := ImagePull(imageTagSource, image)
			if err != nil {
				result["error"] = err
				break
			}
			resultList = append(resultList, image)

		}
		result["list"] = resultList
	}
	log.Printf("The result => %v", resultList)
	return result
}

func ImagePull(imageTagSource string, imageTagTarget string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error on creating a client with opts: %s", err)
		panic(err)
	}

	out, err := cli.ImagePull(ctx, imageTagSource, types.ImagePullOptions{})
	if err != nil {
		log.Printf("Error on pulling an image: %s", err)
		panic(err)
	}

	// Retag the image to what we need
	err = cli.ImageTag(ctx, imageTagSource, imageTagTarget)
	if err != nil {
		log.Printf("Error retagging image from %s to %s", imageTagSource, imageTagTarget)
	}

	defer out.Close()
	io.Copy(os.Stdout, out)

	return nil
}
