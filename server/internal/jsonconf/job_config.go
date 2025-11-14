/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 职业配置
**/

package jsonconf

// JobConfig 职业配置
type JobConfig struct {
	JobId      uint32   `json:"jobId"`      // 职业ID
	Name       string   `json:"name"`       // 职业名称
	Desc       string   `json:"desc"`       // 职业描述
	BaseAttrs  AttrVec  `json:"baseAttrs"`  // 基础属性
	SkillIds   []uint32 `json:"skillIds"`   // 初始技能ID列表
	WeaponType uint32   `json:"weaponType"` // 武器类型
}
