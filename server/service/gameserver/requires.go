/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package main

import (
	// Phase 2 后剩余需要 init() 的包（非特性分片系统）
	_ "postapocgame/server/service/gameserver/internel/app/playeractor/controller"
	_ "postapocgame/server/service/gameserver/internel/app/playeractor/entity"
	_ "postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	_ "postapocgame/server/service/gameserver/internel/app/playeractor/level"
	_ "postapocgame/server/service/gameserver/internel/app/playeractor/message"
	// 注意：Bag/Money/Equip/Skill/Fuben/Recycle 已通过 register.RegisterAll() 显式注册，无需 blank import
)
