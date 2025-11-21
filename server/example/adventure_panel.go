package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"strconv"
	"strings"
	"time"
)

var errPanelExit = errors.New("panel exit")

// AdventurePanel 文字冒险式交互面板
type AdventurePanel struct {
	ctx         context.Context
	manager     *ClientManager
	client      *GameClient
	gatewayAddr string
	loggedIn    bool
	inScene     bool
	activeRole  uint64
	logs        []string
	logLimit    int
	input       *bufio.Reader
}

func NewAdventurePanel(ctx context.Context, mgr *ClientManager) *AdventurePanel {
	return &AdventurePanel{
		ctx:         ctx,
		manager:     mgr,
		gatewayAddr: GatewayAddr,
		logLimit:    8,
	}
}

func (p *AdventurePanel) Run() error {
	p.input = bufio.NewReader(os.Stdin)
	p.appendLog("欢迎来到废土，请选择操作")
	p.tryAutoConnect()

	for {
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		default:
		}

		p.renderFrame()

		line, err := p.readLine()
		if err != nil {
			return err
		}
		if line == "" {
			continue
		}
		if err := p.exec(line); err != nil {
			if errors.Is(err, errPanelExit) {
				_ = p.disconnect()
				return nil
			}
			fmt.Printf("⚠️ %v\n", err)
		}
	}
}

func (p *AdventurePanel) exec(line string) error {
	fields := strings.Fields(line)
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
		client, err := p.requireClient()
		if err != nil {
			return err
		}
		if err := client.RegisterAccount(fields[1], fields[2]); err != nil {
			return err
		}
		p.appendLog("✅ 注册成功: %s", fields[1])
	case "login":
		if len(fields) < 3 {
			return fmt.Errorf("用法: login <account> <password>")
		}
		client, err := p.requireClient()
		if err != nil {
			return err
		}
		if err := client.LoginAccount(fields[1], fields[2]); err != nil {
			return err
		}
		p.loggedIn = true
		p.inScene = false
		p.activeRole = 0
		p.appendLog("✅ 登录成功: %s", fields[1])
	case "roles":
		client, err := p.requireLogin()
		if err != nil {
			return err
		}
		roles, err := client.ListRoles()
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
		client, err := p.requireLogin()
		if err != nil {
			return err
		}
		job := uint32(1)
		sex := uint32(1)
		if len(fields) > 2 {
			if v, err := strconv.ParseUint(fields[2], 10, 32); err == nil {
				job = uint32(v)
			}
		}
		if len(fields) > 3 {
			if v, err := strconv.ParseUint(fields[3], 10, 32); err == nil {
				sex = uint32(v)
			}
		}
		if err := client.CreateRole(fields[1], job, sex); err != nil {
			return err
		}
		p.appendLog("✨ 已提交角色创建：%s", fields[1])
	case "enter":
		if len(fields) < 2 {
			return fmt.Errorf("用法: enter <roleID>")
		}
		client, err := p.requireLogin()
		if err != nil {
			return err
		}
		roleID, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return fmt.Errorf("roleID 无法解析: %w", err)
		}
		if err := client.EnterGame(roleID); err != nil {
			return err
		}
		p.inScene = client.HasEnteredScene()
		p.activeRole = roleID
		p.appendLog("🚪 已进入游戏，角色ID=%d", roleID)
	case "status":
		client, err := p.requireScene()
		if err != nil {
			return err
		}
		p.logStatus(client)
	case "look":
		client, err := p.requireScene()
		if err != nil {
			return err
		}
		p.logLook(client)
	case "move":
		if len(fields) < 3 {
			return fmt.Errorf("用法: move <dx> <dy>")
		}
		client, err := p.requireScene()
		if err != nil {
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
		if err := client.NudgeMove(int32(dx), int32(dy)); err != nil {
			return err
		}
		p.appendLog("🚶 提交移动 Δ(%d,%d)", dx, dy)
	case "attack":
		if len(fields) < 2 {
			return fmt.Errorf("用法: attack <entityHandle>")
		}
		client, err := p.requireScene()
		if err != nil {
			return err
		}
		target, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return fmt.Errorf("entityHandle 解析失败: %w", err)
		}
		if err := client.CastNormalAttack(target); err != nil {
			return err
		}
		p.appendLog("⚔️ 发起普通攻击 Handle=%d", target)
		if hit, err := client.WaitForSkillResult(target, 3*time.Second); err == nil {
			p.appendLog("   ➤ 伤害=%d HP=%d MP=%d State=%d",
				hit.Damage,
				attrValueOrZero(hit.Attrs, attrdef.AttrHP),
				attrValueOrZero(hit.Attrs, attrdef.AttrMP),
				hit.StateFlags,
			)
		} else {
			p.appendLog("   ⚠️ 等待技能结果失败: %v", err)
		}
	case "quit", "exit":
		return errPanelExit
	default:
		p.appendLog("未知命令: %s (输入 help 查看帮助)", cmd)
	}
	return nil
}

