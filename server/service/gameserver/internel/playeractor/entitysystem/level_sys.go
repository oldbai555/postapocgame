package entitysystem

import (
	"context"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem/attrcalc"
)

// LevelSys 等级系统
type LevelSys struct {
	*BaseSystem
	levelData *protocol.SiLevelData
}

// NewLevelSys 创建等级系统
func NewLevelSys() *LevelSys {
	sys := &LevelSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysLevel)),
	}
	return sys
}

func GetLevelSys(ctx context.Context) *LevelSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysLevel))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysLevel)
		return nil
	}
	sys := system.(*LevelSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v]", protocol.SystemId_SysLevel)
		return nil
	}
	return sys
}

// OnInit 系统初始化
func (ls *LevelSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("level sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果level_data不存在，则初始化
	if binaryData.LevelData == nil {
		binaryData.LevelData = &protocol.SiLevelData{
			Level: 1,
			Exp:   0,
		}
	}
	ls.levelData = binaryData.LevelData

	// 确保等级至少为1
	if ls.levelData.Level < 1 {
		ls.levelData.Level = 1
	}
	if ls.levelData.Exp < 0 {
		ls.levelData.Exp = 0
	}

	// 同步经验到货币系统（经验作为货币的一种）
	// 统一以等级系统的经验值为准，同步到货币系统
	if binaryData.MoneyData != nil {
		if binaryData.MoneyData.MoneyMap == nil {
			binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
		}
		expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
		// 统一以等级系统的经验值为准
		binaryData.MoneyData.MoneyMap[expMoneyID] = ls.levelData.Exp
	}
}

// GetLevelData 获取等级数据
func (ls *LevelSys) GetLevelData() *protocol.SiLevelData {
	return ls.levelData
}

// GetLevel 获取当前等级
func (ls *LevelSys) GetLevel() uint32 {
	if ls.levelData == nil {
		return 1
	}
	return ls.levelData.Level
}

// GetExp 获取当前经验
func (ls *LevelSys) GetExp() int64 {
	if ls.levelData == nil {
		return 0
	}
	return ls.levelData.Exp
}

// AddExp 添加经验值（经验作为货币的一种，存储在货币系统中，但由等级系统处理升级逻辑）
func (ls *LevelSys) AddExp(ctx context.Context, exp uint64) error {
	if exp == 0 {
		return nil
	}

	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if ls.levelData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level data not initialized")
	}

	// 经验同时存储在货币系统和等级系统中
	// 1. 更新货币系统中的经验值
	moneySys := GetMoneySys(ctx)
	if moneySys != nil {
		// 经验存储在货币系统中，使用MoneyTypeExp作为key
		expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
		currentExp := moneySys.GetAmount(expMoneyID)
		moneySys.moneyData.MoneyMap[expMoneyID] = currentExp + int64(exp)
	}

	// 2. 更新等级系统中的经验值（用于升级逻辑）
	ls.levelData.Exp += int64(exp)

	// 发布经验变化事件
	playerRole.Publish(gevent.OnPlayerExpChange, map[string]interface{}{
		"exp": ls.levelData.Exp,
	})

	// 检查是否升级
	if err := ls.CheckLevelUp(ctx); err != nil {
		return err
	}

	return nil
}

