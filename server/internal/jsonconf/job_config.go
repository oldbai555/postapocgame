/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 职业配置
**/

package jsonconf

// JobConfig 职业配置
type JobConfig struct {
	JobId    uint32   `json:"jobId"`    // 职业ID
	SkillIds []uint32 `json:"skillIds"` // 初始技能ID列表
}
