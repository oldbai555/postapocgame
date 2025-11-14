package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

// FubenSys 副本系统
type FubenSys struct {
	*BaseSystem
	dungeonData *protocol.SiDungeonData
}

// NewFubenSys 创建副本系统
func NewFubenSys() *FubenSys {
	return &FubenSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysDungeon)),
	}
}

func GetFubenSys(ctx context.Context) *FubenSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysDungeon))
	if system == nil {
		return nil
	}
	fubenSys, ok := system.(*FubenSys)
	if !ok || !fubenSys.IsOpened() {
		return nil
	}
	return fubenSys
}

// OnInit 初始化时从PlayerRoleBinaryData加载副本数据
func (fs *FubenSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("fuben sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果dungeon_data不存在，则初始化
	if binaryData.DungeonData == nil {
		binaryData.DungeonData = &protocol.SiDungeonData{
			Records: make([]*protocol.DungeonRecord, 0),
		}
	}
	fs.dungeonData = binaryData.DungeonData

	// 如果Records为空，初始化为空切片
	if fs.dungeonData.Records == nil {
		fs.dungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	log.Infof("FubenSys initialized: RecordCount=%d", len(fs.dungeonData.Records))
}

// GetDungeonRecord 获取副本记录
func (fs *FubenSys) GetDungeonRecord(dungeonID uint32, difficulty uint32) *protocol.DungeonRecord {
	if fs.dungeonData == nil || fs.dungeonData.Records == nil {
		return nil
	}
	for _, record := range fs.dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record
		}
	}
	return nil
}

// GetOrCreateDungeonRecord 获取或创建副本记录
func (fs *FubenSys) GetOrCreateDungeonRecord(dungeonID uint32, difficulty uint32) *protocol.DungeonRecord {
	if fs.dungeonData == nil {
		return nil
	}
	if fs.dungeonData.Records == nil {
		fs.dungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	// 查找现有记录
	for _, record := range fs.dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record
		}
	}

	// 创建新记录
	now := time.Now().Unix()
	newRecord := &protocol.DungeonRecord{
		DungeonId:     dungeonID,
		Difficulty:    difficulty,
		LastEnterTime: now,
		EnterCount:    0,
		ResetTime:     now,
	}
	fs.dungeonData.Records = append(fs.dungeonData.Records, newRecord)
	return newRecord
}

// CheckDungeonCD 检查副本CD（冷却时间）
func (fs *FubenSys) CheckDungeonCD(dungeonID uint32, difficulty uint32, cdMinutes uint32) (bool, time.Duration) {
	record := fs.GetDungeonRecord(dungeonID, difficulty)
	if record == nil {
		// 没有记录，可以进入
		return true, 0
	}

	now := time.Now()
	lastEnterTime := time.Unix(record.LastEnterTime, 0)
	elapsed := now.Sub(lastEnterTime)
	cdDuration := time.Duration(cdMinutes) * time.Minute

	if elapsed < cdDuration {
		// 还在CD中
		remaining := cdDuration - elapsed
		return false, remaining
	}

	return true, 0
}

// EnterDungeon 进入副本（更新进入时间和次数）
func (fs *FubenSys) EnterDungeon(dungeonID uint32, difficulty uint32) error {
	if fs.dungeonData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "dungeon data not initialized")
	}

	record := fs.GetOrCreateDungeonRecord(dungeonID, difficulty)
	if record == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "failed to create dungeon record")
	}

	now := time.Now()
	lastResetTime := time.Unix(record.ResetTime, 0)

	// 检查是否需要重置（每日重置）
	if now.Sub(lastResetTime) >= 24*time.Hour {
		record.EnterCount = 1
		record.ResetTime = now.Unix()
	} else {
		record.EnterCount++
	}
	record.LastEnterTime = now.Unix()

	return nil
}

// GetDungeonData 获取副本数据（用于协议）
func (fs *FubenSys) GetDungeonData() *protocol.SiDungeonData {
	return fs.dungeonData
}

// GetAllRecords 获取所有副本记录
func (fs *FubenSys) GetAllRecords() []*protocol.DungeonRecord {
	if fs.dungeonData == nil || fs.dungeonData.Records == nil {
		return make([]*protocol.DungeonRecord, 0)
	}
	return fs.dungeonData.Records
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysDungeon), func() iface.ISystem {
		return NewFubenSys()
	})
}
