package gateway

import (
	"reflect"
	"strings"
)

// CheckName 判断 input 中 fields 中指定的字段是否存在
// 不检查 field 中不存在的字段，默认检查"bucket", "key","object"
func CheckName(input interface{}, fields ...string) bool {

	names := make(map[string]bool, len(fields)+2)
	names["bucket"] = true
	names["key"] = true
	names["object"] = true

	for _, v := range fields {
		names[strings.ToLower(v)] = true
	}

	v := reflect.ValueOf(input)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return false
	}

	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		vName := strings.ToLower(v.Type().Field(i).Name)
		// input 内部值为 ptr *string
		if names[vName] && v.Field(i).IsNil() {
			return false
		}
	}
	return true
}
