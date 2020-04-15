package types

import (
	"encoding/json"
	"log"
)

type RequestStruct struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (r *RequestStruct) marshal() (string, error) {
	result, err := json.Marshal(r)
	if err != nil {
		log.Printf("Error decoding the RequestStruct: %s", err)
	}
	return string(result), err
}
