package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ILogger 日志接口
type ILogger interface {
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Stackf(format string, v ...interface{})
	Tracef(format string, v ...interface{})
	Flush()
}

// logger 日志器实现
type logger struct {
	name           string
	level          int32 // 使用 atomic 操作
	bScreen        bool
	path           string
	prefix         string
	maxFileSize    int64
	perm           os.FileMode
	writer         *FileLoggerWriter
	goroutineTrace bool
	enableColor    bool
	mu             sync.RWMutex // 保护配置变更
	closed         atomic.Bool  // 是否已关闭
}

var (
	instance *logger
	initMu   sync.Mutex
	specSkip int32 // 使用 atomic 操作
)

// InitLogger 初始化日志器
func InitLogger(opts ...Option) ILogger {
	initMu.Lock()
	defer initMu.Unlock()

	if instance == nil {
		instance = &logger{
			goroutineTrace: true,
			enableColor:    true,
			level:          int32(InfoLevel),
		}
	}

	for _, opt := range opts {
		opt(instance)
	}

	// 设置默认值
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

	instance.applyEnvOverrides()

	// 初始化 writer
	if instance.writer == nil {
		instance.writer = NewFileLoggerWriter(
			instance.path,
			instance.maxFileSize,
			5,
			func(logName string, lastOpenFileTime *time.Time, isNeverOpenFile bool) (string, bool) {
				return OpenNewFileByByDateHour(logName, lastOpenFileTime, isNeverOpenFile)
			},
			100000,
			instance.perm,
		)
		instance.writer.SetLogName(instance.name)

		// 启动写入协程
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// 写入失败时降级到 stderr
					fmt.Fprintf(os.Stderr, "log writer panic: %v\n", r)
				}
			}()
			if err := instance.writer.Loop(); err != nil {
				fmt.Fprintf(os.Stderr, "log writer error: %v\n", err)
			}
		}()
	}

	pID := os.Getpid()
	pIDStr := strconv.FormatInt(int64(pID), 10)
	// 直接调用实例方法，避免循环依赖
	instance.Infof("======log:%v,pid:%v======logPath:%s======", instance.name, pIDStr, instance.path)

	return instance
}

// GetLogger 获取日志器实例
func GetLogger() ILogger {
	if instance == nil {
		// 如果未初始化，使用默认配置初始化
		return InitLogger()
	}
	return instance
}

// GetWriter 获取写入器
func GetWriter() io.Writer {
	if instance == nil {
		return nil
	}
	return instance.writer
}

// SetLevel 设置日志级别（线程安全）
func SetLevel(l int) {
	if l > FatalLevel || l < TraceLevel {
		return
	}
	if instance != nil {
		atomic.StoreInt32(&instance.level, int32(l))
	}
}

// GetLevel 获取日志级别
func GetLevel() int {
	if instance == nil {
		return InfoLevel
	}
	return int(atomic.LoadInt32(&instance.level))
}

// SetSkipCall 设置跳过调用层级
func SetSkipCall(skip int) {
	atomic.StoreInt32(&specSkip, int32(skip))
}

// Flush 刷新日志
func Flush() {
	if instance != nil && instance.writer != nil {
		instance.writer.Flush()
	}
}

// 日志级别方法实现
func (l *logger) Warnf(format string, v ...interface{}) {
	l.writeLog(WarnLevel, nil, nil, format, v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.writeLog(InfoLevel, nil, nil, format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.writeLog(ErrorLevel, nil, nil, format, v...)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.logFatal(nil, format, v...)
}

func (l *logger) logFatal(fields Fields, format string, v ...interface{}) {
	l.logFatalWithRequester(fields, nil, format, v...)
}

func (l *logger) logFatalWithRequester(fields Fields, requester IRequester, format string, v ...interface{}) {
	req := normalizeRequester(requester)
	callInfo := GetCallInfo(req.GetLogCallStackSkip())
	content := buildContent(format, v...)
	if formatted := formatFields(fields); formatted != "" {
		content += formatted
	}
	record := buildRecord(
		FatalLevel,
		l.levelTemplate(FatalLevel),
		buildTimeInfo(),
		buildTraceInfo(),
		buildCallInfo(callInfo),
		mergePrefixes(l.prefix, req.GetLogPrefix()),
		content,
		l.goroutineTrace,
	)

	l.mu.RLock()
	writer := l.writer
	bScreen := l.bScreen
	name := l.name
	l.mu.RUnlock()

	if writer != nil {
		writer.Write([]byte(record))
		writer.Flush()
	}

	// 写入 panic 文件
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	tf := time.Now()
	panicFile := fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic",
		dir, name, tf.Month(), tf.Day(), tf.Hour(), tf.Minute(), tf.Second())
	os.WriteFile(panicFile, []byte(record), fileMode)

	if bScreen {
		fmt.Printf("%s", record)
	}

	os.Exit(1)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.writeLog(DebugLevel, nil, nil, format, v...)
}

func (l *logger) Stackf(format string, v ...interface{}) {
	l.writeLog(StackLevel, nil, nil, format, v...)
}

func (l *logger) Tracef(format string, v ...interface{}) {
	l.writeLog(TraceLevel, nil, nil, format, v...)
}

func (l *logger) Flush() {
	l.mu.RLock()
	writer := l.writer
	l.mu.RUnlock()
	if writer != nil {
		writer.Flush()
	}
}

// writeLog 统一的日志写入方法
func (l *logger) writeLog(level int, fields Fields, requester IRequester, format string, v ...interface{}) {
	if l.closed.Load() || atomic.LoadInt32(&l.level) > int32(level) {
		return
	}

	req := normalizeRequester(requester)
	content := buildContent(format, v...)
	if formatted := formatFields(fields); formatted != "" {
		content += formatted
	}

	callInfo := GetCallInfo(req.GetLogCallStackSkip())
	record := buildRecord(
		level,
		l.levelTemplate(level),
		buildTimeInfo(),
		buildTraceInfo(),
		buildCallInfo(callInfo),
		mergePrefixes(l.prefix, req.GetLogPrefix()),
		content,
		l.goroutineTrace,
	)

	l.mu.RLock()
	writer := l.writer
	bScreen := l.bScreen
	l.mu.RUnlock()

	if writer != nil {
		writer.Write([]byte(record))
	}

	if bScreen {
		fmt.Printf("%s", record)
	}
}

// 全局函数（保持对外接口不变）
func Tracef(format string, v ...interface{}) {
	if instance != nil {
		instance.Tracef(format, v...)
	}
}

func TracefWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.writeLog(TraceLevel, nil, requester, format, v...)
}

