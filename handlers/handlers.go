package handlers

import (
	"context"
	"io"
	"log"
	"os"
	"reflect"

	// "regexp"

	"fmt"

	Types "../types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var ClientContainerList []types.Container

// var ClientImageList []string
var ClientImageList []types.ImageSummary

// // Update list of image on the client side
// var UpdateClientImageList = func() {
// 	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
// 	if err != nil {
// 		panic(err)
// 	}

// 	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	var imageFoundList []string

// 	for _, image := range images {
// 		// With address
// 		// fmt.Println(image.RepoTags)
// 		matchList := []string{}
// 		tags := image.RepoTags
// 		re := regexp.MustCompile(`^(?P<address>(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):[0-9]+\/())?(?P<name>.*)`)

// 		for _, tag := range tags {
// 			submatch := re.FindStringSubmatch(string(tag))
// 			addressGroup := ""
// 			nameGroup := ""
// 			for i, name := range re.SubexpNames() {
// 				if name == "address" {
// 					addressGroup = submatch[i]
// 				}
// 				if name == "name" {
// 					nameGroup = submatch[i]
// 				}
// 			}
// 			if addressGroup != "" {
// 				continue
// 			}
// 			if nameGroup != "" {
// 				matchList = append(matchList, nameGroup)

// 			}

// 		}
// 		if len(matchList) != 0 {
// 			imageFoundList = append(imageFoundList, matchList[0])
// 		}

// 		// for _, tag := range tags {
// 		// 	if !strings.Contains(tag, *registryUrl) {
// 		// 		matchList = append(matchList, tag)
// 		// 	}
// 		// }
// 		// for _, match := range submatch {
// 		// 	if
// 		// 	matchList = append(matchList, string(match))
// 		// }
// 		// image.RepoTags
// 	}

// 	ClientImageList = imageFoundList
// 	fmt.Println("The list of images")
// 	fmt.Println(ClientImageList)
// 	return
// }

// Update list of image on the client side
var UpdateClientImageList = func() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}
	ClientImageList = images
}

// UpdateClientContainerList : update the client container list for all the servers
func UpdateClientContainerList() {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containerList, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	ClientContainerList = containerList
}

// Status of the client
// func Status(req *Types.RequestStruct) map[string]interface{} {
// 	retu
// }

func HandleRequest(req *Types.RequestStruct, regsitryUrl *string) map[string]interface{} {
	result := make(map[string]interface{})
	resultList := []string{}
	var err error

	switch req.Request {
	case "Status":
		result, err = HandleStatus()
		break

	case "Image.Pull":
		var listArg []string
		log.Printf("The image list to process ==>")
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
			log.Printf("HandleRequest :: {imageTagSource: %s, image: `%s`}", imageTagSource, image)
			err = ImagePullSingle(imageTagSource, image)
			if err != nil {
				result["error"] = err
				break
			}
			resultList = append(resultList, image)

		}
		result["List"] = resultList
		break

	case "Image.Run":
		imageName := req.Arguments["Image.Name"]

		err = RunContainer(imageName.(string))
		if err != nil {
			result["error"] = err
			break
		}
		result["List"] = []string{"Success"}
		break

	case "Container.Stop":
		containerID := req.Arguments["Container.ID"]

		err = StopContainer(containerID.(string))
		if err != nil {
			result["error"] = err
			break
		}
		result["List"] = []string{"Success"}
		break
	}
	log.Printf("The result => %v", resultList)
	return result
}

// Handle the `Status` Request
func HandleStatus() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	UpdateClientImageList()
	result["Image.List"] = ClientImageList

	UpdateClientContainerList()
	result["Container.List"] = ClientContainerList
	return result, nil
}

// Pull a single image and retag it to local image tag
func ImagePullSingle(imageTagSource string, imageTagTarget string) error {
	log.Printf("Called ImagePull with (%s, %s)", imageTagSource, imageTagTarget)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error on creating a client with opts: %s", err)
		panic(err)
	}

	// Pull the image
	out, err := cli.ImagePull(ctx, imageTagSource, types.ImagePullOptions{})
	if err != nil {
		log.Printf("Error on pulling an image: %s", err)
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Retag the image to what we need
	err = cli.ImageTag(ctx, imageTagSource, imageTagTarget)
	if err != nil {
		log.Printf("Error retagging image from %s to %s", imageTagSource, imageTagTarget)
		log.Printf("Error: %s", err)
	}

	// Remove the external image
	_, err = cli.ImageRemove(ctx, imageTagSource, types.ImageRemoveOptions{})
	if err != nil {
		log.Printf("Error removing image %s", imageTagSource)
	}

	return nil
}

// RunContainer starts a new container by it's image name
func RunContainer(imageName string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containerName := ""
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	fmt.Printf("Started a container with ID %s", resp.ID)

	return nil

}

// StopContainer starts a new container by it's image name
func StopContainer(containerTag string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStop(ctx, containerTag, nil)
	if err != nil {
		panic(err)
		return err
	}
	fmt.Printf("Stopped a container with ID %s", containerTag)

	return nil

}
