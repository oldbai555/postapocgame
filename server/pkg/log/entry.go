package log

// Entry 支持结构化字段的日志入口
type Entry struct {
	base   *logger
	fields Fields
}

// WithFields 创建一个带字段的 Entry
func (e *Entry) WithFields(fields Fields) *Entry {
	if e == nil {
		return WithFields(fields)
	}
	return &Entry{
		base:   e.base,
		fields: mergeFields(e.fields, fields),
	}
}

func (e *Entry) log(level int, format string, v ...interface{}) {
	if e == nil || e.base == nil {
		return
	}
	e.base.writeLog(level, e.fields, nil, format, v...)
}

func (e *Entry) Infof(format string, v ...interface{}) {
	e.log(InfoLevel, format, v...)
}

func (e *Entry) Debugf(format string, v ...interface{}) {
	e.log(DebugLevel, format, v...)
}

func (e *Entry) Warnf(format string, v ...interface{}) {
	e.log(WarnLevel, format, v...)
}

func (e *Entry) Errorf(format string, v ...interface{}) {
	e.log(ErrorLevel, format, v...)
}

func (e *Entry) Tracef(format string, v ...interface{}) {
	e.log(TraceLevel, format, v...)
}

func (e *Entry) Stackf(format string, v ...interface{}) {
	e.log(StackLevel, format, v...)
}

func (e *Entry) Fatalf(format string, v ...interface{}) {
	if e == nil || e.base == nil {
		return
	}
	e.base.logFatal(e.fields, format, v...)
}
