package jrpc

import "fmt"

var (
	errorParse          = NewError(-32700, "Parse error", nil)
	errorInvalidRequest = NewError(-32600, "Invalid Request", nil)
	errorMethodNotFound = NewError(-32601, "Method not found", nil)
)

// JRPCError is a JSON-RPC 2.0 error object.
type JRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *JRPCError) Error() string {
	return fmt.Sprintf("jrpc: code: %d, message: %s, data: %+v", e.Code, e.Message, e.Data)
}

// NewError returns a JSON-RPC error with the given code, message, and optional data.
func NewError(code int, message string, data interface{}) error {
	return &JRPCError{Code: code, Message: message, Data: data}
}

// NewErrorInvalidParams returns an Invalid params error (-32602).
func NewErrorInvalidParams(data interface{}) error {
	return &JRPCError{Code: -32602, Message: "Invalid params", Data: data}
}

func errorInternal(data interface{}) *JRPCError {
	return &JRPCError{Code: -32603, Message: "Internal error", Data: data}
}
