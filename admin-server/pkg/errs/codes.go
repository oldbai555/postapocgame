package errs

const (
	// CodeOK 成功。
	CodeOK = 0

	// 通用错误码（1xxxx）
	CodeInternalError = 10001
	CodeBadRequest    = 10002
	CodeUnauthorized  = 10003
	CodeForbidden     = 10004
	CodeNotFound      = 10005
)
