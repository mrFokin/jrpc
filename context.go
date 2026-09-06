package jrpc

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

// Context is the JSON-RPC request context passed to method handlers.
type Context interface {
	// EchoContext returns the Echo context of the HTTP request.
	EchoContext() echo.Context
	// Bind unmarshals request params into v. Missing or empty params leave v unchanged.
	Bind(interface{}) error
	// Result sets the JSON-RPC result. If Result is not called, the response is null.
	Result(interface{}) error
	// Request returns the current JSON-RPC request. Use Request().ID.Value() for the raw JSON id.
	Request() *Request
}

type context struct {
	echo.Context
	request *Request
	result  json.RawMessage
}

// Bind parse input params
func (c *context) Bind(v interface{}) error {
	if len(c.request.Params) == 0 {
		return nil
	}
	if err := json.Unmarshal(c.request.Params, v); err != nil {
		return NewErrorInvalidParams(nil)
	}
	return nil
}

// Result return result of json-rpc method
func (c *context) Result(v interface{}) error {
	res, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.result = res
	return nil
}

// EchoContext return echo context
func (c *context) EchoContext() echo.Context {
	return c.Context
}

// Request return jrpc request
func (c *context) Request() *Request {
	return c.request
}
