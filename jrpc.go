// Package jrpc implements JSON-RPC 2.0 for labstack echo server
package jrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// HandlerFunc - json-rpc handler
type HandlerFunc func(c Context) error

// MiddlewareFunc defines a function to process json-rpc middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type JRPC struct {
	mu      sync.RWMutex
	methods map[string]HandlerFunc
	echo    *echo.Echo
}

// Endpoint create instance of jrpc route
func Endpoint(e *echo.Echo, path string, m ...echo.MiddlewareFunc) *JRPC {
	j := &JRPC{
		methods: make(map[string]HandlerFunc),
		echo:    e,
	}

	j.echo.Add(http.MethodPost, path, j.jrpcHandler, m...)
	return j
}

func handleMethod(ec echo.Context, method HandlerFunc, request *Request) (result json.RawMessage, err error) {
	cc := &context{Context: ec, request: request}
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = errorInternal(fmt.Sprint(r))
		}
	}()
	if e := method(cc); e != nil {
		var rpcErr *JRPCError
		if !errors.As(e, &rpcErr) {
			rpcErr = errorInternal(e.Error())
		}
		return nil, rpcErr
	}
	if cc.result == nil {
		return json.RawMessage("null"), nil
	}
	return cc.result, nil
}

// Method add handler for jrpc method
func (j *JRPC) Method(m string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	h := j.applyMiddleware(handler, middleware...)
	j.mu.Lock()
	j.methods[m] = h
	j.mu.Unlock()
}

func Handle[P, R any](j *JRPC, name string, fn func(Context, P) (R, error), mw ...MiddlewareFunc) {
	j.Method(name, func(c Context) error {
		var p P
		if err := c.Bind(&p); err != nil {
			return err
		}
		res, err := fn(c, p)
		if err != nil {
			return err
		}
		return c.Result(res)
	}, mw...)
}

func (j *JRPC) lookup(name string) HandlerFunc {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.methods[name]
}

func (j *JRPC) applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func (j *JRPC) jrpcHandler(c echo.Context) error {
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if err != nil || mediaType != echo.MIMEApplicationJSON {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType)
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxRequestBody+1))
	if err != nil {
		return c.JSON(http.StatusOK, response{Version: version, Error: errorParse})
	}
	if int64(len(body)) > maxRequestBody {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge)
	}

	batch, rawRequests, err := parseBody(body)
	if err != nil {
		return c.JSON(http.StatusOK, response{Version: version, Error: errorParse})
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

		resp.ID = req.ID.value

		method := j.lookup(req.Method)
		if method == nil {
			if req.ID.present {
				resp.Error = errorMethodNotFound
				responses = append(responses, resp)
			}
			continue
		}

		resp.Result, resp.Error = handleMethod(c, method, req)
		if req.ID.present {
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
