// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit_log

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

type AuditLogExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditLogExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditLogExportLogic {
	return &AuditLogExportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuditLogExportLogic) AuditLogExport(w http.ResponseWriter, r *http.Request, req *types.AuditLogExportReq) error {
	if req == nil {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 查询所有符合条件的日志（不分页）
	auditLogRepo := repository.NewAuditLogRepository(l.svcCtx.Repository)
	list, _, err := auditLogRepo.FindPage(
		l.ctx,
		1,
		10000, // 导出最多10000条
		req.UserId,
		req.Username,
		req.AuditType,
		req.AuditObject,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询审计日志失败", err)
	}

	// 设置响应头，返回 CSV 文件
	filename := fmt.Sprintf("审计日志_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 写入 BOM，确保 Excel 正确识别 UTF-8
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	// 创建 CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入表头
	headers := []string{"ID", "用户ID", "用户名", "审计类型", "审计对象", "审计详情", "IP地址", "用户代理", "创建时间"}
	if err := writer.Write(headers); err != nil {
		return errs.Wrap(errs.CodeInternalError, "写入CSV表头失败", err)
	}

	// 写入数据
	for _, log := range list {
		auditDetail := ""
		if log.AuditDetail.Valid {
			auditDetail = log.AuditDetail.String
		}

		row := []string{
			fmt.Sprintf("%d", log.Id),
			fmt.Sprintf("%d", log.UserId),
			log.Username,
			log.AuditType,
			log.AuditObject,
			auditDetail,
			log.IpAddress,
			log.UserAgent,
			time.Unix(log.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return errs.Wrap(errs.CodeInternalError, "写入CSV数据失败", err)
		}
	}

	return nil
}
