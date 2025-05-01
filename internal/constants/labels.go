package constants

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Prefix = "openresty.huangzehong.me"

	LabelAppName   = "app.kubernetes.io/name"
	LabelInstance  = "app.kubernetes.io/instance"
	LabelManagedBy = "app.kubernetes.io/managed-by"
	LabelComponent = "app.kubernetes.io/component"

	LabelOwnedCR = Prefix + "/cr"
)

func BuildCommonLabels(owner client.Object, component string) map[string]string {
	return map[string]string{
		LabelAppName:   "openresty",
		LabelInstance:  owner.GetName(),
		LabelManagedBy: "openresty-operator",
		LabelComponent: component,
		LabelOwnedCR:   owner.GetName(),
	}
}

func BuildSelectorLabels(owner client.Object) map[string]string {
	return map[string]string{
		LabelAppName:  "openresty",
		LabelInstance: owner.GetName(),
	}
}

func BuildCRLabels(obj client.Object) map[string]string {
	return map[string]string{
		LabelAppName:   "openresty",
		LabelInstance:  obj.GetName(),
		LabelManagedBy: "openresty-operator",
		LabelComponent: obj.GetObjectKind().GroupVersionKind().Kind,
	}
}
