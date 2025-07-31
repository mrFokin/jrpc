package jrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	version         = "2.0"
	batchRequestKey = '['
)

type Request struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      any             `json:"id"`
}

type response struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   error           `json:"error,omitempty"`
	ID      any             `json:"id"`
}

func parseBody(req *http.Request) (batch bool, requests []json.RawMessage, err error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = fmt.Errorf("read body: %w", err)
		return
	}

	if bytes.ContainsRune(body[:1], batchRequestKey) {
		batch = true
		err = json.Unmarshal(body, &requests)
	} else {
		requests = make([]json.RawMessage, 1)
		err = json.Unmarshal(body, &requests[0])
	}

	return
}

func parseRawRequest(raw json.RawMessage) (req *Request, err error) {
	req = &Request{}
	if err = json.Unmarshal(raw, req); err != nil {
		return nil, err
	}

	if req.Version != version {
		return nil, errorInvalidRequest
	}

	return
}
