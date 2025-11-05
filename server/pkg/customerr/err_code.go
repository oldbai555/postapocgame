package customerr

const (
	SUCCESS              = 0
	FAILURE              = 400
	ErrInvalidArg        = 1001
	ErrMysqlConfNotFound = 10001
	ErrOrmInitFailed     = 10002
	ErrDelayQueueOptErr  = 10003
	ErrStorageOptErr     = 10004
	ErrNotFound          = 10005
	ErrCustomError       = 10007 // 自定义错误
	ErrRecordNotFound    = 10008
	ErrHttpError         = 10009
	ErrWrapError         = 10010 // 包装错误
)

var (
	Success        = NewErr(SUCCESS, "ok")
	RecordNotFound = NewErr(ErrMysqlConfNotFound, "mysql conf not fond")
	OrmInitFailed  = NewErr(ErrOrmInitFailed, "orm init failed")
)
