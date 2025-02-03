package jrpc

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

// Context - json-rpc context
type Context interface {
	EchoContext() echo.Context
	Bind(interface{}) error
	Result(interface{}) error
	Request() *Request
}

type context struct {
	echo.Context
	request *Request
	//params json.RawMessage
	result json.RawMessage
}

// Bind parse input params
func (c *context) Bind(v interface{}) error {
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
