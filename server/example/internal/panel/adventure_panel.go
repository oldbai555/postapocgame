package panel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"postapocgame/server/example/internal/client"
	"postapocgame/server/example/internal/systems"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
)

var errPanelExit = errors.New("panel exit")

// AdventurePanel 文字冒险式交互面板
type AdventurePanel struct {
	ctx            context.Context
	manager        *client.Manager
	core           *client.Core
	systems        *systems.Set
	gatewayAddr    string
	loggedIn       bool
	inScene        bool
	activeRole     uint64
	logs           []string
	logLimit       int
	input          *bufio.Reader
	scriptRecorder *os.File
	suppressRecord int
}

func NewAdventurePanel(ctx context.Context, mgr *client.Manager) *AdventurePanel {
	return &AdventurePanel{
		ctx:         ctx,
		manager:     mgr,
		gatewayAddr: client.DefaultGatewayAddr,
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

func (p *AdventurePanel) connect(addr string) error {
	if addr == "" {
		addr = p.gatewayAddr
	}
	if addr == "" {
		addr = client.DefaultGatewayAddr
	}
	if addr == "" {
		return fmt.Errorf("未配置 Gateway 地址")
	}
	if p.core != nil {
		_ = p.disconnect()
	}
	clientID := fmt.Sprintf("panel-%s", tool.GenUUID())
	core := p.manager.CreateClient(clientID, addr)
	if err := core.Start(p.ctx); err != nil {
		p.manager.DestroyClient(clientID)
		return err
	}
	p.core = core
	p.systems = systems.NewSet(core)
	p.gatewayAddr = addr
	p.loggedIn = false
	p.inScene = false
	p.activeRole = 0
	p.appendLog("🔗 已连接 Gateway %s", addr)
	return nil
}

func (p *AdventurePanel) disconnect() error {
	if p.core == nil {
		return nil
	}
	playerID := p.core.GetPlayerID()
	p.manager.DestroyClient(playerID)
	p.core = nil
	p.systems = nil
	p.loggedIn = false
	p.inScene = false
	p.activeRole = 0
	if p.scriptRecorder != nil {
		_ = p.scriptRecorder.Close()
		p.scriptRecorder = nil
	}
	p.appendLog("🔌 已断开 Gateway 连接")
	return nil
}

func (p *AdventurePanel) requireConnection() (*client.Core, error) {
	if p.core == nil {
		return nil, fmt.Errorf("尚未连接 Gateway，请先执行 connect")
	}
	return p.core, nil
}

func (p *AdventurePanel) requireLogin() (*client.Core, error) {
	core, err := p.requireConnection()
	if err != nil {
		return nil, err
	}
	if !p.loggedIn {
		return nil, fmt.Errorf("尚未登录账号，请先执行 login 或通过菜单登录")
	}
	return core, nil
}

func (p *AdventurePanel) requireScene() (*client.Core, error) {
	core, err := p.requireLogin()
	if err != nil {
		return nil, err
	}
	if !p.inScene || !core.HasEnteredScene() {
		return nil, fmt.Errorf("角色尚未进入场景，使用 enter <roleID> 或菜单进入")
	}
	return core, nil
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
	return fmt.Sprintf("POST-APOC CLI v0.2  | srv %s", padRight(p.headerServer(), 11))
}

func (p *AdventurePanel) subHeaderLine() string {
	return fmt.Sprintf("当前账号: %s | 当前时间: %s", padRight(p.currentAccount(), 8), p.currentTime())
}

func (p *AdventurePanel) headerServer() string {
	if p.core != nil {
		return p.gatewayAddr
	}
	if p.gatewayAddr != "" {
		return p.gatewayAddr
	}
	return "-"
}

func (p *AdventurePanel) currentAccount() string {
	if p.loggedIn && p.core != nil {
		if acc := p.core.AccountName(); acc != "" {
			return acc
		}
	}
	return "未登录"
}

func (p *AdventurePanel) currentTime() string {
	if p.core != nil {
		if serverMs, ok := p.core.LastServerTime(); ok {
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
	if p.core == nil {
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
			if _, err := p.requireScene(); err != nil {
				return err
			}
			p.logStatus()
			return nil
		}},
		{keys: []string{"2"}, label: "观察周围", handler: func() error {
			if _, err := p.requireScene(); err != nil {
				return err
			}
			p.logLook()
			return nil
		}},
		{keys: []string{"3"}, label: "移动 ΔX/ΔY", handler: p.actionMovePrompt},
		{keys: []string{"4"}, label: "自动寻路到坐标", handler: p.actionMoveToPrompt},
		{keys: []string{"5"}, label: "查看背包", handler: p.actionShowBag},
		{keys: []string{"6"}, label: "使用物品", handler: p.actionUseItemPrompt},
		{keys: []string{"7"}, label: "进入副本", handler: p.actionEnterDungeonPrompt},
		{keys: []string{"8"}, label: "执行 GM 命令", handler: p.actionGMCommandPrompt},
		{keys: []string{"9"}, label: "运行脚本 Demo", handler: p.actionRunDemoScript},
		{keys: []string{"a"}, label: "普通攻击", handler: p.actionAttackPrompt},
		{keys: []string{"d"}, label: "断开连接", handler: p.disconnect},
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

func (p *AdventurePanel) logStatus() {
	status := p.systems.Scene.Status()
	if status.RoleID == 0 {
		p.appendLog("⚠️ 角色信息尚未同步，可稍后再试")
		return
	}
	p.appendLog("🏷️ 角色 %s (#%d) Lv.%d Scene=%d",
		status.RoleName, status.RoleID, status.Level, status.SceneID)
	p.appendLog("    Pos=(%d,%d) Handle=%d HP=%d MP=%d State=%d",
		status.PosX, status.PosY, status.EntityHandle, status.HP, status.MP, status.StateFlags)
}

func (p *AdventurePanel) logLook() {
	entities := p.systems.Scene.ObservedEntities()
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

// --- action helpers moved to actions.go ---
