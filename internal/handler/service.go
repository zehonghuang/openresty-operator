package handler

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/constants"
	"openresty-operator/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeployServerBlockServices(ctx context.Context, c client.Client, scheme *runtime.Scheme, app *webv1alpha1.OpenResty, log logr.Logger) error {
	for _, name := range app.Spec.Http.ServerRefs {
		var server webv1alpha1.ServerBlock
		if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &server); err != nil {
			return fmt.Errorf("failed to get ServerBlock %s: %w", name, err)
		}

		svc := generateServiceForServer(app, &server)

		if err := createOrUpdateService(ctx, c, scheme, app, svc, log); err != nil {
			return fmt.Errorf("failed to create or update Service for ServerBlock %s: %w", name, err)
		}
	}
	return nil
}

func generateServiceForServer(app *webv1alpha1.OpenResty, server *webv1alpha1.ServerBlock) *corev1.Service {
	port := utils.ParseListenPort(server.Spec.Listen)

	return &corev1.Service{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      server.Name,
			Namespace: app.Namespace,
			Labels:    constants.BuildCommonLabels(server, "service"),
		},
		Spec: corev1.ServiceSpec{
			Selector: constants.BuildSelectorLabels(app),
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       port,
					TargetPort: intstr.FromInt32(int32(port)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func createOrUpdateService(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, svc *corev1.Service, log logr.Logger) error {
	if err := ctrl.SetControllerReference(owner, svc, scheme); err != nil {
		return err
	}

	var existing corev1.Service
	err := c.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Service", "name", svc.Name)
			return c.Create(ctx, svc)
		}
		return err
	}

	svc.ResourceVersion = existing.ResourceVersion
	svc.Spec.ClusterIP = existing.Spec.ClusterIP
	svc.Spec.ClusterIPs = existing.Spec.ClusterIPs
	return c.Update(ctx, svc)
}
