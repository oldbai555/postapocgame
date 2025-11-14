package jsonconf

// ConsumeConfig 通用消耗配置
type ConsumeConfig struct {
	ConsumeId   uint32        `json:"consumeId"`
	Items       []*ItemAmount `json:"items"`
	Description string        `json:"description"`
}
