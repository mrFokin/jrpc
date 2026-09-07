# jrpc

JSON-RPC 2.0 for [Echo](https://echo.labstack.com).

- Single and batch requests
- Named and positional parameters
- Notifications
- Method-level middleware
- Typed handlers via `Method`

## API (v1)

- Echo v4 only. `Endpoint` registers `POST` and accepts `application/json` only.
- Register methods with `Method` (typed params and result) or `Handler` (manual `Bind` / `Result`).
- Return `NewError` or `NewErrorInvalidParams`. A plain `error` and a panic become Internal error (`-32603`).
- No `id` (notification) → HTTP 200 and an empty body. No `Result` → JSON `null`.
- Request id: `c.Request().ID.Value()`.
- Body limit: 1 MiB (`413`). Batch → JSON array of responses.
- Not in this package: `HandleMethod`, a `JRPC` interface, WebSocket, other HTTP frameworks.

## Install

```bash
go get github.com/mrFokin/jrpc
```

## Usage

Register a POST endpoint, then add methods.

```go
e := echo.New()
j := jrpc.Endpoint(e, "/rpc")

j.Method("subtract", func(c jrpc.Context, p []int) (int, error) {
    if len(p) != 2 {
        return 0, jrpc.NewErrorInvalidParams("exactly 2 parameters")
    }
    return p[0] - p[1], nil
})

e.Start(":8080")
```

```http
POST /rpc
Content-Type: application/json

{"jsonrpc":"2.0","method":"subtract","params":[42,23],"id":"1"}
```

```json
{"jsonrpc":"2.0","result":19,"id":"1"}
```

### Named parameters

`Method` binds params to a type and writes the return value as the result.

```go
type subtract struct {
    Subtrahend int `json:"subtrahend"`
    Minuend    int `json:"minuend"`
}

j.Method("subtract", func(c jrpc.Context, p subtract) (int, error) {
    return p.Minuend - p.Subtrahend, nil
})
```

```json
{"jsonrpc":"2.0","method":"subtract","params":{"minuend":42,"subtrahend":23},"id":"1"}
```

### Errors

Return `NewError` or `NewErrorInvalidParams` from a handler. A plain `error` and a panic become Internal error (`-32603`).

```go
j.Handler("fail", func(c jrpc.Context) error {
    return jrpc.NewError(256, "User error", "Additional info")
})
```

```json
{"jsonrpc":"2.0","error":{"code":256,"message":"User error","data":"Additional info"},"id":17}
```

### Middleware

Method middleware wraps a single handler, same as Echo but on the JSON-RPC method.

```go
func logMethod(next jrpc.HandlerFunc) jrpc.HandlerFunc {
    return func(c jrpc.Context) error {
        log.Println(c.Request().Method)
        return next(c)
    }
}

j.Handler("subtract", handler, logMethod)
```

Echo middleware still applies to the HTTP route via `Endpoint`.
