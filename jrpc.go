// Package jrpc implements JSON-RPC 2.0 for labstack echo server
package jrpc

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandlerFunc - json-rpc handler
type HandlerFunc func(c Context) error

// MiddlewareFunc defines a function to process json-rpc middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// JRPC interface
type JRPC interface {
	Method(method string, handler HandlerFunc, middleware ...MiddlewareFunc)
}

type EchoRouter interface {
	Add(method, path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

type jrpc struct {
	methods map[string]HandlerFunc
	echo    EchoRouter
}

// Endpoint create instance of jrpc route
func Endpoint(e EchoRouter, path string, m ...echo.MiddlewareFunc) JRPC {
	j := &jrpc{
		methods: make(map[string]HandlerFunc),
		echo:    e,
	}

	j.echo.Add(http.MethodPost, path, j.jrpcHandler, m...)
	return j
}

// HandleMethod run jrpc handler
func HandleMethod(ec echo.Context, method HandlerFunc, params json.RawMessage) (json.RawMessage, Error) {
	cc := &context{Context: ec, params: params}
	if e := method(cc); e != nil {
		err, ok := e.(*jrpcError)
		if !ok {
			err = errorInternal(e.Error())
		}
		return nil, err
	}
	return cc.result, nil
}

// Method add handler for jrpc method
func (j *jrpc) Method(m string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	h := j.applyMiddleware(handler, middleware...)
	j.methods[m] = h
}

func (j *jrpc) applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func (j *jrpc) jrpcHandler(c echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) != echo.MIMEApplicationJSON {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType)
	}

	if c.Request().ContentLength == 0 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	batch, rawRequests, err := parseBody(c.Request())
	if err != nil {
		resp := response{
			Version: version,
			Error:   errorParse,
		}
		return c.JSON(http.StatusOK, resp)
	}

	if len(rawRequests) == 0 {
		resp := response{
			Version: version,
			Error:   errorInvalidRequest,
		}
		return c.JSON(http.StatusOK, resp)
	}

	responses := make([]response, 0, len(rawRequests))
	for _, raw := range rawRequests {
		resp := response{Version: version}

		req, err := parseRawRequest(raw)
		if err != nil {
			resp.Error = errorInvalidRequest
			responses = append(responses, resp)
			continue
		}

		resp.ID = req.ID

		method := j.methods[req.Method]
		if method == nil {
			resp.Error = errorMethodNotFound
			responses = append(responses, resp)
			continue
		}

		resp.Result, resp.Error = HandleMethod(c, method, req.Params)

		if resp.Error != nil || resp.ID != nil {
			responses = append(responses, resp)
		}
	}

	if len(responses) == 0 {
		return c.NoContent(http.StatusOK)
	}

	if batch {
		return c.JSON(http.StatusOK, responses)
	}

	return c.JSON(http.StatusOK, responses[0])
}
