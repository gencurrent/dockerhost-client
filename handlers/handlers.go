package handlers

import (
	"context"
	"io"
	"log"
	"os"

	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func HandleRequest(command string, arguments map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	switch command {
	case "image.pull":
		address := fmt.Sprintf("%v", arguments["name"])
		err := ImagePull(address)
		if err != nil {
			result["error"] = err
			break
		}
		result["list"] = arguments["imageList"]
	}
	return result
}

func ImagePull(imageAddress string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error on creating a client with opts: %s", err)
		panic(err)
	}

	out, err := cli.ImagePull(ctx, imageAddress, types.ImagePullOptions{})
	if err != nil {
		log.Printf("Error on pulling an image: %s", err)
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	return nil
}
