package handler

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"openresty-operator/internal/utils"
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
	opts func(*metav1.OwnerReference),
	keysToDelete []string,
) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}
	if opts == nil {
		opts = func(reference *metav1.OwnerReference) {}
	}
	if err := ctrl.SetControllerReference(owner, cm, scheme, opts); err != nil {
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

	for _, key := range keysToDelete {
		delete(existing.Data, key)
		log.Info("Deleting key from ConfigMap", "key", key)
	}

	for k, v := range cm.Data {
		existing.Data[k] = v
	}

	if !utils.DeepEqual(existing.Data, cm.Data) {
		log.Info("Updating ConfigMap", "name", name)
		return c.Update(ctx, &existing)
	}

	return nil
}
