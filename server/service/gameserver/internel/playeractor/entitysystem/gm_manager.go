package entitysystem

import (
	"context"
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"strconv"
	"sync"
)

// GMCommandFunc GM命令函数类型
// 参数: playerRole - 玩家角色, args - 命令参数数组
// 返回: bool - 是否执行成功
type GMCommandFunc func(playerRole iface.IPlayerRole, args ...string) bool

// GMCommandInfo GM命令信息
type GMCommandInfo struct {
	Name  string        // GM命令名称
	Level uint32        // 所需GM等级: 0=普通玩家 1=GM 2=高级GM 3=超级GM
	Func  GMCommandFunc // GM命令实现函数
	Desc  string        // 命令描述
	Usage string        // 使用说明
}

// GMManager GM命令管理器
type GMManager struct {
	mu       sync.RWMutex
	commands map[string]*GMCommandInfo // GM命令映射表
}

var (
	globalGMManager *GMManager
	gmManagerOnce   sync.Once
)

// GetGMManager 获取全局GM管理器
func GetGMManager() *GMManager {
	gmManagerOnce.Do(func() {
		globalGMManager = &GMManager{
			commands: make(map[string]*GMCommandInfo),
		}
		// 注册默认GM命令
		globalGMManager.registerDefaultCommands()
	})
	return globalGMManager
}

// RegisterCommand 注册GM命令
func (gm *GMManager) RegisterCommand(name string, level uint32, desc, usage string, fn GMCommandFunc) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.commands[name] = &GMCommandInfo{
		Name:  name,
		Level: level,
		Func:  fn,
		Desc:  desc,
		Usage: usage,
	}

	log.Infof("GM command registered: %s (level=%d)", name, level)
}

// GetCommand 获取GM命令信息
func (gm *GMManager) GetCommand(name string) (*GMCommandInfo, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	cmd, ok := gm.commands[name]
	return cmd, ok
}

// ExecuteCommand 执行GM命令
func (gm *GMManager) ExecuteCommand(ctx context.Context, playerRole iface.IPlayerRole, gmName string, args []string) (bool, string) {
	// 获取GM命令信息
	cmdInfo, ok := gm.GetCommand(gmName)
	if !ok {
		return false, fmt.Sprintf("GM命令不存在: %s", gmName)
	}

	// 检查GM等级
	playerGMLevel := playerRole.GetGMLevel()
	if playerGMLevel < cmdInfo.Level {
		return false, fmt.Sprintf("GM等级不足，需要等级 %d，当前等级 %d", cmdInfo.Level, playerGMLevel)
	}

	// 执行GM命令
	success := cmdInfo.Func(playerRole, args...)
	if success {
		return true, fmt.Sprintf("GM命令执行成功: %s", gmName)
	} else {
		return false, fmt.Sprintf("GM命令执行失败: %s，请检查参数: %s", gmName, cmdInfo.Usage)
	}
}

