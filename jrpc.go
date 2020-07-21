// Package jrpc implements JSON-RPC 2.0 for labstack echo server
package jrpc

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandlerFunc - json-rpc handler
type HandlerFunc func(c Context) error

// HandlerMiddleware - json-rpc  middleware handler
type HandlerMiddleware func(raw json.RawMessage) (json.RawMessage, error)

// JRPC interface
type JRPC interface {
	Method(method string, handler HandlerFunc)
	Middleware(order int, handler HandlerMiddleware)
}

type jrpc struct {
	methods    map[string]HandlerFunc
	echo       *echo.Echo
	middleware map[int]HandlerMiddleware
}

// Endpoint create instance of jrpc route
func Endpoint(e *echo.Echo, path string, m ...echo.MiddlewareFunc) JRPC {
	j := &jrpc{
		methods:    make(map[string]HandlerFunc),
		echo:       e,
		middleware: make(map[int]HandlerMiddleware),
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
func (j *jrpc) Method(m string, h HandlerFunc) {
	j.methods[m] = h
}

// HandleMiddleware run jrpc middleware
func HandleMiddleware(middleware HandlerMiddleware, raw json.RawMessage) (json.RawMessage, Error) {
	raw, e := middleware(raw)
	if e != nil {
		return nil, e
	}
	return raw, nil
}

// Middleware add middleware for jrpc method
func (j *jrpc) Middleware(m int, h HandlerMiddleware) {
	j.middleware[m] = h
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

		for _, v := range j.middleware {
			raw, resp.Error = HandleMiddleware(v, raw)
		}

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
