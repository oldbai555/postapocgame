package log

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	DefaultSkipCall = 4     // 默认跳过调用层级
	maxContentSize  = 15000 // 最大日志内容长度（防止磁盘溢出）
)

// CallInfo 调用信息
type CallInfo struct {
	File     string
	Line     int
	FuncName string
}

// GetCallInfo 获取调用信息
func GetCallInfo(skip int) *CallInfo {
	pc, callFile, callLine, ok := runtime.Caller(skip)
	var callFuncName string
	if ok {
		callFuncName = runtime.FuncForPC(pc).Name()
	}
	filePath, _ := getPackageName(callFuncName)
	filePath = path.Base(filePath)

	return &CallInfo{
		File:     path.Join(filePath, path.Base(callFile)),
		Line:     callLine,
		FuncName: "",
	}
}

// getPackageName 解析包名和函数名
func getPackageName(f string) (filePath string, fileFunc string) {
	slashIndex := strings.LastIndex(f, "/")
	filePath = f
	if slashIndex > 0 {
		idx := strings.Index(f[slashIndex:], ".") + slashIndex
		filePath = f[:idx]
		fileFunc = f[idx+1:]
		return
	}
	return
}

// buildTimeInfo 构建时间信息
func buildTimeInfo() string {
	return time.Now().Format("01-02 15:04:05.9999")
}

// buildTraceInfo 构建追踪信息（预留扩展）
func buildTraceInfo() string {
	return "UNKNOWN"
}

// buildContent 构建日志内容，防止内容过长
func buildContent(format string, v ...interface{}) string {
	content := fmt.Sprintf(format, v...)
	// 保护磁盘，限制日志内容长度
	if size := utf8.RuneCountInString(content); size > maxContentSize {
		content = string([]rune(content)[:maxContentSize]) + "..."
	}
	return content
}

// buildStackInfo 构建堆栈信息
func buildStackInfo() string {
	buf := make([]byte, 4096)
	l := runtime.Stack(buf, true)
	return string(buf[:l])
}

// buildCallInfo 构建调用信息字符串
func buildCallInfo(call *CallInfo) string {
	if call == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d %s", call.File, call.Line, call.FuncName)
}

// buildRecord 构建完整的日志记录
func buildRecord(curLv int, colorInfo, timeInfo, traceInfo, callerInfo, prefix, content string, goroutineTrace bool) string {
	var builder strings.Builder

	var header string
	if goroutineTrace {
		header = fmt.Sprintf("%s %s [%s] [trace: %s] ", timeInfo, prefix, callerInfo, traceInfo)
	} else {
		header = fmt.Sprintf("%s %s [%s] ", timeInfo, prefix, callerInfo)
	}

	builder.WriteString(fmt.Sprintf(colorInfo, header))
	builder.WriteString(content)

	if curLv >= StackLevel {
		builder.WriteString("\n")
		builder.WriteString(buildStackInfo())
	}

	builder.WriteString("\n")
	return builder.String()
}
