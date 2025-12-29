// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login_log

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

type LoginLogExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogExportLogic {
	return &LoginLogExportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogExportLogic) LoginLogExport(w http.ResponseWriter, r *http.Request, req *types.LoginLogExportReq) error {
	if req == nil {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 查询所有符合条件的日志（不分页）
	loginLogRepo := repository.NewLoginLogRepository(l.svcCtx.Repository)
	list, _, err := loginLogRepo.FindPage(
		l.ctx,
		1,
		10000, // 导出最多10000条
		req.UserId,
		req.Username,
		req.Status,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询登录日志失败", err)
	}

	// 设置响应头，返回 CSV 文件
	filename := fmt.Sprintf("登录日志_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 写入 BOM，确保 Excel 正确识别 UTF-8
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	// 创建 CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入表头
	headers := []string{"ID", "用户ID", "用户名", "IP地址", "登录地点", "浏览器", "操作系统", "用户代理", "登录状态", "登录消息", "登录时间", "登出时间", "创建时间"}
	if err := writer.Write(headers); err != nil {
		return errs.Wrap(errs.CodeInternalError, "写入CSV表头失败", err)
	}

	// 写入数据
	for _, log := range list {
		statusText := "失败"
		if log.Status == 1 {
			statusText = "成功"
		}
		loginAtStr := ""
		if log.LoginAt > 0 {
			loginAtStr = time.Unix(log.LoginAt, 0).Format("2006-01-02 15:04:05")
		}
		logoutAtStr := ""
		if log.LogoutAt > 0 {
			logoutAtStr = time.Unix(log.LogoutAt, 0).Format("2006-01-02 15:04:05")
		}
		createdAtStr := ""
		if log.CreatedAt > 0 {
			createdAtStr = time.Unix(log.CreatedAt, 0).Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", log.Id),
			fmt.Sprintf("%d", log.UserId),
			log.Username,
			log.IpAddress,
			log.Location,
			log.Browser,
			log.Os,
			log.UserAgent,
			statusText,
			log.Message,
			loginAtStr,
			logoutAtStr,
			createdAtStr,
		}
		if err := writer.Write(row); err != nil {
			return errs.Wrap(errs.CodeInternalError, "写入CSV数据失败", err)
		}
	}

	return nil
}
