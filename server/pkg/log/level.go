package log

// 日志级别定义
const (
	TraceLevel = iota // Trace级别
	DebugLevel        // Debug级别
	InfoLevel         // Info级别
	WarnLevel         // Warn级别
	ErrorLevel        // Error级别
	StackLevel        // Stack级别
	FatalLevel        // Fatal级别
)
