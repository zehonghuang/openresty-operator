package utils

import (
	"strconv"
	"strings"
)

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

func ParseListenPort(listen string) int32 {
	fields := strings.Fields(listen)
	for _, f := range fields {
		if strings.HasPrefix(f, "[") {
			// IPv6 形式，跳过
			continue
		}
		if p, err := strconv.Atoi(f); err == nil {
			return int32(p)
		}
	}
	return 80 // 默认 fallback
}

func SanitizeLogFormat(format string) string {
	format = strings.ReplaceAll(format, "\r\n", " ")
	format = strings.ReplaceAll(format, "\n", " ")
	format = strings.ReplaceAll(format, "\r", " ")
	return strings.TrimSpace(format)
}
