package customerr

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	// errorTagMap 错误码到Tag的映射，由调用方注册
	errorTagMap      = make(map[int32]string)
	errorTagMapMu    sync.RWMutex
	defaultErrorCode int32 = 1000 // 默认内部错误码，可配置
)

// RegisterErrorTag 注册错误码到Tag的映射（线程安全）
// 建议在应用启动时统一注册，例如从protocol枚举自动注册
func RegisterErrorTag(code int32, tag string) {
	errorTagMapMu.Lock()
	defer errorTagMapMu.Unlock()
	errorTagMap[code] = tag
}

// RegisterErrorTags 批量注册错误码映射
func RegisterErrorTags(tags map[int32]string) {
	errorTagMapMu.Lock()
	defer errorTagMapMu.Unlock()
	for code, tag := range tags {
		errorTagMap[code] = tag
	}
}

// SetDefaultErrorCode 设置默认错误码（当Wrap未指定code时使用）
func SetDefaultErrorCode(code int32) {
	errorTagMapMu.Lock()
	defer errorTagMapMu.Unlock()
	defaultErrorCode = code
}

// GetErrorTag 获取错误码对应的Tag（线程安全）
func GetErrorTag(code int32) string {
	errorTagMapMu.RLock()
	defer errorTagMapMu.RUnlock()
	tag, ok := errorTagMap[code]
	if !ok {
		return "UNKNOWN"
	}
	return tag
}

// CustomErr 全局结构体
// 带错误码，英文tag，详细描述信息，还有调用堆栈关键点
// 便于统一日志检索与定位
//
// 建议每条业务错误都通过NewErrorByCode生成

type CustomErr struct {
	Code    int32
	Tag     string
	Message string
	Caller  string // 调用出错的文件+行号
}

func (e *CustomErr) Error() string {
	return fmt.Sprintf("[Err-%d:%s][%s] %s", e.Code, e.Tag, e.Caller, e.Message)
}

// NewError 是 NewErrorByCode 的语法糖，使用默认错误码
func NewError(format string, args ...interface{}) error {
	return NewErrorByCode(defaultErrorCode, format, args...)
}

// NewErrorByCode 创建带定位的错误，自动采集1级Caller
func NewErrorByCode(code int32, format string, args ...interface{}) error {
	tag := GetErrorTag(code)
	callSite := caller(2)
	var detail = format
	if len(args) > 0 {
		detail = fmt.Sprintf(format, args...)
	}
	return &CustomErr{
		Code:    code,
		Tag:     tag,
		Message: detail,
		Caller:  callSite,
	}
}

// caller 获取调用者文件+行号
func caller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// Wrap 增强版包装底层错误为同样结构
func Wrap(err error, code ...int32) error {
	if err == nil {
		return nil
	}
	errorTagMapMu.RLock()
	defaultCode := defaultErrorCode
	errorTagMapMu.RUnlock()

	cd := defaultCode
	if len(code) > 0 {
		cd = code[0]
	}
	tag := GetErrorTag(cd)
	return &CustomErr{
		Code:    cd,
		Tag:     tag,
		Message: err.Error(),
		Caller:  caller(2),
	}
}
