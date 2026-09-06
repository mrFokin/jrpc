package jrpc_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mrFokin/jrpc"
)

func ExampleEndpoint() {
	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	j.Method("subtract", func(c jrpc.Context) error {
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

func ExampleHandle() {
	type subtract struct {
		Subtrahend int `json:"subtrahend"`
		Minuend    int `json:"minuend"`
	}

	e := echo.New()
	j := jrpc.Endpoint(e, "/rpc")
	jrpc.Handle(j, "subtract", func(c jrpc.Context, p subtract) (int, error) {
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
	j.Method("fail", func(c jrpc.Context) error {
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
