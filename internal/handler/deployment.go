package handler

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/constants"
	"openresty-operator/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func DeployOpenRestyPod(ctx context.Context, c client.Client, scheme *runtime.Scheme, app *webv1alpha1.OpenResty, upstreamsType map[string]webv1alpha1.UpstreamType, log logr.Logger) (error, *appsv1.Deployment) {
	vmResult, err := BuildVolumesAndMounts(ctx, c, app, upstreamsType)
	if err != nil {
		return err, nil
	}

	defaulted := &appsv1.Deployment{}
	scheme.Default(defaulted)

	deployment := BuildDeploymentSpec(app, defaulted, vmResult.Volumes, vmResult.Mounts, vmResult.MetricsPort)

	return CreateOrUpdateDeployment(ctx, c, scheme, app, deployment, log)
}

func CreateOrUpdateDeployment(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	app *webv1alpha1.OpenResty,
	deployment *appsv1.Deployment,
	log logr.Logger,
) (error, *appsv1.Deployment) {
	existing := &appsv1.Deployment{}
	err := c.Get(ctx, types.NamespacedName{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}, existing)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Deployment", "name", deployment.Name)
			if err := ctrl.SetControllerReference(app, deployment, scheme); err != nil {
				return err, nil
			}
			return c.Create(ctx, deployment), deployment
		}
		return fmt.Errorf("failed to get Deployment: %w", err), nil
	}

	if !cmp.Equal(existing.Spec, deployment.Spec) {
		diff := cmp.Diff(existing.Spec, deployment.Spec)
		log.V(4).Info("Deployment spec changed", "diff", diff)

		existing.Spec = deployment.Spec

		if err := ctrl.SetControllerReference(app, existing, scheme); err != nil {
			return err, nil
		}
		log.Info("Updating Deployment", "name", existing.Name)
		return c.Update(ctx, existing), existing
	}

	log.V(4).Info("Deployment up-to-date", "name", existing.Name)
	return nil, existing
}

type VolumeMountResult struct {
	Volumes     []corev1.Volume
	Mounts      []corev1.VolumeMount
	MetricsPort *corev1.ContainerPort
}