// CheckLevelUp 检查并处理升级逻辑
func (ls *LevelSys) CheckLevelUp(ctx context.Context) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if ls.levelData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level data not initialized")
	}

	// 循环检查升级（可能一次获得大量经验，连续升级）
	for {
		// 获取当前等级的配置
		levelConfig, ok := jsonconf.GetConfigManager().GetLevelConfig(ls.levelData.Level)
		if !ok {
			// 没有更高等级的配置，已达到最高等级
			break
		}

		// 检查是否满足升级条件
		if ls.levelData.Exp < int64(levelConfig.ExpNeeded) {
			break
		}

		// 扣除升级所需经验
		ls.levelData.Exp -= int64(levelConfig.ExpNeeded)

		// 同步更新货币系统中的经验值
		moneySys := GetMoneySys(ctx)
		if moneySys != nil {
			expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
			moneySys.moneyData.MoneyMap[expMoneyID] = ls.levelData.Exp
		}

		// 升级
		ls.levelData.Level++

		log.Infof("Player level up: PlayerID=%d, NewLevel=%d, RemainingExp=%d",
			playerRole.GetPlayerRoleId(), ls.levelData.Level, ls.levelData.Exp)

		// 发放升级奖励
		if len(levelConfig.Rewards) > 0 {
			rewards := make([]*jsonconf.ItemAmount, 0, len(levelConfig.Rewards))
			for _, reward := range levelConfig.Rewards {
				rewards = append(rewards, &jsonconf.ItemAmount{
					ItemType: uint32(reward.Type),
					ItemId:   reward.ItemId,
					Count:    int64(reward.Count),
					Bind:     1, // 升级奖励默认绑定
				})
			}
			if err := playerRole.GrantRewards(ctx, rewards); err != nil {
				log.Errorf("Grant level up rewards failed: %v", err)
				// 奖励发放失败不影响升级，只记录日志
			}
		}

		// 发布升级事件（包含level信息，供事件订阅者使用）
		playerRole.Publish(gevent.OnPlayerLevelUp, map[string]interface{}{
			"level": ls.levelData.Level,
		})
	}

	// 标记属性系统需要重算
	attrSys := GetAttrSys(ctx)
	if attrSys != nil {
		attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaLevel))
	}

	return nil
}

// CalculateAttrs 计算等级系统的属性（实现IAttrCalculator接口）
func (ls *LevelSys) CalculateAttrs(ctx context.Context) []*protocol.AttrSt {
	if ls.levelData == nil {
		return nil
	}

	// 从配置表获取等级属性
	levelAttrs := jsonconf.GetConfigManager().GetLevelAttrs(ls.levelData.Level)
	if len(levelAttrs) == 0 {
		return nil
	}

	// 转换为protocol.AttrSt格式
	result := make([]*protocol.AttrSt, 0, len(levelAttrs))
	for attrType, attrValue := range levelAttrs {
		result = append(result, &protocol.AttrSt{
			Type:  attrType,
			Value: int64(attrValue),
		})
	}

	return result
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSys()
	})
	gevent.SubscribePlayerEventH(gevent.OnPlayerLevelUp, func(ctx context.Context, ev *event.Event) {})
	gevent.SubscribePlayerEventH(gevent.OnPlayerExpChange, func(ctx context.Context, ev *event.Event) {})
	attrcalc.Register(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) attrcalc.Calculator {
		return GetLevelSys(ctx)
	})
	attrcalc.RegisterAddRate(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) attrcalc.AddRateCalculator {
		return &levelAddRateCalculator{}
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
	})
}

type levelAddRateCalculator struct {
	levelSys *LevelSys
}

func (c *levelAddRateCalculator) CalculateAddRate(ctx context.Context, _ *icalc.FightAttrCalc) []*protocol.AttrSt {
	if c.levelSys == nil {
		c.levelSys = GetLevelSys(ctx)
	}
	if c.levelSys == nil {
		return nil
	}
	level := c.levelSys.GetLevel()
	if level == 0 {
		return nil
	}
	cfg := jsonconf.GetConfigManager().GetAttrAddRateConfig()
	results := make([]*protocol.AttrSt, 0, 2)
	if cfg.Level.HPRegenPerLevel > 0 {
		regenHP := int64(level) * cfg.Level.HPRegenPerLevel
		if regenHP > 0 {
			results = append(results, &protocol.AttrSt{
				Type:  uint32(attrdef.AttrHPRegen),
				Value: regenHP,
			})
		}
	}
	if cfg.Level.MPRegenPerLevel > 0 {
		regenMP := int64(level) * cfg.Level.MPRegenPerLevel
		if regenMP > 0 {
			results = append(results, &protocol.AttrSt{
				Type:  uint32(attrdef.AttrMPRegen),
				Value: regenMP,
			})
		}
	}
	if len(results) == 0 {
		return nil
	}
	return results
}
