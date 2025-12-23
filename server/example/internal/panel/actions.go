package panel

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func (p *AdventurePanel) exec(line string) error {
	p.recordCommand(line)
	fields := splitFields(line)
	if len(fields) == 0 {
		return nil
	}
	cmd := strings.ToLower(fields[0])

	if action := p.findMenuAction(cmd); action != nil {
		return action()
	}

	switch cmd {
	case "help", "?":
		p.printHelp()
	case "connect":
		addr := ""
		if len(fields) > 1 {
			addr = fields[1]
		}
		return p.connect(addr)
	case "disconnect":
		return p.disconnect()
	case "register":
		if len(fields) < 3 {
			return fmt.Errorf("用法: register <account> <password>")
		}
		if _, err := p.requireConnection(); err != nil {
			return err
		}
		if err := p.systems.Account.Register(fields[1], fields[2]); err != nil {
			return err
		}
		p.appendLog("✅ 注册成功: %s", fields[1])
	case "login":
		if len(fields) < 3 {
			return fmt.Errorf("用法: login <account> <password>")
		}
		if _, err := p.requireConnection(); err != nil {
			return err
		}
		if err := p.systems.Account.Login(fields[1], fields[2]); err != nil {
			return err
		}
		p.loggedIn = true
		p.inScene = false
		p.activeRole = 0
		p.appendLog("✅ 登录成功: %s", fields[1])
	case "roles":
		if _, err := p.requireLogin(); err != nil {
			return err
		}
		roles, err := p.systems.Account.ListRoles()
		if err != nil {
			return err
		}
		if len(roles) == 0 {
			p.appendLog("📭 暂无角色，可使用 create-role 创建")
			return nil
		}
		p.appendLog("📜 角色列表：")
		for _, role := range roles {
			p.appendLog("  • ID=%d 名称=%s 职业=%d 等级=%d", role.RoleId, role.RoleName, role.Job, role.Level)
		}
	case "create-role":
		if len(fields) < 2 {
			return fmt.Errorf("用法: create-role <name> [job] [sex]")
		}
		if _, err := p.requireLogin(); err != nil {
			return err
		}
		job := parseUintDefault(fields, 2, 1)
		sex := parseUintDefault(fields, 3, 1)
		if err := p.systems.Account.CreateRole(fields[1], job, sex); err != nil {
			return err
		}
		p.appendLog("✨ 已提交角色创建：%s", fields[1])
	case "enter":
		if len(fields) < 2 {
			return fmt.Errorf("用法: enter <roleID>")
		}
		if _, err := p.requireLogin(); err != nil {
			return err
		}
		roleID, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return fmt.Errorf("roleID 无法解析: %w", err)
		}
		if err := p.systems.Account.EnterRole(roleID); err != nil {
			return err
		}
		p.inScene = p.core.HasEnteredScene()
		p.activeRole = roleID
		p.appendLog("🚪 已进入游戏，角色ID=%d", roleID)
	case "status":
		if _, err := p.requireScene(); err != nil {
			return err
		}
		p.logStatus()
	case "look":
		if _, err := p.requireScene(); err != nil {
			return err
		}
		p.logLook()
	case "move":
		if len(fields) < 3 {
			return fmt.Errorf("用法: move <dx> <dy>")
		}
		if _, err := p.requireScene(); err != nil {
			return err
		}
		dx, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("dx 不是有效数字: %w", err)
		}
		dy, err := strconv.Atoi(fields[2])
		if err != nil {
			return fmt.Errorf("dy 不是有效数字: %w", err)
		}
		if err := p.systems.Move.MoveDelta(p.ctx, int32(dx), int32(dy), nil); err != nil {
			return err
		}
		p.appendLog("🚶 提交移动 Δ(%d,%d)", dx, dy)
	case "move-to":
		if len(fields) < 3 {
			return fmt.Errorf("用法: move-to <x> <y>")
		}
		if _, err := p.requireScene(); err != nil {
			return err
		}
		xVal, err := strconv.ParseUint(fields[1], 10, 32)
		if err != nil {
			return fmt.Errorf("x 不是有效数字: %w", err)
		}
		yVal, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			return fmt.Errorf("y 不是有效数字: %w", err)
		}
		if err := p.systems.Move.MoveTo(p.ctx, uint32(xVal), uint32(yVal), nil); err != nil {
			return err
		}
		p.appendLog("🚶 自动寻路至 (%d,%d)", xVal, yVal)
	case "move-resume":
		if _, err := p.requireScene(); err != nil {
			return err
		}
		if err := p.systems.Move.Resume(p.ctx, nil); err != nil {
			return err
		}
		p.appendLog("🔁 继续上一次自动移动")
	case "attack":
		if len(fields) < 2 {
			return fmt.Errorf("用法: attack <entityHandle>")
		}
		if _, err := p.requireScene(); err != nil {
			return err
		}
		target, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return fmt.Errorf("entityHandle 解析失败: %w", err)
		}
		if err := p.systems.Combat.NormalAttack(target, 3*time.Second); err != nil {
			return err
		}
		p.appendLog("⚔️ 发起普通攻击 Handle=%d", target)
	case "quit", "exit":
		return errPanelExit
	default:
		p.appendLog("未知命令: %s (输入 help 查看帮助)", cmd)
	}
	return nil
}

