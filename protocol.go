package jrpc

import (
	"encoding/json"
	"io"
)

const (
	version         = "2.0"
	batchRequestKey = '['
	maxRequestBody  = 1 << 20
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

func parseBody(body []byte) (batch bool, requests []json.RawMessage, err error) {
	if len(body) == 0 {
		return false, nil, io.ErrUnexpectedEOF
	}

	if body[0] == batchRequestKey {
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
