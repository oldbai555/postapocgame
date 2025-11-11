/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// MoneyConfig 货币配置
type MoneyConfig struct {
	MoneyId     uint32 `json:"moneyId"`     // 货币Id: 1=金币 2=钻石 3=铜币
	Name        string `json:"name"`        // 货币名称
	Icon        string `json:"icon"`        // 图标
	MaxAmount   uint64 `json:"maxAmount"`   // 最大持有量
	Description string `json:"description"` // 描述
}