func (p *AdventurePanel) printHelp() {
	fmt.Println("\n可用命令：")
	fmt.Println("  help                     显示本帮助")
	fmt.Println("  connect [addr]           连接 Gateway (默认 0.0.0.0:1011)")
	fmt.Println("  disconnect               断开当前连接")
	fmt.Println("  register <acc> <pwd>     注册账号")
	fmt.Println("  login <acc> <pwd>        登录账号")
	fmt.Println("  roles                    查看账号下的角色列表")
	fmt.Println("  create-role <name> ...   创建角色（可指定 job 和 sex）")
	fmt.Println("  enter <roleID>           进入指定角色")
	fmt.Println("  status                   查看当前角色状态")
	fmt.Println("  look                     观察 AOI 内实体")
	fmt.Println("  move <dx> <dy>           按位移提交移动指令")
	fmt.Println("  move-to <x> <y>          自动寻路到指定格子")
	fmt.Println("  move-resume              继续上一次自动寻路")
	fmt.Println("  attack <handle>          对指定实体发起普通攻击")
	fmt.Println("  quit / exit              退出客户端")
}

func (p *AdventurePanel) actionRegister() error {
	if _, err := p.requireConnection(); err != nil {
		return err
	}
	account, err := p.promptInput("账号")
	if err != nil {
		return err
	}
	password, err := p.promptInput("密码")
	if err != nil {
		return err
	}
	if err := p.systems.Account.Register(account, password); err != nil {
		return err
	}
	p.appendLog("✅ 注册成功: %s", account)
	return nil
}

func (p *AdventurePanel) actionLogin() error {
	if _, err := p.requireConnection(); err != nil {
		return err
	}
	account, err := p.promptInput("账号")
	if err != nil {
		return err
	}
	password, err := p.promptInput("密码")
	if err != nil {
		return err
	}
	if err := p.systems.Account.Login(account, password); err != nil {
		return err
	}
	p.loggedIn = true
	p.inScene = false
	p.activeRole = 0
	p.appendLog("✅ 登录成功: %s", account)
	return nil
}

func (p *AdventurePanel) actionListRoles() error {
	if _, err := p.requireLogin(); err != nil {
		return err
	}
	roles, err := p.systems.Account.ListRoles()
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		p.appendLog("📭 暂无角色，可创建新角色")
		return nil
	}
	p.appendLog("📜 角色列表：")
	for _, role := range roles {
		p.appendLog("  • ID=%d 名称=%s 职业=%d 等级=%d", role.RoleId, role.RoleName, role.Job, role.Level)
	}
	return nil
}

func (p *AdventurePanel) actionCreateRole() error {
	if _, err := p.requireLogin(); err != nil {
		return err
	}
	name, err := p.promptInput("角色名")
	if err != nil {
		return err
	}
	jobStr, err := p.promptInput("职业 (默认1)")
	if err != nil {
		return err
	}
	sexStr, err := p.promptInput("性别 (默认1)")
	if err != nil {
		return err
	}
	job := parseInputUintDefault(jobStr, 1)
	sex := parseInputUintDefault(sexStr, 1)
	if err := p.systems.Account.CreateRole(name, job, sex); err != nil {
		return err
	}
	p.appendLog("✨ 创建角色：%s", name)
	return nil
}

func (p *AdventurePanel) actionEnterRole() error {
	if _, err := p.requireLogin(); err != nil {
		return err
	}
	roleStr, err := p.promptInput("角色ID")
	if err != nil {
		return err
	}
	roleID, err := strconv.ParseUint(roleStr, 10, 64)
	if err != nil {
		return err
	}
	if err := p.systems.Account.EnterRole(roleID); err != nil {
		return err
	}
	p.inScene = p.core.HasEnteredScene()
	p.activeRole = roleID
	p.appendLog("🚪 已进入角色 %d", roleID)
	return nil
}

