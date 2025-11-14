package entityhelper

import (
	"fmt"

	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

// BuildEntitySnapshot 构建客户端可见的实体数据
func BuildEntitySnapshot(entity iface.IEntity) *protocol.EntitySt {
	if entity == nil {
		return nil
	}
	pos := entity.GetPosition()
	if pos == nil {
		pos = &argsdef.Position{}
	}
	showName := resolveEntityName(entity)

	return &protocol.EntitySt{
		Hdl:        entity.GetHdl(),
		Id:         entity.GetId(),
		Et:         entity.GetEntityType(),
		PosX:       pos.X,
		PosY:       pos.Y,
		SceneId:    entity.GetSceneId(),
		FbId:       entity.GetFuBenId(),
		Level:      entity.GetLevel(),
		ShowName:   showName,
		Attrs:      BuildAttrMap(entity),
		StateFlags: entity.GetStateFlags(),
	}
}

func resolveEntityName(entity iface.IEntity) string {
	type named interface {
		GetName() string
	}
	if n, ok := entity.(named); ok {
		return n.GetName()
	}
	return fmt.Sprintf("Entity-%d", entity.GetId())
}
