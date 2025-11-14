package main

import (
	"context"
	"fmt"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/pkg/log"
	"time"

	uuid "github.com/satori/go.uuid"
)

// TestIdentity 测试账号信息
type TestIdentity struct {
	Account  string
	Password string
	RoleName string
}

func newTestIdentity(prefix string) TestIdentity {
	id := uuid.NewV4().String()
	return TestIdentity{
		Account:  fmt.Sprintf("%s_acc_%s", prefix, id),
		Password: "Test@123",
		RoleName: fmt.Sprintf("%s_role_%s", prefix, id[:4]),
	}
}

// IntegrationScenario 集成测试流程
type IntegrationScenario struct {
	ctx       context.Context
	manager   *ClientManager
	clientA   *GameClient
	clientB   *GameClient
	identityA TestIdentity
	identityB TestIdentity
}

func NewIntegrationScenario(ctx context.Context, mgr *ClientManager) *IntegrationScenario {
	return &IntegrationScenario{
		ctx:     ctx,
		manager: mgr,
	}
}

func (s *IntegrationScenario) Run() error {
	s.clientA = s.manager.CreateClient("tester-A", GatewayAddr)
	s.clientB = s.manager.CreateClient("tester-B", GatewayAddr)

	if err := s.clientA.Start(s.ctx); err != nil {
		return fmt.Errorf("start client A failed: %w", err)
	}
	if err := s.clientB.Start(s.ctx); err != nil {
		return fmt.Errorf("start client B failed: %w", err)
	}

	s.identityA = newTestIdentity("A")
	s.identityB = newTestIdentity("B")

	log.Infof("▶️ 玩家A 登录流程: %+v", s.identityA)
	if err := s.clientA.RunLoginFlow(s.identityA); err != nil {
		return fmt.Errorf("client A flow failed: %w", err)
	}
	log.Infof("▶️ 玩家B 登录流程: %+v", s.identityB)
	if err := s.clientB.RunLoginFlow(s.identityB); err != nil {
		return fmt.Errorf("client B flow failed: %w", err)
	}

	log.Infof("✅ 两名玩家均已进入默认副本，开始AOI测试")
	if err := s.establishMutualVision(); err != nil {
		return err
	}

	log.Infof("✅ AOI 可见验证完成，开始战斗交互测试")
	if err := s.performCombatDemo(); err != nil {
		return err
	}

	log.Infof("🎉 集成测试流程执行完毕")
	return nil
}

func (s *IntegrationScenario) establishMutualVision() error {
	if err := s.clientB.NudgeMove(30, 0); err != nil {
		return fmt.Errorf("client B move failed: %w", err)
	}
	if _, err := s.clientA.WaitForEntityInView(5 * time.Second); err != nil {
		return fmt.Errorf("client A did not see client B: %w", err)
	}

	if err := s.clientA.NudgeMove(-30, 0); err != nil {
		return fmt.Errorf("client A move failed: %w", err)
	}
	if _, err := s.clientB.WaitForEntityInView(5 * time.Second); err != nil {
		return fmt.Errorf("client B did not see client A: %w", err)
	}
	return nil
}

func (s *IntegrationScenario) performCombatDemo() error {
	target := s.clientB.EntityHandle()
	if target == 0 {
		return fmt.Errorf("client B entity handle not ready")
	}
	if err := s.clientA.CastNormalAttack(target); err != nil {
		return fmt.Errorf("client A cast skill failed: %w", err)
	}

	hitA, err := s.clientA.WaitForSkillResult(target, 5*time.Second)
	if err != nil {
		return fmt.Errorf("client A wait skill result failed: %w", err)
	}
	log.Infof("[A视角] 对B造成伤害: %d, 目标当前HP=%d MP=%d State=%d",
		hitA.Damage,
		attrValueOrZero(hitA.Attrs, attrdef.AttrHP),
		attrValueOrZero(hitA.Attrs, attrdef.AttrMP),
		hitA.StateFlags,
	)

	hitB, err := s.clientB.WaitForSkillResult(target, 5*time.Second)
	if err != nil {
		return fmt.Errorf("client B wait skill result failed: %w", err)
	}
	log.Infof("[B视角] 收到来自A的攻击: %d, 自身HP=%d MP=%d State=%d",
		hitB.Damage,
		attrValueOrZero(hitB.Attrs, attrdef.AttrHP),
		attrValueOrZero(hitB.Attrs, attrdef.AttrMP),
		hitB.StateFlags,
	)
	return nil
}
