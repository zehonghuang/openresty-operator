package utils

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
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

func SplitHostPort(input string) (string, string, error) {
	// 处理带 http/https schema 的 URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return "", "", fmt.Errorf("invalid URL: %v", err)
		}

		host := u.Hostname()
		port := u.Port()
		if port == "" {
			if u.Scheme == "http" {
				port = "80"
			} else if u.Scheme == "https" {
				port = "443"
			}
		}
		return host, port, nil
	}

	// 处理 host:port 的格式
	if strings.Contains(input, ":") {
		host, port, err := net.SplitHostPort(input)
		if err == nil {
			return host, port, nil
		}
		// 可能是域名中带冒号但格式不合法，比如 IPv6 缺 []
		return "", "", fmt.Errorf("invalid host:port format: %v", err)
	}

	// fallback，只有 host 没有端口
	return input, "80", nil
}

func DeepEqual[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}
	return true
}

func DeepEqualMapStringByteSlice(a, b map[string][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !bytes.Equal(v, bv) {
			return false
		}
	}
	return true
}

func MergeMaps(dst, src map[string]string) map[string]string {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func MapValuesNonNil[T any](m map[string]*T) []*T {
	values := make([]*T, 0, len(m))
	for _, v := range m {
		if v != nil {
			values = append(values, v)
		}
	}
	return values
}

func MapList[T any, R any](list []T, f func(T) R) []R {
	out := make([]R, 0, len(list))
	for _, item := range list {
		out = append(out, f(item))
	}
	return out
}
