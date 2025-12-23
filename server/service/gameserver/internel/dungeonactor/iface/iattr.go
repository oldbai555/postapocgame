/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性系统接口
**/

package iface

// IAttrSys 属性系统接口
type IAttrSys interface {
	// GetAttrValue 获取属性值
	GetAttrValue(attrType uint32) int64

	// SetAttrValue 设置属性值
	SetAttrValue(attrType uint32, value int64)

	// AddAttrValue 增加属性值
	AddAttrValue(attrType uint32, delta int64)

	// RunOne 每帧更新（由实体 RunOne 调用）
	RunOne()
}
