# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2026-09-07

Echo v5. Module path is `github.com/mrFokin/jrpc/v2`. Echo v4 remains on `v1.1.x`.

### Changed

- Echo v5
- Module path `github.com/mrFokin/jrpc/v2`
- `Context` is a struct; handlers take `*Context`
- HTTP context is `c.Echo` (`*echo.Context`)

### Removed

- `Context` interface
- `EchoContext()` (`c.Echo` instead)

## [1.1.0] - 2026-09-07

Go 1.27. Typed registration is `Method` on `*JRPC`; manual `HandlerFunc` is `Handler`.

### Added

- API (v1) contract in README
- Godoc examples for middleware, notifications, and `EchoContext`

### Changed

- Go 1.27
- `Method` is a generic method on `*JRPC`: bind params to `P`, return `R`
- `j.Method(name, handler)` → `j.Handler(name, handler)`

### Removed

- Package-level `Handle` (`jrpc.Handle(j, name, fn)` → `j.Method(name, fn)`)

## [1.0.0] - 2026-09-06

JSON-RPC 2.0 request handling is closer to the spec. The public API is smaller and typed.

### Added

- `Handle` — generic typed method registration: bind params to `P`, return `R`
- Panic in a method is recovered as Internal error (`-32603`)
- Request body size limit of 1 MiB (`413 Request Entity Too Large`)
- `application/json` Content-Type is parsed (charset and case-insensitive media type)
- Examples in README and `example_test.go`

### Changed

- `Endpoint` returns `*JRPC` instead of a `JRPC` interface
- `Request.ID` is a typed value: use `ID.Value()` for the raw JSON id
- `Method` registration is thread-safe
- Wrapped `*JRPCError` is detected via `errors.As`
- Empty or missing `params`: `Bind` succeeds and leaves the destination unchanged
- A handler that does not call `Result` responds with JSON `null`
- Empty request body is a JSON-RPC Parse error (`-32700`), not HTTP 400

### Fixed

- Integer ids larger than 2^53 are preserved
- `null` and invalid id types follow the spec
- Notifications (no `id`) do not produce a response, including Method not found
- Non-structured `params` (not an array or object) are Invalid Request
- Empty method name is Invalid Request
- Leading/trailing whitespace around the request body is ignored

### Removed

- Public `HandleMethod`
- `JRPC` interface
- `Error` interface (use `*JRPCError` / `error`)

## [0.9.5] - 2025-07-31

### Changed

- `JRPCError` is exported

## [0.9.4] - 2025-02-03

### Added

- `Context.Request()` returns the current JSON-RPC request

## [0.9.3] - 2022-01-17

### Added

- `Context.EchoContext()`

## [0.9.2] - 2021-12-17

### Added

- Method-level middleware

## [0.9.1] - 2020-06-06

### Added

- Go modules
- Echo v4

[Unreleased]: https://github.com/mrFokin/jrpc/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/mrFokin/jrpc/compare/v1.1.0...v2.0.0
[1.1.0]: https://github.com/mrFokin/jrpc/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/mrFokin/jrpc/compare/v0.9.5...v1.0.0
[0.9.5]: https://github.com/mrFokin/jrpc/compare/v0.9.4...v0.9.5
[0.9.4]: https://github.com/mrFokin/jrpc/compare/v0.9.3...v0.9.4
[0.9.3]: https://github.com/mrFokin/jrpc/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/mrFokin/jrpc/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/mrFokin/jrpc/releases/tag/v0.9.1