func (p *AdventurePanel) connect(addr string) error {
	if addr == "" {
		addr = p.gatewayAddr
	}
	if addr == "" {
		addr = GatewayAddr
	}
	if addr == "" {
		return fmt.Errorf("未配置 Gateway 地址")
	}
	if p.client != nil {
		_ = p.disconnect()
	}
	clientID := fmt.Sprintf("panel-%s", tool.GenUUID())
	client := p.manager.CreateClient(clientID, addr)
	if err := client.Start(p.ctx); err != nil {
		p.manager.DestroyClient(clientID)
		return err
	}
	p.client = client
	p.gatewayAddr = addr
	p.loggedIn = false
	p.inScene = false
	p.activeRole = 0
	p.appendLog("🔗 已连接 Gateway %s", addr)
	return nil
}

func (p *AdventurePanel) disconnect() error {
	if p.client == nil {
		return nil
	}
	playerID := p.client.GetPlayerID()
	p.manager.DestroyClient(playerID)
	p.client = nil
	p.loggedIn = false
	p.inScene = false
	p.activeRole = 0
	p.appendLog("🔌 已断开 Gateway 连接")
	return nil
}

func (p *AdventurePanel) requireClient() (*GameClient, error) {
	if p.client == nil {
		return nil, fmt.Errorf("尚未连接 Gateway，请先执行 connect")
	}
	return p.client, nil
}

func (p *AdventurePanel) requireLogin() (*GameClient, error) {
	client, err := p.requireClient()
	if err != nil {
		return nil, err
	}
	if !p.loggedIn {
		return nil, fmt.Errorf("尚未登录账号，请先执行 login 或通过菜单登录")
	}
	return client, nil
}

func (p *AdventurePanel) requireScene() (*GameClient, error) {
	client, err := p.requireLogin()
	if err != nil {
		return nil, err
	}
	if !p.inScene || !client.HasEnteredScene() {
		return nil, fmt.Errorf("角色尚未进入场景，使用 enter <roleID> 或菜单进入")
	}
	return client, nil
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
	fmt.Println("  attack <handle>          对指定实体发起普通攻击")
	fmt.Println("  quit / exit              退出客户端")
}

func (p *AdventurePanel) logStatus(client *GameClient) {
	status := client.RoleStatus()
	if status.RoleID == 0 {
		p.appendLog("⚠️ 角色信息尚未同步，可稍后再试")
		return
	}
	p.appendLog("🏷️ 角色 %s (#%d) Lv.%d Scene=%d",
		status.RoleName, status.RoleID, status.Level, status.SceneID)
	p.appendLog("    Pos=(%d,%d) Handle=%d HP=%d MP=%d State=%d",
		status.PosX, status.PosY, status.EntityHandle, status.HP, status.MP, status.StateFlags)
}

func (p *AdventurePanel) logLook(client *GameClient) {
	entities := client.ObservedEntities()
	p.appendLog("👁️ 视野扫描：")
	if len(entities) == 0 {
		p.appendLog("  附近没有其他实体")
		return
	}
	for idx, ent := range entities {
		hp := "-"
		if ent.HasHp {
			hp = fmt.Sprintf("%d", ent.Hp)
		}
		mp := "-"
		if ent.HasMp {
			mp = fmt.Sprintf("%d", ent.Mp)
		}
		p.appendLog("  [%d] Handle=%d Pos=(%d,%d) HP=%s MP=%s State=%d",
			idx+1, ent.Handle, ent.PosX, ent.PosY, hp, mp, ent.StateFlags)
	}
}

func (p *AdventurePanel) tryAutoConnect() {
	if err := p.connect(p.gatewayAddr); err != nil {
		log.Warnf("auto connect failed: %v", err)
		p.appendLog("⚠️ 自动连接失败，请手动输入 connect <ip:port>")
	}
}

