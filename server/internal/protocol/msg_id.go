package protocol

// Client to Server 消息Id
const (
	C2S_Verify     uint16 = 1<<8 | 1 // 校验登陆
	C2S_QueryRoles uint16 = 1<<8 | 2 // 查询角色列表
	C2S_CreateRole uint16 = 1<<8 | 3 // 创建角色
	C2S_EnterGame  uint16 = 1<<8 | 4 // 进入游戏
	C2S_Reconnect  uint16 = 1<<8 | 5 // 重连请求
)

// Server to Client 消息Id
const (
	S2C_RoleList         uint16 = 1<<8 | 1     // 角色列表
	S2C_CreateRoleResult uint16 = 1<<8 | 2     // 创建角色结果
	S2C_EnterScene       uint16 = 1<<8 | 3     // 进入场景
	S2C_UpdateBagData    uint16 = 1<<8 | 4     // 更新背包数据
	S2C_QuestData        uint16 = 1<<8 | 5     // 任务数据
	S2C_LevelData        uint16 = 1<<8 | 6     // 等级数据
	S2C_BagData          uint16 = 1<<8 | 7     // 背包数据
	S2C_VipData          uint16 = 1<<8 | 8     // VIP数据
	S2C_MoneyData        uint16 = 1<<8 | 9     // 货币数据
	S2C_AttrData         uint16 = 1<<8 | 10    // 属性数据
	S2C_MailList         uint16 = 1<<8 | 11    // 邮件列表
	S2C_MailDetail       uint16 = 1<<8 | 12    // 邮件详情
	S2C_ReconnectKey     uint16 = 1<<8 | 13    // 重连密钥
	S2C_LoginSuccess     uint16 = 1<<8 | 14    // 登录成功
	S2C_Error            uint16 = 255<<8 | 255 // 错误消息
)

// DungeonServer RPC 消息Id
const (
	RPC_EnterDungeon  uint16 = 5001 // 进入副本
	RPC_LeaveDungeon  uint16 = 5002 // 离开副本
	RPC_MoveInDungeon uint16 = 5003 // 副本内移动
)
