package customerr

// GetErrCode 获取错误码（兼容新旧错误类型）
func GetErrCode(err error) int32 {
	if err == nil {
		return 0
	}
	// 优先支持新的 CustomErr 类型
	if p, ok := err.(*CustomErr); ok {
		return p.Code
	}
	// 兼容旧的 Error 类型（如果存在）
	// 注意：旧的 Error 类型已删除，这里保留兼容性检查
	return -1
}

// GetErrMsgByErr 获取错误消息
func GetErrMsgByErr(err error) string {
	if err == nil {
		return "success"
	}
	// 优先支持新的 CustomErr 类型
	if x, ok := err.(*CustomErr); ok {
		return x.Message
	}
	return err.Error()
}
