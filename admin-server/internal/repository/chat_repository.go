package repository

import (
	"context"
	"postapocgame/admin-server/internal/model"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ChatRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Chat, error)
	FindByUserID(ctx context.Context, userID uint64) ([]model.Chat, error)
	FindUsersByChatID(ctx context.Context, chatID uint64) ([]model.ChatUser, error)
	Create(ctx context.Context, chat *model.Chat) error
	Update(ctx context.Context, chat *model.Chat) error
	DeleteByID(ctx context.Context, id uint64) error
	FindPrivateChatByUserIDs(ctx context.Context, userID1, userID2 uint64) (*model.Chat, error)
	FindGroups(ctx context.Context, page, pageSize int64, name string) ([]model.Chat, int64, error)
	CountMembersByChatID(ctx context.Context, chatID uint64) (int64, error)
}

type chatRepository struct {
	model model.ChatModel
	conn  sqlx.SqlConn
}

func NewChatRepository(repo *Repository) ChatRepository {
	return &chatRepository{model: repo.ChatModel, conn: repo.DB}
}

func (r *chatRepository) FindByID(ctx context.Context, id uint64) (*model.Chat, error) {
	return r.model.FindOne(ctx, id)
}

func (r *chatRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.Chat, error) {
	// 通过chat_user关联表查询用户参与的所有聊天
	query := `
		SELECT c.* 
		FROM chat c
		INNER JOIN chat_user cu ON c.id = cu.chat_id
		WHERE cu.user_id = ? AND c.deleted_at = 0
		ORDER BY c.created_at DESC
	`
	var list []model.Chat
	err := r.conn.QueryRowsCtx(ctx, &list, query, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatRepository) FindUsersByChatID(ctx context.Context, chatID uint64) ([]model.ChatUser, error) {
	// 查询聊天中的所有用户
	query := `
		SELECT * 
		FROM chat_user
		WHERE chat_id = ?
		ORDER BY joined_at ASC
	`
	var list []model.ChatUser
	err := r.conn.QueryRowsCtx(ctx, &list, query, chatID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatRepository) Create(ctx context.Context, chat *model.Chat) error {
	_, err := r.model.Insert(ctx, chat)
	return err
}

func (r *chatRepository) Update(ctx context.Context, chat *model.Chat) error {
	return r.model.Update(ctx, chat)
}

func (r *chatRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *chatRepository) FindPrivateChatByUserIDs(ctx context.Context, userID1, userID2 uint64) (*model.Chat, error) {
	// 查找两个用户之间的私聊（type=1）
	// 私聊必须包含且仅包含这两个用户
	query := `
		SELECT c.* 
		FROM chat c
		INNER JOIN chat_user cu1 ON c.id = cu1.chat_id AND cu1.user_id = ?
		INNER JOIN chat_user cu2 ON c.id = cu2.chat_id AND cu2.user_id = ?
		WHERE c.type = 1 AND c.deleted_at = 0
		GROUP BY c.id
		HAVING COUNT(DISTINCT cu.user_id) = 2
		LIMIT 1
	`
	// 修复：使用子查询来正确计算用户数
	query = `
		SELECT c.* 
		FROM chat c
		WHERE c.type = 1 AND c.deleted_at = 0
		AND EXISTS (SELECT 1 FROM chat_user cu1 WHERE cu1.chat_id = c.id AND cu1.user_id = ?)
		AND EXISTS (SELECT 1 FROM chat_user cu2 WHERE cu2.chat_id = c.id AND cu2.user_id = ?)
		AND (SELECT COUNT(*) FROM chat_user cu WHERE cu.chat_id = c.id) = 2
		LIMIT 1
	`
	var chat model.Chat
	err := r.conn.QueryRowCtx(ctx, &chat, query, userID1, userID2)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// FindGroups 查询群组列表（分页、搜索）
func (r *chatRepository) FindGroups(ctx context.Context, page, pageSize int64, name string) ([]model.Chat, int64, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}

	// 只查询群组（type=2）
	conditions = append(conditions, "type = 2")
	conditions = append(conditions, "deleted_at = 0")

	// 名称搜索
	if name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM `chat` " + whereClause
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := "SELECT * FROM `chat` " + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	var list []model.Chat
	err = r.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CountMembersByChatID 统计群组成员数量
func (r *chatRepository) CountMembersByChatID(ctx context.Context, chatID uint64) (int64, error) {
	query := "SELECT COUNT(*) FROM `chat_user` WHERE chat_id = ?"
	var count int64
	err := r.conn.QueryRowCtx(ctx, &count, query, chatID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
