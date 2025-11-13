package protocol

import (
	"postapocgame/server/pkg/customerr"
)

// InitErrorCodes 初始化错误码映射，将protocol枚举注册到customerr
// 建议在应用启动时（main函数开始处）调用
func InitErrorCodes() {
	// 设置默认错误码
	customerr.SetDefaultErrorCode(int32(ErrorCode_Internal_Error))

	// 批量注册所有错误码映射
	errorTags := map[int32]string{
		int32(ErrorCode_Success):         "Success",
		int32(ErrorCode_Internal_Error):  "Internal_Error",
		int32(ErrorCode_Param_Invalid):   "Param_Invalid",
		int32(ErrorCode_Network_Timeout): "Network_Timeout",
		int32(ErrorCode_Player_NotFound): "Player_NotFound",
		int32(ErrorCode_Item_NotEnough):  "Item_NotEnough",
		// 后续新增错误码在这里继续添加
	}
	customerr.RegisterErrorTags(errorTags)
}