type menuOption struct {
	keys    []string
	label   string
	handler func() error
}

func (p *AdventurePanel) renderFrame() {
	width := 58
	inner := width - 2
	border := strings.Repeat("─", inner)
	fmt.Printf("\n┌%s┐\n", border)
	fmt.Printf("│ %s │\n", padRight(p.headerLine(), inner-2))
	fmt.Printf("│ %s │\n", padRight(p.subHeaderLine(), inner-2))
	fmt.Printf("├%s┤\n", border)
	fmt.Printf("│ %s │\n", padRight("[Log]", inner-2))
	for _, line := range p.logLines(p.logLimit) {
		if line == "" {
			fmt.Printf("│ %s │\n", padRight("", inner-2))
		} else {
			fmt.Printf("│ - %s │\n", padRight(line, inner-4))
		}
	}
	fmt.Printf("├%s┤\n", border)
	options := p.currentMenuOptions()
	for _, opt := range options {
		keyLabel := strings.Join(opt.keys, "/")
		fmt.Printf("│ %s │\n", padRight(fmt.Sprintf("%s. %s", keyLabel, opt.label), inner-2))
	}
	fmt.Printf("├%s┤\n", border)
	fmt.Printf("│ %s │\n", padRight("cli> 输入数字或命令", inner-2))
	fmt.Printf("└%s┘\n", border)
	fmt.Print("cli> ")
}

func (p *AdventurePanel) headerLine() string {
	return fmt.Sprintf("POST-APOC CLI v0.1  | srv %s", padRight(p.headerServer(), 11))
}

func (p *AdventurePanel) subHeaderLine() string {
	return fmt.Sprintf("当前账号: %s | 当前时间: %s", padRight(p.currentAccount(), 8), p.currentTime())
}

func (p *AdventurePanel) headerServer() string {
	if p.client != nil {
		return p.gatewayAddr
	}
	if p.gatewayAddr != "" {
		return p.gatewayAddr
	}
	return "-"
}

func (p *AdventurePanel) currentAccount() string {
	if p.loggedIn && p.client != nil {
		if acc := p.client.AccountName(); acc != "" {
			return acc
		}
	}
	return "未登录"
}

func (p *AdventurePanel) currentTime() string {
	if p.client != nil {
		if serverMs, ok := p.client.LastServerTime(); ok {
			t := time.UnixMilli(serverMs)
			return t.Format("2006-01-02 15:04:05")
		}
	}
	return servertime.Now().Format("2006-01-02 15:04:05")
}

func (p *AdventurePanel) appendLog(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.logs = append(p.logs, msg)
	if len(p.logs) > 50 {
		p.logs = p.logs[len(p.logs)-50:]
	}
}

func (p *AdventurePanel) logLines(limit int) []string {
	lines := make([]string, limit)
	count := len(p.logs)
	if count == 0 {
		lines[limit-1] = "暂无日志"
		return lines
	}
	for i := 0; i < limit; i++ {
		idx := count - limit + i
		if idx >= 0 && idx < count {
			lines[i] = p.logs[idx]
		} else {
			lines[i] = ""
		}
	}
	return lines
}

func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

