package tool

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var j = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func init() {
	extra.RegisterFuzzyDecoders()
}

// JsonMarshal 利用json-iterator进行json编码
func JsonMarshal(v interface{}) ([]byte, error) {
	return j.Marshal(v)
}

// JsonUnmarshal 利用json-iterator进行json解码
func JsonUnmarshal(data []byte, v interface{}) error {
	return j.Unmarshal(data, v)
}