func BuildVolumesAndMounts(ctx context.Context, c client.Client, app *webv1alpha1.OpenResty, upstreamTypes map[string]webv1alpha1.UpstreamType) (*VolumeMountResult, error) {
	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount
	locationSeen := map[string]bool{}

	// --- Mount main nginx.conf ---
	volumes = append(volumes, corev1.Volume{
		Name: "main-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "openresty-" + app.Name + "-main",
				},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "main-config",
		MountPath: utils.NginxConfPath,
		SubPath:   "nginx.conf",
	})

	// --- Mount ServerBlock & Location ---
	for _, serverName := range app.Spec.Http.ServerRefs {
		// mount serverblock ConfigMap
		volumes = append(volumes, corev1.Volume{
			Name: "serverblock-" + serverName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "serverblock-" + serverName,
					},
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "serverblock-" + serverName,
			MountPath: utils.NginxServerConfigDir + "/" + serverName,
		})

		// fetch ServerBlock to get LocationRefs
		var server webv1alpha1.ServerBlock
		if err := c.Get(ctx, types.NamespacedName{Name: serverName, Namespace: app.Namespace}, &server); err != nil {
			return nil, err
		}

		for _, locName := range server.Spec.LocationRefs {
			if locationSeen[locName] {
				continue
			}
			locationSeen[locName] = true

			volumes = append(volumes, corev1.Volume{
				Name: "location-" + locName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "location-" + locName,
						},
					},
				},
			})
			mounts = append(mounts, corev1.VolumeMount{
				Name:      "location-" + locName,
				MountPath: utils.NginxLocationConfigDir + "/" + locName,
			})
		}
	}

	// --- Mount Upstream ---
	for _, upstreamName := range app.Spec.Http.UpstreamRefs {
		volumes = append(volumes, corev1.Volume{
			Name: "upstream-" + upstreamName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "upstream-" + upstreamName,
					},
				},
			},
		})

		path := ""
		if upstreamTypes[upstreamName] == webv1alpha1.UpstreamTypeAddress {
			path = utils.NginxUpstreamConfigDir + "/" + upstreamName
		} else {
			path = utils.NginxLuaLibUpstreamDir + "/" + upstreamName
		}

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "upstream-" + upstreamName,
			MountPath: path,
		})
	}

	var secretList corev1.SecretList
	if err := c.List(ctx, &secretList, client.InNamespace(app.Namespace),
		client.MatchingLabels{
			constants.LabelComponent: "secret",
			constants.LabelManagedBy: "openresty-operator",
		}); err != nil {
		return nil, err
	}
	for _, secret := range secretList.Items {
		volumes = append(volumes, corev1.Volume{
			Name: "secret-" + secret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		})

		locName := secret.Annotations[constants.AnnotationSecretHeaders]
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "secret-" + secret.Name,
			MountPath: utils.NginxLuaLibSecretDir + "/" + locName,
		})
	}

	// --- Mount Logs ---
	if app.Spec.LogVolume.Type == webv1alpha1.LogVolumeTypeEmptyDir {
		volumes = append(volumes, corev1.Volume{
			Name: "nginx-logs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

	} else {
		volumes = append(volumes, corev1.Volume{
			Name: "nginx-logs",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: app.Spec.LogVolume.PersistentVolumeClaim,
				},
			},
		})
	}
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "nginx-logs",
		MountPath: utils.NginxLogDir,
	})

	// --- Metrics Port (optional) ---
	var metricsPort *corev1.ContainerPort
	if app.Spec.MetricsServer != nil && app.Spec.MetricsServer.Enable {
		port := "8080"
		if app.Spec.MetricsServer.Listen != "" {
			port = app.Spec.MetricsServer.Listen
		}
		metricsPort = &corev1.ContainerPort{
			Name:          "metrics",
			ContainerPort: utils.ParseListenPort(port),
			Protocol:      corev1.ProtocolTCP,
		}
	}

	return &VolumeMountResult{
		Volumes:     volumes,
		Mounts:      mounts,
		MetricsPort: metricsPort,
	}, nil
}

func BuildDeploymentSpec(app *webv1alpha1.OpenResty, defaulted *appsv1.Deployment, volumes []corev1.Volume, mounts []corev1.VolumeMount, metricsPort *corev1.ContainerPort) *appsv1.Deployment {
	name := "openresty-" + app.Name
	replicas := int32(1)
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: app.Namespace,
			Labels:    constants.BuildCommonLabels(app, "deployment"),
		},
		Spec: defaulted.Spec,
	}

	dep.Spec.Replicas = ptr.To(replicas)
	dep.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: constants.BuildSelectorLabels(app),
	}
	dep.Spec.Template.ObjectMeta.Labels = constants.BuildCommonLabels(app, "pod")
	// dep.Spec.Template.Annotations = buildPrometheusAnnotations(app.Spec.MetricsServer)

	if v, ok := app.Annotations[constants.AnnotationTriggerHash]; ok {
		if dep.Spec.Template.Annotations == nil {
			dep.Spec.Template.Annotations = make(map[string]string)
		}
		dep.Spec.Template.Annotations[constants.AnnotationTriggerHash] = v
	}

	dep.Spec.Template.Spec.Volumes = volumes
	dep.Spec.Template.Spec.ShareProcessNamespace = ptr.To(true)
	dep.Spec.Template.Spec.NodeSelector = app.Spec.NodeSelector
	dep.Spec.Template.Spec.Affinity = app.Spec.Affinity
	dep.Spec.Template.Spec.Tolerations = app.Spec.Tolerations
	dep.Spec.Template.Spec.TerminationGracePeriodSeconds = app.Spec.TerminationGracePeriodSeconds
	dep.Spec.Template.Spec.PriorityClassName = app.Spec.PriorityClassName

	// 注入 containers
	if len(app.Spec.Image) == 0 {
		app.Spec.Image = "gintonic1glass/openresty:alpine-1.1.9"
	}
	openrestyContainer := corev1.Container{
		Name:      "openresty",
		Image:     app.Spec.Image,
		Resources: app.Spec.Resources,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 80,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		VolumeMounts: mounts,
	}

	if metricsPort != nil {
		openrestyContainer.Ports = append(openrestyContainer.Ports, *metricsPort)
	}

	reloadAgentContainer := corev1.Container{
		Name:  "reload-agent",
		Image: "gintonic1glass/reload-agent:v0.1.6",
		Ports: []corev1.ContainerPort{
			{
				Name:          "reload-metrics",
				ContainerPort: 19091,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env:          app.Spec.ReloadAgentEnv,
		VolumeMounts: mounts[1:], // reload agent 不挂主 nginx.conf
	}

	dep.Spec.Template.Spec.Containers = []corev1.Container{
		openrestyContainer,
		reloadAgentContainer,
	}

	return dep
}

func buildPrometheusAnnotations(metrics *webv1alpha1.MetricsServer) map[string]string {
	if metrics == nil || !metrics.Enable {
		return map[string]string{}
	}
	port := defaultOr(metrics.Listen, "9091")
	path := defaultOr(metrics.Path, "/metrics")

	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   port,
		"prometheus.io/path":   path,
	}
}

