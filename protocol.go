package jrpc

import (
	"bytes"
	"encoding/json"
	"io"
)

const (
	version         = "2.0"
	batchRequestKey = '['
	maxRequestBody  = 1 << 20
)

type id struct {
	value   json.RawMessage
	present bool
}

func (i *id) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v.(type) {
	case nil, string, float64:
		i.value = bytes.Clone(data)
		i.present = true
		return nil
	default:
		return errorInvalidRequest
	}
}

// Value returns the raw JSON id, or nil if the request is a notification.
func (i id) Value() json.RawMessage { return i.value }

// Request is a JSON-RPC 2.0 request. Use ID.Value() for the raw JSON id.
type Request struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      id              `json:"id"`
}

type response struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   error           `json:"error,omitempty"`
	ID      any             `json:"id"`
}

func parseBody(body []byte) (batch bool, requests []json.RawMessage, err error) {
	body = bytes.TrimSpace(body)
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

	if req.Version != version || req.Method == "" {
		return nil, errorInvalidRequest
	}

	if params := bytes.TrimSpace(req.Params); len(params) > 0 && params[0] != '[' && params[0] != '{' {
		return nil, errorInvalidRequest
	}

	return
}
