package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type logger struct {
	name           string      // 日志名字
	level          int         // 日志等级
	bScreen        bool        // 是否打印屏幕
	path           string      // 目录
	prefix         string      // 标识
	maxFileSize    int64       // 文件大小
	perm           os.FileMode // 文件权限
	writer         *FileLoggerWriter
	goroutineTrace bool
}

type ILogger interface {
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Stackf(format string, v ...interface{})
	Tracef(format string, v ...interface{})
}

const (
	DefaultSkipCall = 3 //跳过等级
)

var (
	instance *logger
	initMu   sync.Mutex
	specSkip int
)

type CallInfoSt struct {
	File     string
	Line     int
	FuncName string
}

// SetLevel 设置日志级别
func SetLevel(l int) {
	if l > FatalLevel || l < TraceLevel {
		return
	}
	if nil != instance {
		instance.level = l
	}
}

func GetLevel() int {
	return instance.level
}

func SetSkipCall(skip int) {
	specSkip = skip
}

func GetSkipCall() int {
	if specSkip <= 0 {
		return DefaultSkipCall
	}
	return specSkip
}

func InitLogger(opts ...Option) ILogger {
	initMu.Lock()
	defer initMu.Unlock()

	if nil == instance {
		instance = &logger{
			goroutineTrace: true,
		}
	}
	for _, opt := range opts {
		opt(instance)
	}

	//log文件夹不存在则先创建
	if instance.path == "" {
		dir := os.Getenv("TLOGDIR")
		if len(dir) > 0 {
			instance.path = dir
		} else {
			instance.path = DefaultLogPath
		}
	}

	if instance.maxFileSize == 0 {
		instance.maxFileSize = LogFileMaxSize
	}

	if instance.perm == 0 {
		instance.perm = fileMode
	}

	if instance.writer == nil {
		instance.writer = NewFileLoggerWriter(instance.path, instance.maxFileSize, 5, OpenNewFileByByDateHour, 100000, instance.perm)

		go func() {
			err := instance.writer.Loop()
			if err != nil {
				panic(err)
			}
		}()
	}

	pID := os.Getpid()
	pIDStr := strconv.FormatInt(int64(pID), 10)
	Infof("======log:%v,pid:%v======logPath:%s======", instance.name, pIDStr, instance.path)

	return instance
}

func GetLogger() ILogger {
	return instance
}

func GetWriter() io.Writer {
	return instance.writer
}

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

func GetCallInfo(skip int) *CallInfoSt {
	pc, callFile, callLine, ok := runtime.Caller(skip)
	var callFuncName string
	if ok {
		// 拿到调用方法
		callFuncName = runtime.FuncForPC(pc).Name()
	}
	filePath, fileFunc := getPackageName(callFuncName)

	fileFunc = ""
	filePath = path.Base(filePath)

	return &CallInfoSt{
		File:     path.Join(filePath, path.Base(callFile)),
		Line:     callLine,
		FuncName: fileFunc,
	}
}

func Flush() {
	instance.writer.Flush()
}

func buildTimeInfo() string {
	return time.Now().Format("01-02 15:04:05.9999")
}

func buildTraceInfo() string {
	traceId := "UNKNOWN"
	return traceId
}

func buildContent(format string, v ...interface{}) string {
	content := fmt.Sprintf(format, v...)
	// protect disk
	if size := utf8.RuneCountInString(content); size > 15000 {
		content = string([]rune(content)[:15000]) + "..."
	}
	return content
}

func buildStackInfo() string {
	buf := make([]byte, 4096)
	l := runtime.Stack(buf, true)
	return string(buf[:l])
}

func buildCallInfo(call *CallInfoSt) string {
	if call == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d %s", call.File, call.Line, call.FuncName)
}

func buildRecord(curLv int, colorInfo, timeInfo, traceInfo, callerInfo, prefix, content string) string {
	var builder strings.Builder

	var header string
	if instance.goroutineTrace {
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

// 跟踪类型日志
func Tracef(format string, v ...interface{}) {
	instance.Tracef(format, v...)
}

// 调试类型日志
func Debugf(format string, v ...interface{}) {
	instance.Debugf(format, v...)
}

// 警告类型日志
func Warnf(format string, v ...interface{}) {
	instance.Warnf(format, v...)
}

// 程序信息类型日志
func Infof(format string, v ...interface{}) {
	instance.Infof(format, v...)
}

// 错误类型日志
func Errorf(format string, v ...interface{}) {
	instance.Errorf(format, v...)
}

// 堆栈debug日志
func Stackf(format string, v ...interface{}) {
	instance.Stackf(format, v...)
}

// 致命错误类型日志
func Fatalf(format string, v ...interface{}) {
	instance.Fatalf(format, v...)
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if l.level > WarnLevel {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(WarnLevel, warnColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write([]byte(record))

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if l.level > InfoLevel {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(InfoLevel, infoColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write([]byte(record))

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(ErrorLevel, errorColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write([]byte(record))

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	callInfo := GetCallInfo(GetSkipCall())
	content := buildContent(format, v...)
	record := buildRecord(FatalLevel, fatalColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, content)
	l.writer.Write([]byte(record))

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	tf := time.Now()
	os.WriteFile(fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic", dir, l.name, tf.Month(), tf.Day(), tf.Hour(), tf.Minute(), tf.Second()), []byte(record), fileMode)

	if l.bScreen {
		fmt.Printf("%s", record)
	}

	os.Exit(1)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if l.level > DebugLevel {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(DebugLevel, debugColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write([]byte(record))

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Stackf(format string, v ...interface{}) {
	if l.level > StackLevel {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(StackLevel, stackColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write([]byte(record))

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Tracef(format string, v ...interface{}) {
	if l.level > TraceLevel {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(TraceLevel, traceColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write([]byte(record))
	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Flush() {
	l.writer.Flush()
}