func CreateOrUpdateServiceMonitor(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, labels, annotations map[string]string, log logr.Logger) error {

	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:        owner.GetName(),
			Namespace:   owner.GetNamespace(),
			Labels:      utils.MergeMaps(constants.BuildCommonLabels(owner, "service-monitor"), labels),
			Annotations: annotations,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: constants.BuildSelectorLabels(owner),
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{owner.GetNamespace()},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "metrics",
					Path:     "/metrics",
					Interval: "30s",
				},
				{
					Port:     "reload-metrics",
					Path:     "/metrics",
					Interval: "30s",
				},
			},
		},
	}

	var existing monitoringv1.ServiceMonitor
	err := c.Get(ctx, types.NamespacedName{Name: owner.GetName(), Namespace: owner.GetNamespace()}, &existing)
	if errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(owner, sm, scheme); err != nil {
			return err
		}
		return c.Create(ctx, sm)
	} else if err != nil {
		return err
	}

	if !cmp.Equal(existing.Spec, sm.Spec) {
		diff := cmp.Diff(existing.Spec, sm.Spec)
		log.V(4).Info("ServiceMonitor spec changed", "diff", diff)

		existing.Spec = sm.Spec

		if err := ctrl.SetControllerReference(sm, &existing, scheme); err != nil {
			return err
		}
		log.Info("Updating ServiceMonitor", "name", existing.Name)
		return c.Update(ctx, &existing)
	}
	return nil
}

func CreateOrUpdateMetricsService(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	app *webv1alpha1.OpenResty,
) error {
	name := app.Name + "-metrics"

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: app.Namespace,
			Labels:    constants.BuildCommonLabels(app, "service"),
		},
		Spec: corev1.ServiceSpec{
			Selector: constants.BuildSelectorLabels(app),
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       9090,
					TargetPort: intstr.Parse(app.Spec.MetricsServer.Listen),
				},
				{
					Name:       "reload-metrics",
					Port:       19091,
					TargetPort: intstr.FromInt32(19091),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(app, svc, scheme); err != nil {
		return err
	}

	var existing corev1.Service
	err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &existing)
	if errors.IsNotFound(err) {
		return c.Create(ctx, svc)
	} else if err != nil {
		return err
	}

	// Apply changes to existing Service
	existing.Spec.Selector = svc.Spec.Selector
	existing.Spec.Ports = svc.Spec.Ports
	return c.Update(ctx, &existing)
}
