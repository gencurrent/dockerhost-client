package types

import (
	"encoding/json"
	"log"
)

type RequestStruct struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ResponseStruct struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

func UnmarshalResponseStruct(responseBody []byte) (*ResponseStruct, error) {
	var responseStruct *ResponseStruct 
	err := json.Unmarshal(responseBody, responseStruct)
	if err != nil {
		log.Printf("Error decoding the ResponseStruct: %s", err)
		return nil, err
	}
	return responseStruct, nil
}



func (r *ResponseStruct) Marshal() (string, error) {
	result, err := json.Marshal(r)
	if err != nil {
		log.Printf("Error encoding the ResponseStruct: %s", err)
	}
	return string(result), err
}
