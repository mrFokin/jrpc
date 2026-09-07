package jrpc

import (
	"encoding/json"

	"github.com/labstack/echo/v5"
)

// Context is the JSON-RPC request context passed to method handlers.
type Context struct {
	// Echo is the Echo context of the HTTP request.
	Echo *echo.Context

	request *Request
	result  json.RawMessage
}

// Bind unmarshals request params into v. Missing or empty params leave v unchanged.
func (c *Context) Bind(v interface{}) error {
	if len(c.request.Params) == 0 {
		return nil
	}
	if err := json.Unmarshal(c.request.Params, v); err != nil {
		return NewErrorInvalidParams(nil)
	}
	return nil
}

// Result sets the JSON-RPC result. If Result is not called, the response is null.
func (c *Context) Result(v interface{}) error {
	res, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.result = res
	return nil
}

// Request returns the current JSON-RPC request. Use Request().ID.Value() for the raw JSON id.
func (c *Context) Request() *Request {
	return c.request
}
