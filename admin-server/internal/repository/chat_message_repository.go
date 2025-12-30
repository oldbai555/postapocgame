package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type ChatMessageRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.ChatMessage, error)
	FindPage(ctx context.Context, page, pageSize int64, roomId string, userId uint64) ([]model.ChatMessage, int64, error)
	FindByChatID(ctx context.Context, page, pageSize int64, chatId uint64) ([]model.ChatMessage, int64, error)
	FindPrivateMessages(ctx context.Context, page, pageSize int64, currentUserId, targetUserId uint64) ([]model.ChatMessage, int64, error)
	Create(ctx context.Context, message *model.ChatMessage) error
	Update(ctx context.Context, message *model.ChatMessage) error
	DeleteByID(ctx context.Context, id uint64) error
}

type chatMessageRepository struct {
	model model.ChatMessageModel
	conn  sqlx.SqlConn
}

func NewChatMessageRepository(repo *Repository) ChatMessageRepository {
	return &chatMessageRepository{model: repo.ChatMessageModel, conn: repo.DB}
}

func (r *chatMessageRepository) FindByID(ctx context.Context, id uint64) (*model.ChatMessage, error) {
	return r.model.FindOne(ctx, id)
}

func (r *chatMessageRepository) FindPage(ctx context.Context, page, pageSize int64, roomId string, userId uint64) ([]model.ChatMessage, int64, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}

	if roomId != "" {
		conditions = append(conditions, "room_id = ?")
		args = append(args, roomId)
		// 群聊时，to_user_id 应该为 0
		conditions = append(conditions, "to_user_id = 0")
	} else if userId > 0 {
		// 私聊：查询与指定用户相关的消息（包括发送和接收）
		// 注意：这里需要结合当前用户ID来过滤，但当前方法没有当前用户ID参数
		// 暂时保持原逻辑，前端需要额外过滤
		conditions = append(conditions, "(from_user_id = ? OR to_user_id = ?)")
		args = append(args, userId, userId)
	}

	// 添加软删除条件
	conditions = append(conditions, "deleted_at = 0")

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `chat_message` %s", whereClause)
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `chat_message` %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)

	var list []model.ChatMessage
	err = r.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *chatMessageRepository) FindByChatID(ctx context.Context, page, pageSize int64, chatId uint64) ([]model.ChatMessage, int64, error) {
	// 根据 chatId 查询消息，如果 chatId == 0，则查询所有消息
	var whereClause string
	var args []interface{}
	if chatId == 0 {
		whereClause = "WHERE deleted_at = 0"
		args = []interface{}{}
	} else {
		whereClause = "WHERE chat_id = ? AND deleted_at = 0"
		args = []interface{}{chatId}
	}

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `chat_message` %s", whereClause)
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `chat_message` %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)

	var list []model.ChatMessage
	err = r.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *chatMessageRepository) FindPrivateMessages(ctx context.Context, page, pageSize int64, currentUserId, targetUserId uint64) ([]model.ChatMessage, int64, error) {
	// 查询当前用户和指定用户之间的私聊消息
	whereClause := "WHERE ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)) AND deleted_at = 0"
	args := []interface{}{currentUserId, targetUserId, targetUserId, currentUserId}

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `chat_message` %s", whereClause)
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM `chat_message` %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)

	var list []model.ChatMessage
	err = r.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *chatMessageRepository) Create(ctx context.Context, message *model.ChatMessage) error {
	_, err := r.model.Insert(ctx, message)
	return err
}

func (r *chatMessageRepository) Update(ctx context.Context, message *model.ChatMessage) error {
	return r.model.Update(ctx, message)
}

func (r *chatMessageRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}
