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
		if customerr.GetErrCode(err) < 0 {
			msg := fmt.Sprintf("go-routine err %v", err)
			log.Errorf(msg)
			// 错误通知
		}
	}()
}

func GoV2(fn func() error) {
	go func() {
		defer CatchPanic(func(err interface{}) {

		})
		err := fn()
		if customerr.GetErrCode(err) < 0 {
			msg := fmt.Sprintf("go-routine err:%v", err)
			log.Errorf(msg)
			// 错误通知
		}
	}()
}

func Run(fn func()) {
	defer CatchPanic(func(err interface{}) {})
	fn()
}

func CatchPanic(panicCallback func(err interface{})) {
	if err := recover(); err != any(nil) {
		log.Errorf("PROCESS PANIC: err %s", err)
		st := debug.Stack()
		if len(st) > 0 {
			log.Errorf("dump stack (%s):", err)
			lines := strings.Split(string(st), "\n")
			for _, line := range lines {
				log.Errorf("%s", line)
			}
		} else {
			log.Errorf("stack is empty (%s)", err)
		}
		if panicCallback != nil {
			panicCallback(err)
		}
	}
}
