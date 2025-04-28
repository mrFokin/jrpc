package jrpc

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHTTPMethod(t *testing.T) {
	methods := []string{"GET", "HEAD", "PUT", "PATCH", "DELETE", "CONNECT", "TRACE"}

	e := echo.New()
	Endpoint(e, "/")
	for _, method := range methods {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equalf(t, http.StatusMethodNotAllowed, rec.Code, "HTTP method %s must return error", method)
	}
}

func TestContentType(t *testing.T) {
	types := []string{echo.MIMEApplicationJavaScript, echo.MIMEApplicationXML, echo.MIMEApplicationForm}

	e := echo.New()
	Endpoint(e, "/")
	for _, tp := range types {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(echo.HeaderContentType, tp)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equalf(t, http.StatusUnsupportedMediaType, rec.Code, "Request with Content-Type %s must return error 415 Unsupported Media Type", tp)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.NotEqual(t, http.StatusUnsupportedMediaType, rec.Code, "Request with Content-Type application/json mustn't return error")
}

func TestEmptyRequest(t *testing.T) {

	e := echo.New()
	Endpoint(e, "/")
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Empty request must return error 400 Bad Request")
}

func TestGroupRouter(t *testing.T) {

	e := echo.New()
	g := e.Group("/group")
	Endpoint(g, "/jrpc")
	req := httptest.NewRequest(http.MethodPost, "/group/jrpc", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Empty request must return error 400 Bad Request")
}

func TestHandler(t *testing.T) {
	testCases := []struct {
		when string
		req  string
		res  string
	}{
		{
			when: "when rpc call with an empty Array",
			req:  `[]`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`,
		},
		{
			when: "when rpc call with an invalid Batch (but not empty)",
			req:  `[1]`,
			res:  `[{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}]`,
		},
		{
			when: "when rpc call with invalid Batch",
			req:  `[1,2,3]`,
			res: `[` +
				`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null},` +
				`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null},` +
				`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}` +
				`]`,
		},
		{
			when: "when rpc call with invalid JSON",
			req:  `{"jsonrpc":"2.0","method":"foobar,"params":"bar","baz]`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`,
		},
		{
			when: "when rpc call Batch, invalid JSON",
			req: `[` +
				`{"jsonrpc":"2.0","method":"sum","params":[1,2,4],"id": "1"},` +
				`{"jsonrpc":"2.0","method"` +
				`]`,
			res: `{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`,
		},
		{
			when: "when rpc call with invalid Request object",
			req:  `{"jsonrpc":"2.0","method":1,"params":"bar"}`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`,
		},
		{
			when: "when rpc call with incorrect version",
			req:  `{"jsonrpc":"1.0","method":"test"}`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`,
		},
		{
			when: "when rpc call of non-existent method",
			req:  `{"jsonrpc":"2.0","method":"non.exist","id":"1"}`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":"1"}`,
		},
		{
			when: "when rpc call with positional parameters",
			req:  `{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":"1"}`,
			res:  `{"jsonrpc":"2.0","result":19,"id":"1"}`,
		},
		{
			when: "when rpc call with named parameters",
			req: `[` +
				`{"jsonrpc":"2.0","method":"subtract.object","params":{"subtrahend":23,"minuend":42},"id":"3"},` +
				`{"jsonrpc":"2.0","method":"subtract.object","params":{"minuend":42,"subtrahend":23},"id":"4"}` +
				`]`,
			res: `[` +
				`{"jsonrpc":"2.0","result":19,"id":"3"},` +
				`{"jsonrpc":"2.0","result":19,"id":"4"}` +
				`]`,
		},
		{
			when: "when rpc call with invalid parameters",
			req: `[` +
				`{"jsonrpc":"2.0","method":"subtract","params":[42,23,31],"id":"2"},` +
				`{"jsonrpc":"2.0","method":"subtract","params":42,"id":"3"}` +
				`]`,
			res: `[` +
				`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params","data":"There must be exactly 2 parameters"},"id":"2"},` +
				`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params"},"id":"3"}` +
				`]`,
		},
		{
			when: "when rpc method returns standart error",
			req:  `{"jsonrpc":"2.0","method":"error.standart","id":"3"}`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error","data":"Error message"},"id":"3"}`,
		},
		{
			when: "when rpc method return user-specific error",
			req:  `{"jsonrpc":"2.0","method":"error.user","id":17}`,
			res:  `{"jsonrpc":"2.0","error":{"code":256,"message":"User error","data":"Additional info"},"id":17}`,
		},
		{
			when: "when rpc call is a Notification",
			req:  `{"jsonrpc":"2.0","method":"subtract","params":[42,23]}`,
			res:  ``,
		},
		{
			when: "rpc call Batch (all notifications)",
			req: `[` +
				`{"jsonrpc":"2.0","method":"subtract","params":[42,23]},` +
				`{"jsonrpc":"2.0","method":"subtract","params":[23,42]}` +
				`]`,
			res: ``,
		},
		{
			when: "rpc call Batch",
			req: `[` +
				`{"jsonrpc":"2.0","method":"subtract","params":[23,42],"id":"1"},` +
				`{"jsonrpc":"2.0","method":"notify","params":[7]},` +
				`{"jsonrpc":"2.0","method":"subtract","params":[15,12],"id":"2"},` +
				`{"foo":"bar"},` +
				`{"jsonrpc":"2.0","method":"not.exist","params":{"name":"myself"},"id":"5"},` +
				`{"jsonrpc":"2.0","method":"get_data","id":"9"}` +
				`]`,
			res: `[` +
				`{"jsonrpc":"2.0","result":-19,"id":"1"},` +
				`{"jsonrpc":"2.0","result":3,"id":"2"},` +
				`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null},` +
				`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":"5"},` +
				`{"jsonrpc":"2.0","result":["hello",5],"id":"9"}` +
				`]`,
		},
	}

	e := echo.New()
	j := Endpoint(e, "/")
	j.Method("subtract", methodSubtract)
	j.Method("subtract.object", methodSubtractWithObject)
	j.Method("error.standart", methodWithStandartError)
	j.Method("error.user", methodWithUserError)
	j.Method("notify", methodNotify)
	j.Method("get_data", methodWithoutParams)

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equalf(t, http.StatusOK, rec.Code, "Invalid response code %s", tc.when)
		assert.Equalf(t, tc.res, strings.TrimRight(rec.Body.String(), "\n"), "Invalid response body %s", tc.when)
	}
}

