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

// parseRawRequest parses a single request object. Member names are matched
// case-sensitively. On an Invalid Request error it still returns the request
// with ID set if the id member is valid, so the error response can echo it.
func parseRawRequest(raw json.RawMessage) (req *Request, err error) {
	var members map[string]json.RawMessage
	if err = json.Unmarshal(raw, &members); err != nil {
		return nil, err
	}

	req = &Request{Params: members["params"]}
	if v, ok := members["id"]; ok {
		if err = req.ID.UnmarshalJSON(v); err != nil {
			return nil, err
		}
	}

	if !unmarshalString(members["jsonrpc"], &req.Version) || req.Version != version ||
		!unmarshalString(members["method"], &req.Method) {
		return req, errorInvalidRequest
	}

	if params := bytes.TrimSpace(req.Params); len(params) > 0 && params[0] != '[' && params[0] != '{' {
		return req, errorInvalidRequest
	}

	return
}

// unmarshalString decodes v into dst and reports whether v is a JSON string.
func unmarshalString(v json.RawMessage, dst *string) bool {
	return len(v) > 0 && v[0] == '"' && json.Unmarshal(v, dst) == nil
}
