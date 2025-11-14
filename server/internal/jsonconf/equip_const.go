package jsonconf

// 装备槽位定义
const (
	EquipSlotWeapon    uint32 = 1 // 武器
	EquipSlotHelmet    uint32 = 2 // 头盔
	EquipSlotArmor     uint32 = 3 // 护甲
	EquipSlotShoes     uint32 = 4 // 鞋子
	EquipSlotAccessory uint32 = 5 // 饰品
)

// GetEquipSlotName 获取装备槽位名称
func GetEquipSlotName(slot uint32) string {
	names := map[uint32]string{
		EquipSlotWeapon:    "武器",
		EquipSlotHelmet:    "头盔",
		EquipSlotArmor:     "护甲",
		EquipSlotShoes:     "鞋子",
		EquipSlotAccessory: "饰品",
	}
	if name, ok := names[slot]; ok {
		return name
	}
	return "未知槽位"
}