func (p *AdventurePanel) readLine() (string, error) {
	text, err := p.input.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func (p *AdventurePanel) promptInput(label string) (string, error) {
	fmt.Printf("%s: ", label)
	return p.readLine()
}

func (p *AdventurePanel) currentMenuOptions() []menuOption {
	if p.client == nil {
		return []menuOption{
			{keys: []string{"1"}, label: "连接默认网关", handler: func() error { return p.connect("") }},
			{keys: []string{"2"}, label: "输入地址连接", handler: p.promptConnect},
			{keys: []string{"0", "q"}, label: "退出", handler: func() error { return errPanelExit }},
		}
	}
	if !p.loggedIn {
		return []menuOption{
			{keys: []string{"1"}, label: "注册账号", handler: p.actionRegister},
			{keys: []string{"2"}, label: "登录账号", handler: p.actionLogin},
			{keys: []string{"3"}, label: "断开连接", handler: p.disconnect},
			{keys: []string{"0", "q"}, label: "退出", handler: func() error { return errPanelExit }},
		}
	}
	if !p.inScene {
		return []menuOption{
			{keys: []string{"1"}, label: "查看角色列表", handler: p.actionListRoles},
			{keys: []string{"2"}, label: "创建角色", handler: p.actionCreateRole},
			{keys: []string{"3"}, label: "进入角色", handler: p.actionEnterRole},
			{keys: []string{"4"}, label: "注销账号", handler: p.actionLogout},
			{keys: []string{"5"}, label: "断开连接", handler: p.disconnect},
			{keys: []string{"0", "q"}, label: "退出", handler: func() error { return errPanelExit }},
		}
	}
	return []menuOption{
		{keys: []string{"1"}, label: "查看状态", handler: func() error {
			client, err := p.requireScene()
			if err != nil {
				return err
			}
			p.logStatus(client)
			return nil
		}},
		{keys: []string{"2"}, label: "观察周围", handler: func() error {
			client, err := p.requireScene()
			if err != nil {
				return err
			}
			p.logLook(client)
			return nil
		}},
		{keys: []string{"3"}, label: "移动角色", handler: p.actionMovePrompt},
		{keys: []string{"4"}, label: "普通攻击", handler: p.actionAttackPrompt},
		{keys: []string{"5"}, label: "断开连接", handler: p.disconnect},
		{keys: []string{"0", "q"}, label: "退出", handler: func() error { return errPanelExit }},
	}
}

func (p *AdventurePanel) findMenuAction(input string) func() error {
	opts := p.currentMenuOptions()
	for _, opt := range opts {
		for _, key := range opt.keys {
			if strings.EqualFold(input, key) {
				return opt.handler
			}
		}
	}
	return nil
}

func (p *AdventurePanel) promptConnect() error {
	addr, err := p.promptInput("输入地址 (ip:port)")
	if err != nil {
		return err
	}
	if addr == "" {
		return fmt.Errorf("地址不可为空")
	}
	return p.connect(addr)
}

func (p *AdventurePanel) actionRegister() error {
	client, err := p.requireClient()
	if err != nil {
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
	if err := client.RegisterAccount(account, password); err != nil {
		return err
	}
	p.appendLog("✅ 注册成功: %s", account)
	return nil
}

func (p *AdventurePanel) actionLogin() error {
	client, err := p.requireClient()
	if err != nil {
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
	if err := client.LoginAccount(account, password); err != nil {
		return err
	}
	p.loggedIn = true
	p.inScene = false
	p.activeRole = 0
	p.appendLog("✅ 登录成功: %s", account)
	return nil
}

func (p *AdventurePanel) actionListRoles() error {
	client, err := p.requireLogin()
	if err != nil {
		return err
	}
	roles, err := client.ListRoles()
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
	client, err := p.requireLogin()
	if err != nil {
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
	job := uint32(1)
	sex := uint32(1)
	if jobStr != "" {
		if v, err := strconv.ParseUint(jobStr, 10, 32); err == nil {
			job = uint32(v)
		}
	}
	if sexStr != "" {
		if v, err := strconv.ParseUint(sexStr, 10, 32); err == nil {
			sex = uint32(v)
		}
	}
	if err := client.CreateRole(name, job, sex); err != nil {
		return err
	}
	p.appendLog("✨ 创建角色：%s", name)
	return nil
}

func (p *AdventurePanel) actionEnterRole() error {
	client, err := p.requireLogin()
	if err != nil {
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
	if err := client.EnterGame(roleID); err != nil {
		return err
	}
	p.inScene = client.HasEnteredScene()
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
	client, err := p.requireScene()
	if err != nil {
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
	if err := client.NudgeMove(int32(dx), int32(dy)); err != nil {
		return err
	}
	p.appendLog("🚶 移动 Δ(%d,%d)", dx, dy)
	return nil
}

func (p *AdventurePanel) actionAttackPrompt() error {
	client, err := p.requireScene()
	if err != nil {
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
	if err := client.CastNormalAttack(target); err != nil {
		return err
	}
	p.appendLog("⚔️ 攻击目标 %d", target)
	if hit, err := client.WaitForSkillResult(target, 3*time.Second); err == nil {
		p.appendLog("   ➤ 伤害=%d HP=%d MP=%d State=%d",
			hit.Damage,
			attrValueOrZero(hit.Attrs, attrdef.AttrHP),
			attrValueOrZero(hit.Attrs, attrdef.AttrMP),
			hit.StateFlags,
		)
	}
	return nil
}
