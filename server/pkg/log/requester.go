package log

import (
	"strings"
	"sync/atomic"
)

// IRequester allows callers to customize log prefix and call stack skip depth.
type IRequester interface {
	GetLogPrefix() string
	GetLogCallStackSkip() int
}

type requester struct {
	prefix string
	skip   int
}

func (r *requester) GetLogPrefix() string {
	return r.prefix
}

func (r *requester) GetLogCallStackSkip() int {
	if r.skip <= 0 {
		return DefaultSkipCall
	}
	return r.skip
}

// NewRequester constructs a requester with the given prefix and skip depth.
func NewRequester(prefix string, skip int) IRequester {
	cleanPrefix := strings.TrimSpace(prefix)
	return &requester{
		prefix: cleanPrefix,
		skip:   skip,
	}
}

// GetSkipCall returns a default requester using the configured global skip depth.
// When SetSkipCall has not been called, it falls back to DefaultSkipCall with an empty prefix.
func GetSkipCall() IRequester {
	skip := int(atomic.LoadInt32(&specSkip))
	if skip <= 0 {
		skip = DefaultSkipCall
	}
	return &requester{
		prefix: "",
		skip:   skip,
	}
}

func normalizeRequester(r IRequester) IRequester {
	if r != nil {
		return r
	}
	return GetSkipCall()
}
