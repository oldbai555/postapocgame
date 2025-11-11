/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// VipConfig VIP配置
type VipConfig struct {
	Level       uint32         `json:"level"`       // VIP等级
	ExpNeeded   uint64         `json:"expNeeded"`   // 升到下一级所需经验
	Privileges  []VipPrivilege `json:"privileges"`  // 特权列表
	Description string         `json:"description"` // 描述
}

// VipPrivilege VIP特权
type VipPrivilege struct {
	Type  uint32 `json:"type"`  // 特权类型: 1=背包扩容 2=副本次数 3=经验加成
	Value uint32 `json:"value"` // 特权值
}
