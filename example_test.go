package jrpc_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mrFokin/jrpc/v2"
)

func ExampleEndpoint() {
	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Method("subtract", func(c *jrpc.Context, p []int) (int, error) {
		if len(p) != 2 {
			return 0, jrpc.NewErrorInvalidParams("exactly 2 parameters")
		}
		return p[0] - p[1], nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":"1"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Print(strings.TrimRight(rec.Body.String(), "\n"))
	// Output: {"jsonrpc":"2.0","result":19,"id":"1"}
}

func ExampleJRPC_Handler() {
	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Handler("subtract", func(c *jrpc.Context) error {
		var p []int
		if err := c.Bind(&p); err != nil {
			return err
		}
		if len(p) != 2 {
			return jrpc.NewErrorInvalidParams("exactly 2 parameters")
		}
		return c.Result(p[0] - p[1])
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":"1"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Print(strings.TrimRight(rec.Body.String(), "\n"))
	// Output: {"jsonrpc":"2.0","result":19,"id":"1"}
}

func ExampleJRPC_Method() {
	type subtract struct {
		Subtrahend int `json:"subtrahend"`
		Minuend    int `json:"minuend"`
	}

	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Method("subtract", func(c *jrpc.Context, p subtract) (int, error) {
		return p.Minuend - p.Subtrahend, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"subtract","params":{"minuend":42,"subtrahend":23},"id":"1"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Print(strings.TrimRight(rec.Body.String(), "\n"))
	// Output: {"jsonrpc":"2.0","result":19,"id":"1"}
}

func ExampleNewError() {
	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Handler("fail", func(c *jrpc.Context) error {
		return jrpc.NewError(256, "User error", "Additional info")
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"fail","id":17}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Print(strings.TrimRight(rec.Body.String(), "\n"))
	// Output: {"jsonrpc":"2.0","error":{"code":256,"message":"User error","data":"Additional info"},"id":17}
}

func ExampleJRPC_Handler_middleware() {
	requireAuth := func(next jrpc.HandlerFunc) jrpc.HandlerFunc {
		return func(c *jrpc.Context) error {
			if c.Echo.Request().Header.Get("Authorization") == "" {
				return jrpc.NewError(401, "unauthorized", nil)
			}
			return next(c)
		}
	}

	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Handler("ping", func(c *jrpc.Context) error {
		return c.Result("ok")
	}, requireAuth)

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"ping","id":1}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Print(strings.TrimRight(rec.Body.String(), "\n"))
	// Output: {"jsonrpc":"2.0","error":{"code":401,"message":"unauthorized"},"id":1}
}

func Example_notification() {
	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Handler("ping", func(c *jrpc.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"ping"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Printf("%d %q", rec.Code, rec.Body.String())
	// Output: 200 ""
}

func ExampleContext() {
	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Handler("who", func(c *jrpc.Context) error {
		return c.Result(c.Echo.Request().Header.Get("X-User"))
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(
		`{"jsonrpc":"2.0","method":"who","id":1}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User", "denis")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	fmt.Print(strings.TrimRight(rec.Body.String(), "\n"))
	// Output: {"jsonrpc":"2.0","result":"denis","id":1}
}
