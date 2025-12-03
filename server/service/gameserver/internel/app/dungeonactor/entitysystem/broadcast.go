// broadcast.go 封装场景层广播逻辑，避免各系统重复编解码。
package entitysystem

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
)

// BroadcastSceneProto 将消息广播给场景内所有实体
func BroadcastSceneProto(scene iface.IScene, protoId uint16, payload proto.Message) {
	if scene == nil || payload == nil {
		return
	}
	data, err := proto.Marshal(payload)
	if err != nil {
		log.Errorf("broadcast marshal failed: %v", err)
		return
	}
	for _, et := range scene.GetAllEntities() {
		if et == nil {
			continue
		}
		_ = et.SendMessage(protoId, data)
	}
}
