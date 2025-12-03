package attr

import (
	"testing"

	"postapocgame/server/internal/protocol"
)

func TestCompareAttrVecUseCase_Execute(t *testing.T) {
	uc := NewCompareAttrVecUseCase()

	t.Run("两个 nil 应该相等", func(t *testing.T) {
		result := uc.Execute(nil, nil)
		if !result {
			t.Error("Execute(nil, nil) = false, want true")
		}
	})

	t.Run("一个 nil 一个非 nil 应该不相等", func(t *testing.T) {
		vec := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
			},
		}
		result := uc.Execute(nil, vec)
		if result {
			t.Error("Execute(nil, vec) = true, want false")
		}
		result = uc.Execute(vec, nil)
		if result {
			t.Error("Execute(vec, nil) = true, want false")
		}
	})

	t.Run("相同属性向量应该相等", func(t *testing.T) {
		vec1 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
				{Type: 2, Value: 200},
			},
		}
		vec2 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
				{Type: 2, Value: 200},
			},
		}
		result := uc.Execute(vec1, vec2)
		if !result {
			t.Error("Execute(vec1, vec2) = false, want true")
		}
	})

	t.Run("不同属性值应该不相等", func(t *testing.T) {
		vec1 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
			},
		}
		vec2 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 200},
			},
		}
		result := uc.Execute(vec1, vec2)
		if result {
			t.Error("Execute(vec1, vec2) = true, want false (different values)")
		}
	})

	t.Run("不同属性类型应该不相等", func(t *testing.T) {
		vec1 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
			},
		}
		vec2 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 2, Value: 100},
			},
		}
		result := uc.Execute(vec1, vec2)
		if result {
			t.Error("Execute(vec1, vec2) = true, want false (different types)")
		}
	})

	t.Run("不同数量属性应该不相等", func(t *testing.T) {
		vec1 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
			},
		}
		vec2 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
				{Type: 2, Value: 200},
			},
		}
		result := uc.Execute(vec1, vec2)
		if result {
			t.Error("Execute(vec1, vec2) = true, want false (different count)")
		}
	})

	t.Run("属性顺序不同但内容相同应该相等", func(t *testing.T) {
		vec1 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
				{Type: 2, Value: 200},
			},
		}
		vec2 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 2, Value: 200},
				{Type: 1, Value: 100},
			},
		}
		result := uc.Execute(vec1, vec2)
		if !result {
			t.Error("Execute(vec1, vec2) = false, want true (same content, different order)")
		}
	})

	t.Run("忽略 nil 属性", func(t *testing.T) {
		vec1 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
				nil,
				{Type: 2, Value: 200},
			},
		}
		vec2 := &protocol.AttrVec{
			Attrs: []*protocol.AttrSt{
				{Type: 1, Value: 100},
				{Type: 2, Value: 200},
			},
		}
		result := uc.Execute(vec1, vec2)
		if !result {
			t.Error("Execute(vec1, vec2) = false, want true (nil attributes should be ignored)")
		}
	})
}
