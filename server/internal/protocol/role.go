package protocol

// RoleInfo 角色信息
type RoleInfo struct {
	RoleId uint64 `json:"roleId"` // 角色Id
	Job    uint32 `json:"job"`    // 职业
	Sex    uint32 `json:"sex"`    // 性别: 0=女, 1=男
	Name   string `json:"name"`   // 名字
	Level  uint32 `json:"level"`  // 等级
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles []*RoleInfo `json:"roles"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Job  uint32 `json:"job"`  // 职业
	Sex  uint32 `json:"sex"`  // 性别
	Name string `json:"name"` // 名字
}

// CreateRoleResponse 创建角色响应
type CreateRoleResponse struct {
	Success bool      `json:"success"`
	Role    *RoleInfo `json:"playerrole,omitempty"`
	ErrMsg  string    `json:"errMsg,omitempty"`
}

// EnterSceneResponse 进入场景响应
type EnterSceneResponse struct {
	SceneId  uint32    `json:"sceneId"`  // 场景Id
	RoleInfo *RoleInfo `json:"roleInfo"` // 角色信息
	PosX     float32   `json:"posX"`     // X坐标
	PosY     float32   `json:"posY"`     // Y坐标
}

// MoveRequest 移动请求
type MoveRequest struct {
	TargetX float32 `json:"targetX"` // 目标X坐标
	TargetY float32 `json:"targetY"` // 目标Y坐标
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code   int32  `json:"code"`
	ErrMsg string `json:"errMsg"`
}
