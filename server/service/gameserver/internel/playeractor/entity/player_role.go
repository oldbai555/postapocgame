package entity

import (
	"context"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"time"

	"google.golang.org/protobuf/proto"
)

// PlayerRole 玩家角色
type PlayerRole struct {
	// 基础信息
	SessionId  string
	SimpleData *protocol.PlayerSimpleData
	MainData   *protocol.PlayerRoleMainData
	BinaryData *protocol.PlayerRoleBinaryData

	// 重连相关
	ReconnectKey string
	IsOnline     bool
	DisconnectAt time.Time

	// Runtime 依赖聚合
	runtime *deps.Runtime

	// 系统管理器
	sysMgr       iface.ISystemMgr
	_1sChecker   *tool.TimeChecker
	_5minChecker *tool.TimeChecker
	timeCursor   timeCursorMark
}

type timeCursorMark struct {
	hour     int
	day      int
	month    int
	year     int
	week     int
	weekYear int
}

// NewPlayerRole 创建玩家角色
func NewPlayerRole(sessionId string, roleInfo *protocol.PlayerSimpleData) *PlayerRole {
	pr := &PlayerRole{
		SessionId:    sessionId,
		SimpleData:   roleInfo,
		IsOnline:     true,
		ReconnectKey: generateReconnectKey(sessionId, roleInfo.RoleId),
		// 创建 Runtime 实例（Phase 2D：聚合依赖，使用 deps 工厂函数）
		runtime: deps.NewRuntime(
			deps.NewPlayerGateway(),
			deps.NewRoleRepository(),
			deps.NewNetworkGateway(),
			deps.NewDungeonServerGateway(),
		),
	}
	// 创建系统管理器
	pr.sysMgr = entitysystem.NewSysMgr()

	// 从数据库加载BinaryData
	binaryData, err := database.GetPlayerBinaryData(uint(roleInfo.RoleId))
	if err != nil {
		log.Errorf("load player binary data failed: %v", err)
	}
	// 确保BinaryData不为nil
	if binaryData == nil {
		binaryData = &protocol.PlayerRoleBinaryData{}
	}
	if binaryData.SysOpenStatus == nil {
		binaryData.SysOpenStatus = make(map[uint32]uint32)
	}

	pr.BinaryData = binaryData

	mainData, err := database.GetPlayerMainData(uint(roleInfo.RoleId))
	if err != nil {
		log.Warnf("load player main data failed: %v", err)
		mainData = &protocol.PlayerRoleMainData{
			RoleId:   roleInfo.RoleId,
			Job:      roleInfo.Job,
			Sex:      roleInfo.Sex,
			Level:    roleInfo.Level,
			RoleName: roleInfo.RoleName,
			GmLevel:  roleInfo.GmLevel,
		}
	} else {
		mainData.RoleName = roleInfo.RoleName
		mainData.GmLevel = roleInfo.GmLevel
	}
	pr.MainData = mainData

	err = pr.sysMgr.OnInit(pr.WithContext(context.TODO()))
	if err != nil {
		log.Errorf("sys mgr on init failed, err:%v", err)
	}

	pr._1sChecker = tool.NewTimeChecker(time.Second)
	pr._5minChecker = tool.NewTimeChecker(5 * time.Minute)
	pr.timeCursor = newTimeCursorMark(servertime.Now())

	return pr
}

// OnLogin 登录回调
func (pr *PlayerRole) OnLogin() error {
	log.Infof("[PlayerRole] OnLogin: RoleId=%d, SessionId=%s", pr.SimpleData.RoleId, pr.SessionId)

	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	now := servertime.Now()
	pr.touchLoginTime(now)
	pr.timeCursor = newTimeCursorMark(now)
	pr.handleOfflineRollover(now)

	// 直接触发系统层登录处理
	pr.sysMgr.OnRoleLogin(pr.WithContext(context.TODO()))

	var resp protocol.S2CLoginRoleReq
	resp.ReconnectKey = pr.ReconnectKey
	resp.RoleData = pr.SimpleData

	return pr.SendProtoMessage(uint16(protocol.S2CProtocol_S2CLoginRole), &resp)
}

// OnLogout 登出回调
func (pr *PlayerRole) OnLogout() error {
	log.Infof("[PlayerRole] OnLogout: RoleId=%d", pr.SimpleData.RoleId)

	pr.IsOnline = false
	pr.touchLogoutTime(servertime.Now())

	// 保存BinaryData到数据库
	if pr.BinaryData != nil {
		if err := database.SavePlayerBinaryData(uint(pr.SimpleData.RoleId), pr.BinaryData); err != nil {
			log.Errorf("save player binary data failed: %v", err)
		}
	}

	return nil
}

// OnDisconnect 断线回调
func (pr *PlayerRole) OnDisconnect() {
	log.Infof("[PlayerRole] OnDisconnect: RoleId=%d", pr.SimpleData.RoleId)

	pr.IsOnline = false
	pr.DisconnectAt = servertime.Now()
}

