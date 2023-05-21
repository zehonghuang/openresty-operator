package utils

import "strings"

func SanitizeName(name string) string {
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}

func BoolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
