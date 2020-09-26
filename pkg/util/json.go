package util

import "strings"

var replacer = strings.NewReplacer("\"", "")

func removeDoulbeQuotes(s string) string {
	return replacer.Replace(s)
}

// ConvertJSONToString remove "" from string
func ConvertJSONToString(b []byte) string {

	return removeDoulbeQuotes(string(b))
}
