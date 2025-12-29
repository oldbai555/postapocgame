// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package operation_log

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationLogExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperationLogExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationLogExportLogic {
	return &OperationLogExportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OperationLogExportLogic) OperationLogExport(w http.ResponseWriter, r *http.Request, req *types.OperationLogExportReq) error {
	if req == nil {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 查询所有符合条件的日志（不分页）
	operationLogRepo := repository.NewOperationLogRepository(l.svcCtx.Repository)
	list, _, err := operationLogRepo.FindPage(
		l.ctx,
		1,
		10000, // 导出最多10000条
		req.UserId,
		req.Username,
		req.OperationType,
		req.OperationObject,
		req.Method,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询操作日志失败", err)
	}

	// 设置响应头，返回 CSV 文件
	filename := fmt.Sprintf("操作日志_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 写入 BOM，确保 Excel 正确识别 UTF-8
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	// 创建 CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入表头
	headers := []string{"ID", "用户ID", "用户名", "操作类型", "操作对象", "请求方法", "请求路径", "请求参数", "响应状态码", "响应消息", "IP地址", "用户代理", "耗时(ms)", "创建时间"}
	if err := writer.Write(headers); err != nil {
		return errs.Wrap(errs.CodeInternalError, "写入CSV表头失败", err)
	}

	// 写入数据
	for _, log := range list {
		requestParams := ""
		if log.RequestParams.Valid {
			requestParams = log.RequestParams.String
		}

		row := []string{
			fmt.Sprintf("%d", log.Id),
			fmt.Sprintf("%d", log.UserId),
			log.Username,
			log.OperationType,
			log.OperationObject,
			log.Method,
			log.Path,
			requestParams,
			fmt.Sprintf("%d", log.ResponseCode),
			log.ResponseMsg,
			log.IpAddress,
			log.UserAgent,
			fmt.Sprintf("%d", log.Duration),
			time.Unix(log.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return errs.Wrap(errs.CodeInternalError, "写入CSV数据失败", err)
		}
	}

	return nil
}
