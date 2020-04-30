package main

import (
	"log"
	"testing"
	"encoding/json"

	Handlers "./handlers"
)

func TestAbs(t *testing.T) {
	imgSource := `192.168.1.63:5000/alpine:latest`
	imgTarget := `alpine`
	var err error

	Handlers.ImagePullSingle(imgSource, imgTarget)
	log.Printf("Succesfully retagged from %s to %s", imgSource, imgTarget)
	
	Handlers.UpdateClientContainerList()
	containerList := Handlers.ClientContainerList
	log.Printf("The container list = ")
	log.Printf("%v", containerList)
	if len(containerList) == 0 {
		t.Fatalf("The ContainerList is very small: %d", len(containerList))
	}
	
	var encoded []byte
	encoded, err = json.Marshal(containerList)
	if err != nil {
		t.Fatalf("Could not transform to JSON, %s", err)
	}
	log.Printf("The encoded stuff: %s", string(encoded))
	// for _, container := range containerList {
	// }

	Handlers.StopContainer("d5d85c25860a")
	
}