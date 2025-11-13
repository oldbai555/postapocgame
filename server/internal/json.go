package internal

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

func Marshal(v interface{}) ([]byte, error) {
	return j.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return j.Unmarshal(data, v)
}
