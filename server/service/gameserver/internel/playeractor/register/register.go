package register

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/playeractor/controller"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/level"
	"postapocgame/server/service/gameserver/internel/playeractor/router"
	"postapocgame/server/service/gameserver/internel/playeractor/skill"
)

func All(rt *deps.Runtime) {
	registerSkillHandlers()

	// 注册所有系统工厂
	level.RegisterSystemFactory(rt)
	skill.RegisterSystemFactory(rt)
}

// registerSkillHandlers 注册技能相关协议处理器
func registerSkillHandlers() {
	skillController := controller.NewSkillController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUseSkill), skillController.HandleUseSkill)
}
