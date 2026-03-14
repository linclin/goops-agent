package utils

import (
	"encoding/json"
)

// JsonStr 将数据序列化为 JSON 字符串
//
// 该函数将任意数据结构序列化为 JSON 格式的字符串。
// 如果序列化失败，返回空字符串。
//
// 参数:
//   - data: 要序列化的数据，可以是结构体、map、切片等
//
// 返回:
//   - string: JSON 格式的字符串，失败时返回空字符串
//
// 使用示例:
//
//	// 序列化结构体
//	user := User{Name: "张三", Age: 25}
//	jsonStr := utils.JsonStr(user)
//	// 输出: {"name":"张三","age":25}
//
//	// 序列化 map
//	data := map[string]interface{}{"key": "value"}
//	jsonStr := utils.JsonStr(data)
//	// 输出: {"key":"value"}
//
//	// 序列化切片
//	list := []int{1, 2, 3}
//	jsonStr := utils.JsonStr(list)
//	// 输出: [1,2,3]
func JsonStr(data interface{}) string {
	// 使用 json.Marshal 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		// 序列化失败返回空字符串
		return ""
	}
	// 返回 JSON 字符串
	return string(jsonData)
}