// Close 关闭回调（3分钟超时或主动登出）
func (pr *PlayerRole) Close() error {
	log.Infof("[PlayerRole] Close: RoleId=%d", pr.SimpleData.RoleId)

	// 调用登出
	err := pr.OnLogout()
	if err != nil {
		log.Errorf("err:%v", err)
	}
	return nil
}

func (pr *PlayerRole) GetPlayerRoleId() uint64 {
	return pr.SimpleData.RoleId
}

func (pr *PlayerRole) GetSessionId() string {
	return pr.SessionId
}

// GetRuntime 获取 Runtime 实例（Phase 2D：供系统从 PlayerRole 获取依赖）
func (pr *PlayerRole) GetRuntime() *deps.Runtime {
	return pr.runtime
}

func (pr *PlayerRole) GetReconnectKey() string {
	return pr.ReconnectKey
}

func (pr *PlayerRole) GetGMLevel() uint32 {
	if pr.SimpleData == nil {
		return 0
	}
	return pr.SimpleData.GetGmLevel()
}

func (pr *PlayerRole) GetJob() uint32 {
	if pr.SimpleData == nil {
		return 0
	}
	return pr.SimpleData.Job
}

func (pr *PlayerRole) GetPlayerSimpleData() *protocol.PlayerSimpleData {
	return pr.SimpleData
}

func (pr *PlayerRole) GetSystem(sysId uint32) iface.ISystem {
	return pr.sysMgr.GetSystem(sysId)
}

func (pr *PlayerRole) SendMessage(protoId uint16, data []byte) error {
	return gatewaylink.SendToSession(pr.SessionId, protoId, data)
}

func (pr *PlayerRole) SendProtoMessage(protoId uint16, v proto.Message) error {
	bytes, err := proto.Marshal(v)
	if err != nil {
		return customerr.Wrap(err)
	}
	return pr.SendMessage(protoId, bytes)
}

func (pr *PlayerRole) WithContext(parentCtx context.Context) context.Context {
	var ctx = parentCtx
	if ctx == nil {
		ctx = context.Background()
	}
	// 注入 PlayerRole
	ctx = context.WithValue(ctx, gshare.ContextKeyRole, pr)
	// 注入 SessionId，便于日志与下游链路从 ctx 还原会话信息
	if pr.SessionId != "" {
		ctx = context.WithValue(ctx, gshare.ContextKeySession, pr.SessionId)
	}
	// 注入 Runtime，方便下游通过 deps.FromContext(ctx) 获取运行时依赖
	if pr.runtime != nil {
		ctx = pr.runtime.WithContext(ctx)
	}
	return ctx
}
func (pr *PlayerRole) GetSysStatus(sysId uint32) bool {
	idxInt := sysId / 32
	idxByte := sysId % 32

	flag := pr.GetBinaryData().SysOpenStatus[idxInt]

	return tool.IsSetBit(flag, idxByte)
}

func (pr *PlayerRole) GetSysStatusData() map[uint32]uint32 {
	return pr.GetBinaryData().SysOpenStatus
}

func (pr *PlayerRole) SetSysStatus(sysId uint32, isOpen bool) {
	idxInt := sysId / 32
	idxByte := sysId % 32

	binary := pr.GetBinaryData()
	if isOpen {
		binary.SysOpenStatus[idxInt] = tool.SetBit(binary.SysOpenStatus[idxInt], idxByte)
	} else {
		binary.SysOpenStatus[idxInt] = tool.ClearBit(binary.SysOpenStatus[idxInt], idxByte)
	}
}

func (pr *PlayerRole) GetSysMgr() iface.ISystemMgr {
	return pr.sysMgr
}

func (pr *PlayerRole) CallDungeonActor(ctx context.Context, msgId uint16, data []byte) error {
	msgCtx := pr.WithContext(ctx)
	if pr.runtime == nil {
		return customerr.NewError("runtime is nil")
	}
	return pr.runtime.DungeonGateway().AsyncCall(msgCtx, pr.GetSessionId(), msgId, data)
}

func (pr *PlayerRole) timeSync() {
	resp := &protocol.S2CTimeSyncReq{
		ServerTimeMs: servertime.UnixMilli(),
	}
	err := pr.SendProtoMessage(uint16(protocol.S2CProtocol_S2CTimeSync), resp)
	if err != nil {
		log.Errorf("send time sync failed, err: %v", err)
	}
}

// RunOne 每帧调用，处理属性增量更新等
func (pr *PlayerRole) RunOne() {
	if !pr.IsOnline {
		return
	}

	pr.handleTimeEvents()

	if pr._1sChecker.CheckAndSet(true) {
		pr.timeSync()
	}

	if pr._5minChecker.CheckAndSet(true) {
		err := pr.SaveToDB()
		if err != nil {
			log.Errorf("save player binary data failed: %v", err)
		}
	}

}

