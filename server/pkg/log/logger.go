package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

// GetSkipCall 获取跳过调用层级
func GetSkipCall() int {
	skip := atomic.LoadInt32(&specSkip)
	if skip <= 0 {
		return DefaultSkipCall
	}
	return int(skip)
}

// Flush 刷新日志
func Flush() {
	if instance != nil && instance.writer != nil {
		instance.writer.Flush()
	}
}

// 日志级别方法实现
func (l *logger) Warnf(format string, v ...interface{}) {
	if atomic.LoadInt32(&l.level) > int32(WarnLevel) {
		return
	}
	l.writeLog(WarnLevel, warnColor, format, v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	if atomic.LoadInt32(&l.level) > int32(InfoLevel) {
		return
	}
	l.writeLog(InfoLevel, infoColor, format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if atomic.LoadInt32(&l.level) > int32(ErrorLevel) {
		return
	}
	l.writeLog(ErrorLevel, errorColor, format, v...)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	callInfo := GetCallInfo(GetSkipCall())
	content := buildContent(format, v...)
	record := buildRecord(
		FatalLevel,
		fatalColor,
		buildTimeInfo(),
		buildTraceInfo(),
		buildCallInfo(callInfo),
		l.prefix,
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
	if atomic.LoadInt32(&l.level) > int32(DebugLevel) {
		return
	}
	l.writeLog(DebugLevel, debugColor, format, v...)
}

func (l *logger) Stackf(format string, v ...interface{}) {
	if atomic.LoadInt32(&l.level) > int32(StackLevel) {
		return
	}
	l.writeLog(StackLevel, stackColor, format, v...)
}

func (l *logger) Tracef(format string, v ...interface{}) {
	if atomic.LoadInt32(&l.level) > int32(TraceLevel) {
		return
	}
	l.writeLog(TraceLevel, traceColor, format, v...)
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
func (l *logger) writeLog(level int, colorInfo, format string, v ...interface{}) {
	if l.closed.Load() {
		return
	}

	callInfo := GetCallInfo(GetSkipCall())
	record := buildRecord(
		level,
		colorInfo,
		buildTimeInfo(),
		buildTraceInfo(),
		buildCallInfo(callInfo),
		l.prefix,
		buildContent(format, v...),
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

func Debugf(format string, v ...interface{}) {
	if instance != nil {
		instance.Debugf(format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if instance != nil {
		instance.Warnf(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if instance != nil {
		instance.Infof(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if instance != nil {
		instance.Errorf(format, v...)
	}
}

func Stackf(format string, v ...interface{}) {
	if instance != nil {
		instance.Stackf(format, v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	if instance != nil {
		instance.Fatalf(format, v...)
	}
}