func (p *AdventurePanel) actionLogout() error {
	p.loggedIn = false
	p.inScene = false
	p.activeRole = 0
	p.appendLog("👤 已注销账号")
	return nil
}

func (p *AdventurePanel) actionMovePrompt() error {
	if _, err := p.requireScene(); err != nil {
		return err
	}
	dxStr, err := p.promptInput("ΔX")
	if err != nil {
		return err
	}
	dyStr, err := p.promptInput("ΔY")
	if err != nil {
		return err
	}
	dx, err := strconv.Atoi(dxStr)
	if err != nil {
		return err
	}
	dy, err := strconv.Atoi(dyStr)
	if err != nil {
		return err
	}
	if err := p.systems.Move.MoveDelta(p.ctx, int32(dx), int32(dy), nil); err != nil {
		return err
	}
	p.appendLog("🚶 移动 Δ(%d,%d)", dx, dy)
	return nil
}

func (p *AdventurePanel) actionAttackPrompt() error {
	if _, err := p.requireScene(); err != nil {
		return err
	}
	targetStr, err := p.promptInput("目标 Handle")
	if err != nil {
		return err
	}
	target, err := strconv.ParseUint(targetStr, 10, 64)
	if err != nil {
		return err
	}
	if err := p.systems.Combat.NormalAttack(target, 3*time.Second); err != nil {
		return err
	}
	p.appendLog("⚔️ 攻击目标 %d", target)
	return nil
}

func (p *AdventurePanel) actionMoveToPrompt() error {
	if _, err := p.requireScene(); err != nil {
		return err
	}
	xStr, err := p.promptInput("目标 X 坐标")
	if err != nil {
		return err
	}
	yStr, err := p.promptInput("目标 Y 坐标")
	if err != nil {
		return err
	}
	xVal, err := strconv.ParseUint(xStr, 10, 32)
	if err != nil {
		return err
	}
	yVal, err := strconv.ParseUint(yStr, 10, 32)
	if err != nil {
		return err
	}
	if err := p.systems.Move.MoveTo(p.ctx, uint32(xVal), uint32(yVal), nil); err != nil {
		return err
	}
	p.appendLog("🚶 自动寻路至 (%d,%d)", xVal, yVal)
	return nil
}

func splitFields(line string) []string {
	return strings.Fields(line)
}

func parseUintDefault(fields []string, idx int, def uint32) uint32 {
	if len(fields) <= idx {
		return def
	}
	if v, err := strconv.ParseUint(fields[idx], 10, 32); err == nil {
		return uint32(v)
	}
	return def
}

func parseInputUintDefault(input string, def uint32) uint32 {
	if input == "" {
		return def
	}
	if v, err := strconv.ParseUint(input, 10, 32); err == nil {
		return uint32(v)
	}
	return def
}

func (p *AdventurePanel) startRecording(path string) error {
	if p.scriptRecorder != nil {
		_ = p.scriptRecorder.Close()
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	p.scriptRecorder = file
	p.appendLog("📝 开始录制命令到 %s", path)
	return nil
}

func (p *AdventurePanel) stopRecording() {
	if p.scriptRecorder == nil {
		return
	}
	_ = p.scriptRecorder.Close()
	p.scriptRecorder = nil
	p.appendLog("🛑 命令录制结束")
}

func (p *AdventurePanel) recordCommand(line string) {
	if p.scriptRecorder == nil || p.suppressRecord > 0 {
		return
	}
	if _, err := fmt.Fprintln(p.scriptRecorder, line); err != nil {
		p.appendLog("⚠️ 写入脚本失败: %v", err)
	}
}

func (p *AdventurePanel) runScript(path string, delay time.Duration) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	p.suppressRecord++
	defer func() { p.suppressRecord-- }()

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p.appendLog("📜[%d] %s", lineNum, line)
		if err := p.exec(line); err != nil {
			return fmt.Errorf("脚本 %s 第 %d 行执行失败: %w", path, lineNum, err)
		}
		if delay > 0 {
			select {
			case <-p.ctx.Done():
				return p.ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	p.appendLog("✅ 脚本执行完成: %s", path)
	return nil
}

func parseDefault(fields []string, idx int, def string) string {
	if len(fields) <= idx {
		return def
	}
	return fields[idx]
}

func attrValueOrZero(attrs map[uint32]int64, attrType uint32) int64 {
	if attrs == nil {
		return 0
	}
	if val, ok := attrs[attrType]; ok {
		return val
	}
	return 0
}
