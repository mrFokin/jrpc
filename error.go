package jrpc

import "fmt"

var (
	errorParse          = NewError(-32700, "Parse error", nil)
	errorInvalidRequest = NewError(-32600, "Invalid Request", nil)
	errorMethodNotFound = NewError(-32601, "Method not found", nil)
)

// Error - json-rpc error interface
type Error interface {
	Error() string
}

type jrpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error - implements error interface
func (e *jrpcError) Error() string {
	return fmt.Sprintf("jrpc: code: %d, message: %s, data: %+v", e.Code, e.Message, e.Data)
}

// NewError creates instance of error
func NewError(code int, message string, data interface{}) error {
	return &jrpcError{Code: code, Message: message, Data: data}
}

// NewErrorInvalidParams - helper func for create instance of Invalid params error
func NewErrorInvalidParams(data interface{}) error {
	return &jrpcError{Code: -32602, Message: "Invalid params", Data: data}
}

func errorInternal(data interface{}) *jrpcError {
	return &jrpcError{Code: -32603, Message: "Internal error", Data: data}
}
