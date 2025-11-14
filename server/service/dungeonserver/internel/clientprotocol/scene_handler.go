package clientprotocol

import (
	"google.golang.org/protobuf/proto"
	"math/rand"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func init() {
	Register(uint16(protocol.C2SProtocol_C2SChangeScene), handleChangeScene)
}

func handleChangeScene(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SChangeSceneReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}

	// 获取当前场景ID和副本ID
	currentSceneId := entity.GetSceneId()
	currentFuBenId := scene.GetFuBenId()
	targetSceneId := req.SceneId

	// 获取当前副本实例
	fuBen := scene.GetFuBen()
	if fuBen == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "副本不存在")
	}

	// 获取目标场景
	targetScene := fuBen.GetScene(targetSceneId)
	if targetScene == nil {
		resp := &protocol.S2CChangeSceneResultReq{
			Success: false,
			Message: "目标场景不存在",
			SceneId: targetSceneId,
		}
		return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
	}

	// 检查是否在同一副本内（不同副本不能跨副本场景切换）
	if targetScene.GetFuBenId() != currentFuBenId {
		resp := &protocol.S2CChangeSceneResultReq{
			Success: false,
			Message: "不能跨副本切换场景",
			SceneId: targetSceneId,
		}
		return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
	}

	// 如果目标场景就是当前场景，直接返回成功
	if targetSceneId == currentSceneId {
		resp := &protocol.S2CChangeSceneResultReq{
			Success: true,
			Message: "切换成功",
			SceneId: targetSceneId,
		}
		return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
	}

	// 从当前场景移除实体
	scene.RemoveEntity(entity.GetHdl())

	// 将实体添加到目标场景
	// 从场景配置获取出生点
	configMgr := jsonconf.GetConfigManager()
	sceneConfig, _ := configMgr.GetSceneConfig(targetSceneId)
	var x, y uint32
	if sceneConfig != nil && sceneConfig.BornArea != nil {
		// 从出生点范围随机选择
		bornArea := sceneConfig.BornArea
		if bornArea.X2 > bornArea.X1 && bornArea.Y2 > bornArea.Y1 {
			x = bornArea.X1 + uint32(rand.Intn(int(bornArea.X2-bornArea.X1)))
			y = bornArea.Y1 + uint32(rand.Intn(int(bornArea.Y2-bornArea.Y1)))
		} else {
			// 使用默认位置
			x, y = 100, 100
		}
	} else {
		// 使用默认位置
		x, y = 100, 100
	}
	entity.SetPosition(x, y)
	targetScene.AddEntity(entity)

	log.Infof("Entity %d changed scene from %d to %d", entity.GetHdl(), currentSceneId, targetSceneId)

	// 发送切换成功响应
	resp := &protocol.S2CChangeSceneResultReq{
		Success: true,
		Message: "切换成功",
		SceneId: targetSceneId,
	}
	return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
}
