package handler

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdateConfigMap(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	owner client.Object,
	name string,
	namespace string,
	labels map[string]string,
	data map[string]string,
	log logr.Logger,
) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}

	if err := ctrl.SetControllerReference(owner, cm, scheme); err != nil {
		return err
	}

	var existing corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap", "name", name)
			return c.Create(ctx, cm)
		}
		return err
	}

	if !deepEqual(existing.Data, cm.Data) {
		log.Info("Updating ConfigMap", "name", name)
		existing.Data = cm.Data
		return c.Update(ctx, &existing)
	}

	return nil
}

func deepEqual(a, b map[string]string) bool {
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
