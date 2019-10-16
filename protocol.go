package jrpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type request struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type response struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   error           `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

func parseBody(req *http.Request) (batch bool, requests []json.RawMessage, err error) {
	body, err := ioutil.ReadAll(req.Body)

	if bytes.ContainsRune(body[:1], batchRequestKey) {
		batch = true
		err = json.Unmarshal(body, &requests)
	} else {
		requests = make([]json.RawMessage, 1)
		err = json.Unmarshal(body, &requests[0])
	}

	return
}

func parseRawRequest(raw json.RawMessage) (req *request, err error) {
	req = &request{}
	if err = json.Unmarshal(raw, req); err != nil {
		return nil, err
	}

	if req.Version != version {
		return nil, errorInvalidRequest
	}

	return
}
