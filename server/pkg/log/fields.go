package log

import (
	"fmt"
	"sort"
	"strings"
)

// Fields 结构化日志字段
type Fields map[string]interface{}

func cloneFields(src Fields) Fields {
	if len(src) == 0 {
		return nil
	}
	dst := make(Fields, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeFields(a, b Fields) Fields {
	if len(a) == 0 {
		return cloneFields(b)
	}
	if len(b) == 0 {
		return cloneFields(a)
	}
	dst := make(Fields, len(a)+len(b))
	for k, v := range a {
		dst[k] = v
	}
	for k, v := range b {
		dst[k] = v
	}
	return dst
}

func formatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString(" |")
	for _, key := range keys {
		builder.WriteString(" ")
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(fmt.Sprint(fields[key]))
	}
	return builder.String()
}
