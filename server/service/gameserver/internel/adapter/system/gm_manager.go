package system

import (
	"context"
	"fmt"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"strconv"
	"strings"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
)

// GMCommandFunc GM命令函数类型
type GMCommandFunc func(playerRole iface.IPlayerRole, args ...string) bool

// GMCommandInfo GM命令信息
type GMCommandInfo struct {
	Name  string        // GM命令名称
	Level uint32        // 所需GM等级
	Func  GMCommandFunc // GM命令实现函数
	Desc  string        // 命令描述
	Usage string        // 使用说明
}

// GMManager GM命令管理器
type GMManager struct {
	commands map[string]*GMCommandInfo
}

func NewGMManager() *GMManager {
	mgr := &GMManager{
		commands: make(map[string]*GMCommandInfo),
	}
	mgr.registerDefaultCommands()
	return mgr
}

// RegisterCommand 注册GM命令
func (gm *GMManager) RegisterCommand(name string, level uint32, desc, usage string, fn GMCommandFunc) {
	gm.commands[name] = &GMCommandInfo{
		Name:  name,
		Level: level,
		Func:  fn,
		Desc:  desc,
		Usage: usage,
	}
}

// GetCommand 获取GM命令信息
func (gm *GMManager) GetCommand(name string) (*GMCommandInfo, bool) {
	cmd, ok := gm.commands[name]
	return cmd, ok
}

// ExecuteCommand 执行GM命令
func (gm *GMManager) ExecuteCommand(_ context.Context, playerRole iface.IPlayerRole, gmName string, args []string) (bool, string) {
	cmdInfo, ok := gm.GetCommand(gmName)
	if !ok {
		return false, fmt.Sprintf("GM命令不存在: %s", gmName)
	}

	playerGMLevel := playerRole.GetGMLevel()
	if playerGMLevel < cmdInfo.Level {
		return false, fmt.Sprintf("GM等级不足，需要等级 %d，当前等级 %d", cmdInfo.Level, playerGMLevel)
	}

	success := cmdInfo.Func(playerRole, args...)
	if success {
		return true, fmt.Sprintf("GM命令执行成功: %s", gmName)
	}
	return false, fmt.Sprintf("GM命令执行失败: %s，请检查参数: %s", gmName, cmdInfo.Usage)
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
		amount, err2 := strconv.ParseInt(args[1], 10, 10)
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
		newLevel, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil || newLevel < 1 || newLevel > 100 {
			return false
		}
		ctx := playerRole.WithContext(context.Background())
		levelSys := GetLevelSys(ctx)
		if levelSys == nil {
			return false
		}
		levelData, err := levelSys.GetLevelData(ctx)
		if err != nil || levelData == nil {
			return false
		}
		levelData.Level = uint32(newLevel)
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
			rewards = parseItemString(itemsStr)
		}
		if err := SendSystemMail(playerRole.GetPlayerRoleId(), title, content, rewards); err != nil {
			log.Errorf("SendSystemMail failed: %v", err)
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
		if err := SendSystemNotification(playerRole.GetPlayerRoleId(), title, content, notifType, priority); err != nil {
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
		if err := SendSystemNotificationToAll(title, content, notifType, priority); err != nil {
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
			rewards = parseItemString(itemsStr)
		}
		if err := SendSystemMailToAll(title, content, rewards); err != nil {
			log.Errorf("SendSystemMailToAll failed: %v", err)
			return false
		}
		return true
	})
}

// parseItemString 解析 GM 命令中的物品字符串，格式示例：
// "1001:10:1|1002:5:0" → []*jsonconf.ItemSt
func parseItemString(itemsStr string) []*jsonconf.ItemSt {
	if itemsStr == "" {
		return nil
	}
	parts := strings.Split(itemsStr, "|")
	result := make([]*jsonconf.ItemSt, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, ":")
		if len(fields) < 2 {
			continue
		}
		itemID, err1 := strconv.ParseUint(fields[0], 10, 32)
		count, err2 := strconv.ParseUint(fields[1], 10, 32)
		if len(fields) >= 3 {
			if b, err := strconv.ParseUint(fields[2], 10, 32); err == nil {
				_ = uint32(b) // 绑定信息当前未写入 ItemSt，占位以兼容旧格式
			}
		}
		if err1 != nil || err2 != nil || count == 0 {
			continue
		}
		result = append(result, &jsonconf.ItemSt{
			ItemId: uint32(itemID),
			Count:  uint32(count),
			Type:   uint32(0), // 由调用侧或配置决定类型；GM 命令通常只关心数量
		})
	}
	return result
}