// registerDefaultCommands 注册默认GM命令
func (gm *GMManager) registerDefaultCommands() {
	// 添加经验
	gm.RegisterCommand("addexp", 1, "添加经验", "addexp <经验值>", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 1 {
			return false
		}
		exp, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return false
		}
		ctx := playerRole.WithContext(context.Background())
		levelSys := GetLevelSys(ctx)
		if levelSys == nil {
			return false
		}
		if err := levelSys.AddExp(ctx, exp); err != nil {
			log.Errorf("AddExp failed: %v", err)
			return false
		}
		return true
	})

	// 添加金币
	gm.RegisterCommand("addmoney", 1, "添加金币", "addmoney <货币ID> <数量>", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 2 {
			return false
		}
		moneyID, err1 := strconv.ParseUint(args[0], 10, 32)
		amount, err2 := strconv.ParseInt(args[1], 10, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		ctx := playerRole.WithContext(context.Background())
		moneySys := GetMoneySys(ctx)
		if moneySys == nil {
			return false
		}
		if err := moneySys.AddMoney(ctx, uint32(moneyID), amount); err != nil {
			log.Errorf("AddMoney failed: %v", err)
			return false
		}
		return true
	})

	// 添加物品
	gm.RegisterCommand("additem", 1, "添加物品", "additem <物品ID> <数量> [绑定:0/1]", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 2 {
			return false
		}
		itemID, err1 := strconv.ParseUint(args[0], 10, 32)
		count, err2 := strconv.ParseUint(args[1], 10, 32)
		if err1 != nil || err2 != nil {
			return false
		}
		bind := uint32(0)
		if len(args) >= 3 {
			if b, err := strconv.ParseUint(args[2], 10, 32); err == nil {
				bind = uint32(b)
			}
		}
		ctx := playerRole.WithContext(context.Background())
		bagSys := GetBagSys(ctx)
		if bagSys == nil {
			return false
		}
		if err := bagSys.AddItem(ctx, uint32(itemID), uint32(count), bind); err != nil {
			log.Errorf("AddItem failed: %v", err)
			return false
		}
		return true
	})

	// 设置等级
	gm.RegisterCommand("setlevel", 2, "设置等级", "setlevel <等级>", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 1 {
			return false
		}
		level, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil || level < 1 || level > 100 {
			return false
		}
		ctx := playerRole.WithContext(context.Background())
		levelSys := GetLevelSys(ctx)
		if levelSys == nil {
			return false
		}
		levelData := levelSys.GetLevelData()
		if levelData == nil {
			return false
		}
		levelData.Level = uint32(level)
		levelData.Exp = 0
		return true
	})

	// 发送系统邮件
	gm.RegisterCommand("sendmail", 2, "发送系统邮件", "sendmail <标题> <内容> [物品ID:数量:绑定|...]", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 2 {
			return false
		}
		title := args[0]
		content := args[1]
		var rewards []*jsonconf.ItemSt
		if len(args) >= 3 {
			// 解析物品列表: 格式 "itemId:count:bind|itemId:count:bind|..."
			itemsStr := args[2]
			items := parseItemString(itemsStr)
			rewards = items
		}
		ctx := playerRole.WithContext(context.Background())
		mailSys := GetMailSys(ctx)
		if mailSys == nil {
			return false
		}
		if err := mailSys.SendCustomMail(ctx, title, content, rewards); err != nil {
			log.Errorf("SendCustomMail failed: %v", err)
			return false
		}
		return true
	})

	// 发送系统通知
	gm.RegisterCommand("sendnotice", 2, "发送系统通知", "sendnotice <标题> <内容> [类型:1-4] [优先级:1-3]", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 2 {
			return false
		}
		title := args[0]
		content := args[1]
		notifType := uint32(4) // 默认其他类型
		priority := uint32(2)  // 默认中等优先级
		if len(args) >= 3 {
			if t, err := strconv.ParseUint(args[2], 10, 32); err == nil {
				notifType = uint32(t)
			}
		}
		if len(args) >= 4 {
			if p, err := strconv.ParseUint(args[3], 10, 32); err == nil {
				priority = uint32(p)
			}
		}
		gmSys := GetGMSys()
		if err := gmSys.SendSystemNotification(playerRole.GetPlayerRoleId(), title, content, notifType, priority); err != nil {
			log.Errorf("SendSystemNotification failed: %v", err)
			return false
		}
		return true
	})

	// 全服发送系统通知
	gm.RegisterCommand("sendnoticeall", 3, "全服发送系统通知", "sendnoticeall <标题> <内容> [类型:1-4] [优先级:1-3]", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 2 {
			return false
		}
		title := args[0]
		content := args[1]
		notifType := uint32(4)
		priority := uint32(2)
		if len(args) >= 3 {
			if t, err := strconv.ParseUint(args[2], 10, 32); err == nil {
				notifType = uint32(t)
			}
		}
		if len(args) >= 4 {
			if p, err := strconv.ParseUint(args[3], 10, 32); err == nil {
				priority = uint32(p)
			}
		}
		gmSys := GetGMSys()
		if err := gmSys.SendSystemNotificationToAll(title, content, notifType, priority); err != nil {
			log.Errorf("SendSystemNotificationToAll failed: %v", err)
			return false
		}
		return true
	})

	// 全服发送系统邮件
	gm.RegisterCommand("sendmailall", 3, "全服发送系统邮件", "sendmailall <标题> <内容> [物品ID:数量:绑定|...]", func(playerRole iface.IPlayerRole, args ...string) bool {
		if len(args) < 2 {
			return false
		}
		title := args[0]
		content := args[1]
		var rewards []*jsonconf.ItemSt
		if len(args) >= 3 {
			itemsStr := args[2]
			items := parseItemString(itemsStr)
			rewards = items
		}
		gmSys := GetGMSys()
		if err := gmSys.SendSystemMailToAll(title, content, rewards); err != nil {
			log.Errorf("SendSystemMailToAll failed: %v", err)
			return false
		}
		return true
	})

	log.Infof("Default GM commands registered: %d commands", len(gm.commands))
}

// parseItemString 解析物品字符串 "itemId:count:bind|itemId:count:bind|..."
func parseItemString(itemsStr string) []*jsonconf.ItemSt {
	if itemsStr == "" {
		return nil
	}
	items := make([]*jsonconf.ItemSt, 0)
	parts := splitString(itemsStr, "|")
	for _, part := range parts {
		if part == "" {
			continue
		}
		itemParts := splitString(part, ":")
		if len(itemParts) < 2 {
			continue
		}
		itemID, err1 := strconv.ParseUint(itemParts[0], 10, 32)
		count, err2 := strconv.ParseUint(itemParts[1], 10, 32)
		if err1 != nil || err2 != nil {
			continue
		}
		// bind参数暂不使用，ItemSt结构体暂不支持bind字段
		items = append(items, &jsonconf.ItemSt{
			ItemId: uint32(itemID),
			Count:  uint32(count),
			Type:   1, // 默认普通物品类型
		})
	}
	return items
}

// splitString 分割字符串
func splitString(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	if len(result) == 0 {
		return []string{s}
	}
	return result
}
