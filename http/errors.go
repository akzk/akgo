package http

import "net/http"

// Error 包含发生该错误时的返回码与返回信息
type Error struct {
	Code int    // 响应码
	Msg  string // 响应英文简述
	Text string // 响应中文详述
}

// NewError 返回一种新的错误类型
func NewError(code int, msg string, text string) *Error {
	return &Error{code, msg, text}
}

// DefaultErr 未注册的错误类型，直接输入错误信息，返回Error类型指针
func DefaultErr(info string) *Error {
	return &Error{403, "UnRegisteredError", info}
}

// 错误类型集合
var (
	ErrMethodNotSupport = NewError(http.StatusForbidden, "MethodNotSupport", "该URL不支持当前模式的访问")
)
