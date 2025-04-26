package utils

import (
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func SetFrom[T comparable](items []T) map[T]struct{} {
	result := make(map[T]struct{}, len(items))
	for _, item := range items {
		result[item] = struct{}{}
	}
	return result
}

func DrainChan[T any](ch <-chan T) []T {
	var result []T
	for v := range ch {
		result = append(result, v)
	}
	return result
}

func IsSpecChanged(oldObj, newObj client.Object) bool {
	oldSpec, ok1 := extractSpec(oldObj)
	newSpec, ok2 := extractSpec(newObj)
	if !ok1 || !ok2 {
		return true
	}
	return !reflect.DeepEqual(oldSpec, newSpec)
}

func extractSpec(obj client.Object) (interface{}, bool) {
	switch o := obj.(type) {
	case *webv1alpha1.OpenResty:
		return o.Spec, true
	case *webv1alpha1.ServerBlock:
		return o.Spec, true
	case *webv1alpha1.Location:
		return o.Spec, true
	default:
		return nil, false
	}
}

func EqualSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
