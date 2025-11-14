/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 通用属性结构定义
**/

package jsonconf

// Attr 通用属性结构
type Attr struct {
	Type  uint32 `json:"type"`  // 属性枚举
	Value uint64 `json:"value"` // 属性值
}

// AttrVec 属性向量
type AttrVec []*Attr

// ToMap 将属性向量转换为map
func (av AttrVec) ToMap() map[uint32]uint64 {
	result := make(map[uint32]uint64)
	for _, attr := range av {
		if attr != nil {
			result[attr.Type] = attr.Value
		}
	}
	return result
}

// FromMap 从map创建属性向量
func (av *AttrVec) FromMap(m map[uint32]uint64) {
	*av = make(AttrVec, 0, len(m))
	for k, v := range m {
		*av = append(*av, &Attr{
			Type:  k,
			Value: v,
		})
	}
}