func TestMiddleware(t *testing.T) {
	testCases := []struct {
		when string
		req  string
		res  string
	}{
		{
			when: "when first middleware return err",
			req:  `{"jsonrpc":"2.0","method":"middleware","params":123,"id":"8"}`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params","data":"First"},"id":"8"}`,
		},
		{
			when: "when second middleware return err",
			req:  `{"jsonrpc":"2.0","method":"middleware","params":234,"id":"9"}`,
			res:  `{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params","data":"Second"},"id":"9"}`,
		},
		{
			when: "when middlewares is ok",
			req:  `{"jsonrpc":"2.0","method":"middleware","params":321,"id":"10"}`,
			res:  `{"jsonrpc":"2.0","result":"Param: 321","id":"10"}`,
		},
	}

	e := echo.New()
	j := Endpoint(e, "/")
	j.Method("middleware", methodWithParameter, middlewareFirst, middlewareSecond)

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equalf(t, http.StatusOK, rec.Code, "Invalid response code %s", tc.when)
		assert.Equalf(t, tc.res, strings.TrimRight(rec.Body.String(), "\n"), "Invalid response body %s", tc.when)
	}
}

func methodSubtract(c Context) error {
	var p []int
	if err := c.Bind(&p); err != nil {
		return err
	}

	if len(p) != 2 {
		return NewErrorInvalidParams("There must be exactly 2 parameters")
	}

	return c.Result(p[0] - p[1])
}

type subtract struct {
	Subtrahend int `json:"subtrahend"`
	Minuend    int `json:"minuend"`
}

func methodSubtractWithObject(c Context) error {
	p := subtract{}
	if err := c.Bind(&p); err != nil {
		return err
	}

	return c.Result(p.Minuend - p.Subtrahend)
}

func methodWithStandartError(c Context) error {
	c.Result("Result must be ignored")
	return errors.New("Error message")
}

func methodWithUserError(c Context) error {
	c.Result("Result must be ignored")
	return NewError(256, "User error", "Additional info")
}

func methodNotify(c Context) error {
	return nil
}

func methodWithoutParams(c Context) error {
	res := []interface{}{"hello", 5}
	return c.Result(res)
}

func methodWithParameter(c Context) error {
	var i int
	if err := c.Bind(&i); err != nil {
		return err
	}

	return c.Result(fmt.Sprintf("Param: %d", i))
}

func middlewareFirst(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		var i int
		if err := c.Bind(&i); err != nil {
			return err
		}

		if i == 123 {
			return NewErrorInvalidParams("First")
		}

		return next(c)
	}
}

func middlewareSecond(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		var i int
		if err := c.Bind(&i); err != nil {
			return err
		}

		if i == 234 {
			return NewErrorInvalidParams("Second")
		}

		return next(c)
	}
}
