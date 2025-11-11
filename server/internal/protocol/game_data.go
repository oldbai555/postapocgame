package protocol

import "postapocgame/server/pkg/tool"

// ItemType 道具类型
type ItemType uint32

const (
	ItemTypeMoney     ItemType = 1 // 货币
	ItemTypeEquipment ItemType = 2 // 装备
	ItemTypeMaterial  ItemType = 3 // 材料
	ItemTypeConsume   ItemType = 4 // 消耗品
)

// Item 道具
type Item struct {
	ItemId uint32   `json:"itemId"` // 道具Id
	Type   ItemType `json:"type"`   // 道具类型
	Count  uint32   `json:"count"`  // 数量
}

// SelectRoleRequest 选择角色请求
type SelectRoleRequest struct {
	RoleId uint64 `json:"roleId"` // 角色Id
}

// ReconnectRequest 重连请求
type ReconnectRequest struct {
	ReconnectKey string `json:"reconnectKey"` // 重连密钥
}

// ReconnectResponse 重连响应
type ReconnectResponse struct {
	Success bool   `json:"success"`
	ErrMsg  string `json:"errMsg,omitempty"`
}

// LoginSuccessResponse 登录成功响应
type LoginSuccessResponse struct {
	ReconnectKey string    `json:"reconnectKey"` // 重连密钥
	RoleInfo     *RoleInfo `json:"roleInfo"`     // 角色信息
}

// QuestData 任务数据
type QuestData struct {
	Quests []Quest `json:"quests"`
}

type Quest struct {
	QuestId  uint32 `json:"questId"`  // 任务Id
	Progress uint32 `json:"progress"` // 进度
	Status   uint32 `json:"status"`   // 状态 0=未接取 1=进行中 2=已完成 3=已领取
}

// LevelData 等级数据
type LevelData struct {
	Level uint32 `json:"level"` // 等级
	Exp   uint64 `json:"exp"`   // 经验
}

// BagData 背包数据
type BagData struct {
	Capacity uint32 `json:"capacity"` // 容量
	Items    []Item `json:"items"`    // 道具列表
}

// VipData VIP数据
type VipData struct {
	VipLevel uint32 `json:"vipLevel"` // VIP等级
	VipExp   uint64 `json:"vipExp"`   // VIP经验
}

// MoneyData 货币数据
type MoneyData struct {
	Gold    uint64 `json:"gold"`    // 金币
	Diamond uint64 `json:"diamond"` // 钻石
	Coin    uint64 `json:"coin"`    // 铜币
}

// AttrData 属性数据
type AttrData struct {
	HP      uint32 `json:"hp"`      // 生命值
	MP      uint32 `json:"mp"`      // 魔法值
	Attack  uint32 `json:"attack"`  // 攻击力
	Defense uint32 `json:"defense"` // 防御力
	Speed   uint32 `json:"speed"`   // 速度
}

// Mail 邮件
type Mail struct {
	MailId     uint64 `json:"mailId"`     // 邮件Id
	Title      string `json:"title"`      // 标题
	Content    string `json:"content"`    // 内容
	Sender     string `json:"sender"`     // 发送者
	SendTime   int64  `json:"sendTime"`   // 发送时间
	HasReward  bool   `json:"hasReward"`  // 是否有奖励
	IsRead     bool   `json:"isRead"`     // 是否已读
	IsReceived bool   `json:"isReceived"` // 是否已领取
	Rewards    []Item `json:"rewards"`    // 奖励列表
}

// MailListResponse 邮件列表响应
type MailListResponse struct {
	Mails []Mail `json:"mails"`
}

// MailDetailResponse 邮件详情响应
type MailDetailResponse struct {
	Mail Mail `json:"mail"`
}

// UnmarshalSelectRoleRequest 反序列化选择角色请求
func UnmarshalSelectRoleRequest(data []byte) (*SelectRoleRequest, error) {
	var req SelectRoleRequest
	err := tool.JsonUnmarshal(data, &req)
	return &req, err
}

// UnmarshalReconnectRequest 反序列化重连请求
func UnmarshalReconnectRequest(data []byte) (*ReconnectRequest, error) {
	var req ReconnectRequest
	err := tool.JsonUnmarshal(data, &req)
	return &req, err
}
