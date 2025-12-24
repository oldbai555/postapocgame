package response

import (
	"context"
	"net/http"

	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Envelope 统一响应结构。
type Envelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// OkCtx 返回成功响应。
func OkCtx(ctx context.Context, w http.ResponseWriter, data interface{}) {
	resp := Envelope{
		Code: errs.CodeOK,
		Msg:  "OK",
		Data: data,
	}
	httpx.OkJsonCtx(ctx, w, resp)
}

// ErrorCtx 根据错误类型返回统一错误响应，并记录日志。
func ErrorCtx(ctx context.Context, w http.ResponseWriter, err error) {
	logger := logx.WithContext(ctx)
	if err == nil {
		OkCtx(ctx, w, nil)
		return
	}

	if be, ok := errs.FromError(err); ok {
		// 业务错误：按业务码返回，记录一行业务日志。
		logger.Errorf("biz error, code=%d, msg=%s, err=%+v", be.Code, be.Message, be.Err)
		writeJSON(ctx, w, http.StatusBadRequest, Envelope{
			Code: be.Code,
			Msg:  be.Message,
		})
		return
	}

	// 未知错误：统一映射为内部错误，避免泄露细节。
	logger.Errorf("internal error: %+v", err)
	writeJSON(ctx, w, http.StatusInternalServerError, Envelope{
		Code: errs.CodeInternalError,
		Msg:  "Internal Server Error",
	})
}

func writeJSON(ctx context.Context, w http.ResponseWriter, status int, v interface{}) {
	w.WriteHeader(status)
	httpx.OkJsonCtx(ctx, w, v)
}