func Debugf(format string, v ...interface{}) {
	if instance != nil {
		instance.Debugf(format, v...)
	}
}

func DebugfWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.writeLog(DebugLevel, nil, requester, format, v...)
}

func Warnf(format string, v ...interface{}) {
	if instance != nil {
		instance.Warnf(format, v...)
	}
}

func WarnfWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.writeLog(WarnLevel, nil, requester, format, v...)
}

func Infof(format string, v ...interface{}) {
	if instance != nil {
		instance.Infof(format, v...)
	}
}

func InfofWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.writeLog(InfoLevel, nil, requester, format, v...)
}

func Errorf(format string, v ...interface{}) {
	if instance != nil {
		instance.Errorf(format, v...)
	}
}

func ErrorfWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.writeLog(ErrorLevel, nil, requester, format, v...)
}

func Stackf(format string, v ...interface{}) {
	if instance != nil {
		instance.Stackf(format, v...)
	}
}

func StackfWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.writeLog(StackLevel, nil, requester, format, v...)
}

func Fatalf(format string, v ...interface{}) {
	if instance != nil {
		instance.logFatal(nil, format, v...)
	}
}

func FatalfWithRequester(requester IRequester, format string, v ...interface{}) {
	if instance == nil {
		return
	}
	instance.logFatalWithRequester(nil, requester, format, v...)
}

func InfofWithFields(fields Fields, format string, v ...interface{}) {
	if instance == nil {
		InitLogger()
	}
	instance.writeLog(InfoLevel, fields, nil, format, v...)
}

func ErrorfWithFields(fields Fields, format string, v ...interface{}) {
	if instance == nil {
		InitLogger()
	}
	instance.writeLog(ErrorLevel, fields, nil, format, v...)
}

func DebugfWithFields(fields Fields, format string, v ...interface{}) {
	if instance == nil {
		InitLogger()
	}
	instance.writeLog(DebugLevel, fields, nil, format, v...)
}

func WithFields(fields Fields) *Entry {
	if instance == nil {
		InitLogger()
	}
	return &Entry{
		base:   instance,
		fields: cloneFields(fields),
	}
}

func (l *logger) levelTemplate(level int) string {
	if l.enableColor {
		return levelColorTemplate(level)
	}
	return levelPlainTemplate(level)
}

func levelColorTemplate(level int) string {
	switch level {
	case TraceLevel:
		return traceColor
	case DebugLevel:
		return debugColor
	case InfoLevel:
		return infoColor
	case WarnLevel:
		return warnColor
	case ErrorLevel:
		return errorColor
	case StackLevel:
		return stackColor
	case FatalLevel:
		return fatalColor
	default:
		return infoColor
	}
}

func levelPlainTemplate(level int) string {
	switch level {
	case TraceLevel:
		return tracePlain
	case DebugLevel:
		return debugPlain
	case InfoLevel:
		return infoPlain
	case WarnLevel:
		return warnPlain
	case ErrorLevel:
		return errorPlain
	case StackLevel:
		return stackPlain
	case FatalLevel:
		return fatalPlain
	default:
		return infoPlain
	}
}

func (l *logger) applyEnvOverrides() {
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		if parsed, ok := parseLevel(strings.ToLower(lvl)); ok {
			atomic.StoreInt32(&l.level, int32(parsed))
		}
	}
	if color := os.Getenv("LOG_COLOR"); color != "" {
		l.enableColor = !isFalse(color)
	}
}

func parseLevel(val string) (int, bool) {
	switch strings.ToLower(val) {
	case "trace":
		return TraceLevel, true
	case "debug":
		return DebugLevel, true
	case "info":
		return InfoLevel, true
	case "warn", "warning":
		return WarnLevel, true
	case "error":
		return ErrorLevel, true
	case "stack":
		return StackLevel, true
	case "fatal":
		return FatalLevel, true
	default:
		return 0, false
	}
}

func isFalse(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "0", "false", "no", "off":
		return true
	default:
		return false
	}
}

func mergePrefixes(basePrefix, extraPrefix string) string {
	base := strings.TrimSpace(basePrefix)
	extra := strings.TrimSpace(extraPrefix)

	switch {
	case base == "":
		return extra
	case extra == "":
		return base
	default:
		return base + " " + extra
	}
}
