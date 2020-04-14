package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	Handlers "./handlers"
)

type RequestStruct struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

var RequestQueue []RequestStruct

func GetLastRequest() RequestStruct {
	if len(RequestQueue) == 0 {
		return RequestStruct{
			Request:   "status",
			Arguments: make(map[string]interface{}),
		}
	} else {
		element := RequestQueue[0]
		RequestQueue = RequestQueue[1:]
		return element
	}
}

var addr = flag.String("addr", "192.168.1.63:8000", "http service address")

func main() {

	args := make(map[string]interface{})
	args["newImage"] = "Just joking"
	toAppend := RequestStruct{
		Request:   "docker.image",
		Arguments: args,
	}
	RequestQueue = append(RequestQueue, toAppend)
	log.Printf("The last request: %v", GetLastRequest().Request)
	log.Printf("The last request: %v", GetLastRequest().Request)

	log.Printf("Started the client")

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "rpc"}
	log.Printf("Connecting to %s:", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Dial error: %s", err)
		return
	}

	defer c.Close()
	requestMessage := make(chan RequestStruct)
	done := make(chan struct{})

	messageRead := make(chan string)

	go func() {
		defer close(done)
		defer close(requestMessage)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			messageRead <- string(message)
			log.Printf("recv: %s", message)
		}
	}()

	// // mmap := make(map[string]interface{})

	// // messageStruct := &RequestStruct{
	// // 	Request:   "image.list",
	// // 	Arguments: mmap,
	// // }
	// // messageString, err := json.Marshal(messageStruct)
	// // if err != nil {
	// // 	log.Printf("Could not encode the json :%s", err)
	// // }
	ticker := time.NewTicker(time.Second)

	for {
		log.Printf("Waiting for 'done'")
		select {
		case <-done:
			return
		case <-ticker.C:
			log.Printf("Ticker detected")
			newMessage := <-messageRead
			log.Printf("The message after ticker == %s", newMessage)
			
			var resultStruct RequestStruct
			err := json.Unmarshal([]byte(newMessage), &resultStruct)
			if err != nil{
				log.Printf("Error decoding request from server: %s", err)
				panic(err)
				continue
			}
			// 		// messageRequest := <-requestMessage
			requestResult := Handlers.HandleRequest(resultStruct.Request, resultStruct.Arguments)

			// var responseType make(map[string]interface{}
			encoded, err := json.Marshal(requestResult)
			if err != nil {
				log.Printf("Error encoding the result into the JSON string: %s", err)
				panic(err)
			}

			err = c.WriteMessage(websocket.TextMessage, []byte(encoded))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}
		// // Read Image list
		// err = c.WriteMessage(websocket.TextMessage, []byte(messageString))
		// if err != nil {
		// 	log.Printf("Error writing a message : %s", err)
		// 	return
		// }
		// err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// if err != nil {
		// 	log.Printf("Error closing a socket: %s", err)
		// 	return
	}
}
