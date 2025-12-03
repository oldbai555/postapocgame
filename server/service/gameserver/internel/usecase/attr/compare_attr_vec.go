package attr

import (
	"postapocgame/server/internal/protocol"
)

// CompareAttrVecUseCase 属性向量比较用例
// 负责比较两个属性向量是否相等（纯业务逻辑）
type CompareAttrVecUseCase struct{}

// NewCompareAttrVecUseCase 创建属性向量比较用例
func NewCompareAttrVecUseCase() *CompareAttrVecUseCase {
	return &CompareAttrVecUseCase{}
}

// Execute 执行属性向量比较用例
// 返回两个属性向量是否相等
func (uc *CompareAttrVecUseCase) Execute(a, b *protocol.AttrVec) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a.Attrs) != len(b.Attrs) {
		return false
	}
	// 使用临时 map 进行属性值比较
	tmp := make(map[uint32]int64, len(a.Attrs))
	for _, attr := range a.Attrs {
		if attr == nil {
			continue
		}
		tmp[attr.Type] = attr.Value
	}
	for _, attr := range b.Attrs {
		if attr == nil {
			continue
		}
		value, ok := tmp[attr.Type]
		if !ok || value != attr.Value {
			return false
		}
		delete(tmp, attr.Type)
	}
	return len(tmp) == 0
}
