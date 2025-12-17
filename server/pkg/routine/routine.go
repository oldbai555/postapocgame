package routine

import (
	"context"
	"fmt"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"runtime/debug"
	"strings"
)

func Go(ctx context.Context, logic func(ctx context.Context) error) {
	// 可以考虑放 traceId 链路追踪
	go func() {
		defer CatchPanic(func(err interface{}) {

		})
		err := logic(ctx)
		if err != nil {
			code := customerr.GetErrCode(err)
			if code < 0 {
				msg := fmt.Sprintf("go-routine err %v", err)
				log.Errorf("%s", msg)
				// 错误通知
			}
		}
	}()
}

func GoV2(fn func() error) {
	go func() {
		defer CatchPanic(func(err interface{}) {

		})
		err := fn()
		if err != nil {
			code := customerr.GetErrCode(err)
			if code < 0 {
				msg := fmt.Sprintf("go-routine err:%v", err)
				log.Errorf("%s", msg)
				// 错误通知
			}
		}
	}()
}

func Run(fn func()) {
	defer CatchPanic(func(err interface{}) {})
	fn()
}

func CatchPanic(panicCallback func(err interface{})) {
	if err := recover(); err != any(nil) {
		var sb strings.Builder
		sb.WriteString("PROCESS PANIC:\n")
		st := debug.Stack()
		if len(st) > 0 {
			sb.WriteString(fmt.Sprintf("dump stack (%s):\n", err))
			sb.WriteString(fmt.Sprintf("%s", string(st)))
		} else {
			sb.WriteString(fmt.Sprintf("stack is empty (%s)", err))
		}
		log.Errorf("%s", sb.String())
		if panicCallback != nil {
			panicCallback(err)
		}
	}
}