func (pr *PlayerRole) OnNewHour(ctx context.Context) {
	if ctx == nil {
		ctx = pr.WithContext(context.TODO())
	}
	pr.sysMgr.OnNewHour(ctx)
}

func (pr *PlayerRole) OnNewDay(ctx context.Context) {
	if ctx == nil {
		ctx = pr.WithContext(context.TODO())
	}
	pr.sysMgr.OnNewDay(ctx)
}

func (pr *PlayerRole) OnNewWeek(ctx context.Context) {
	if ctx == nil {
		ctx = pr.WithContext(context.TODO())
	}
	pr.sysMgr.OnNewWeek(ctx)
}

func (pr *PlayerRole) OnNewMonth(ctx context.Context) {
	if ctx == nil {
		ctx = pr.WithContext(context.TODO())
	}
	pr.sysMgr.OnNewMonth(ctx)
}

func (pr *PlayerRole) OnNewYear(ctx context.Context) {
	if ctx == nil {
		ctx = pr.WithContext(context.TODO())
	}
	pr.sysMgr.OnNewYear(ctx)
}

func (pr *PlayerRole) handleTimeEvents() {
	if !pr.IsOnline {
		return
	}
	if pr.timeCursor.isZero() {
		pr.timeCursor = newTimeCursorMark(servertime.Now())
		return
	}

	now := servertime.Now().In(time.Local)
	ctx := pr.WithContext(context.TODO())

	if pr.timeCursor.year != now.Year() {
		pr.timeCursor.year = now.Year()
		pr.OnNewYear(ctx)
	}
	if pr.timeCursor.month != int(now.Month()) {
		pr.timeCursor.month = int(now.Month())
		pr.OnNewMonth(ctx)
	}
	isoYear, week := now.ISOWeek()
	if pr.timeCursor.weekYear != isoYear || pr.timeCursor.week != week {
		pr.timeCursor.weekYear = isoYear
		pr.timeCursor.week = week
		pr.OnNewWeek(ctx)
	}
	if pr.timeCursor.day != now.YearDay() {
		pr.timeCursor.day = now.YearDay()
		pr.OnNewDay(ctx)
	}
	if pr.timeCursor.hour != now.Hour() {
		pr.timeCursor.hour = now.Hour()
		pr.OnNewHour(ctx)
	}
}

// SaveToDB 立即将玩家的数据存储到Player角色表
func (pr *PlayerRole) SaveToDB() error {
	if pr.BinaryData == nil {
		return nil
	}
	if err := database.SavePlayerBinaryData(uint(pr.SimpleData.RoleId), pr.BinaryData); err != nil {
		log.Errorf("save player binary data failed: %v", err)
		return err
	}
	log.Infof("PlayerRole SaveToDB success: RoleId=%d", pr.SimpleData.RoleId)
	return nil
}

func newTimeCursorMark(t time.Time) timeCursorMark {
	isoYear, week := t.ISOWeek()
	return timeCursorMark{
		hour:     t.Hour(),
		day:      t.YearDay(),
		month:    int(t.Month()),
		year:     t.Year(),
		week:     week,
		weekYear: isoYear,
	}
}

func (tc timeCursorMark) isZero() bool {
	return tc.year == 0
}

func (pr *PlayerRole) touchLoginTime(now time.Time) {
	if pr.MainData != nil {
		pr.MainData.LastLoginTime = now.Unix()
	}
	if err := database.UpdatePlayerLoginTime(uint(pr.SimpleData.RoleId), now); err != nil {
		log.Warnf("update login time failed: %v", err)
	}
}

func (pr *PlayerRole) touchLogoutTime(now time.Time) {
	if pr.MainData != nil {
		pr.MainData.LastLogoutTime = now.Unix()
	}
	if err := database.UpdatePlayerLogoutTime(uint(pr.SimpleData.RoleId), now); err != nil {
		log.Warnf("update logout time failed: %v", err)
	}
}

func (pr *PlayerRole) handleOfflineRollover(now time.Time) {
	if pr.MainData == nil || pr.MainData.LastLogoutTime == 0 {
		return
	}
	last := time.Unix(pr.MainData.LastLogoutTime, 0).In(time.Local)
	now = now.In(time.Local)
	ctx := pr.WithContext(context.TODO())

	if last.Year() != now.Year() {
		pr.OnNewYear(ctx)
	}
	if last.Year() != now.Year() || last.Month() != now.Month() {
		pr.OnNewMonth(ctx)
	}
	lastIsoYear, lastWeek := last.ISOWeek()
	nowIsoYear, nowWeek := now.ISOWeek()
	if lastIsoYear != nowIsoYear || lastWeek != nowWeek {
		pr.OnNewWeek(ctx)
	}
	if last.YearDay() != now.YearDay() || last.Year() != now.Year() {
		pr.OnNewDay(ctx)
	}
}
