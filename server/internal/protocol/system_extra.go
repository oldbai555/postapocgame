package protocol

// system_extra.go
// 为兼容仍然依赖聊天系统 ID 的代码，这里补充 SysChat 的枚举值。
// 后续如果在 proto/csproto/system.proto 中正式恢复该枚举，请同步更新并重新生成 pb.go。

const (
	// SystemId_SysChat 聊天系统（仅用于 SystemAdapter 挂载 ID）
	// 注意：当前 proto 中未声明该枚举，实际持久化数据不会使用该 SystemId。
	SystemId_SysChat SystemId = 12
)
