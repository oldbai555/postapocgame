package clientprotocol

import (
	"google.golang.org/protobuf/proto"
	"math"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func init() {
	Register(uint16(protocol.C2SProtocol_C2SGetNearestMonster), handleGetNearestMonster)
}

func handleGetNearestMonster(role iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SGetNearestMonsterReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(role)
	if err != nil {
		return err
	}

	// 获取角色位置
	rolePos := role.GetPosition()
	if rolePos == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "role position not found")
	}

	// 获取场景中所有实体
	allEntities := scene.GetAllEntities()

	// 查找最近的怪物
	var nearestMonster iface.IEntity
	var minDistance float64 = math.MaxFloat64

	for _, e := range allEntities {
		// 只处理怪物实体
		if e.GetEntityType() != uint32(protocol.EntityType_EtMonster) {
			continue
		}

		// 跳过死亡的怪物
		if e.IsDead() {
			continue
		}

		// 计算距离
		monsterPos := e.GetPosition()
		if monsterPos == nil {
			continue
		}

		dx := float64(monsterPos.X) - float64(rolePos.X)
		dy := float64(monsterPos.Y) - float64(rolePos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < minDistance {
			minDistance = distance
			nearestMonster = e
		}
	}

	// 构造响应
	resp := &protocol.S2CGetNearestMonsterResultReq{
		Success: nearestMonster != nil,
	}

	if nearestMonster != nil {
		monsterPos := nearestMonster.GetPosition()
		monsterEntity, ok := nearestMonster.(*entity.MonsterEntity)
		if ok {
			resp.MonsterHdl = nearestMonster.GetHdl()
			resp.MonsterId = monsterEntity.GetMonsterId()
			resp.MonsterName = monsterEntity.GetName()
			if monsterPos != nil {
				resp.PosX = monsterPos.X
				resp.PosY = monsterPos.Y
			}
			resp.Distance = float32(minDistance)
			resp.Message = "找到最近怪物"
		} else {
			resp.Success = false
			resp.Message = "怪物实体类型错误"
		}
	} else {
		resp.Message = "未找到怪物"
	}

	return role.SendProtoMessage(uint16(protocol.S2CProtocol_S2CGetNearestMonsterResult), resp)
}
