package utils

import "encoding/json"

func AsJson(input interface{}) string {
	jsonB, _ := json.Marshal(input)
	return string(jsonB)
}

func AsPrettyJson(input interface{}) string {
	jsonB, _ := json.MarshalIndent(input, "", "  ")
	return string(jsonB)
}
