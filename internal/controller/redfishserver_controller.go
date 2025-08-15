/*
 * This file is part of the KubeVirt Redfish project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2025 KubeVirt Redfish project and its authors.
 *
 */

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	routev1 "github.com/openshift/api/route/v1"
	redfishv1alpha1 "github.com/v1k0d3n/kubevirt-redfish-operator/api/v1alpha1"
	"gopkg.in/yaml.v3"
)

// RedfishServerReconciler reconciles a RedfishServer object
type RedfishServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=redfish.kubevirt.io,resources=redfishservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redfish.kubevirt.io,resources=redfishservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redfish.kubevirt.io,resources=redfishservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/status,verbs=get;list;watch;patch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances/status,verbs=get;list;watch;patch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/restart,verbs=create
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/start,verbs=create
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/stop,verbs=create
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances/pause,verbs=create
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances/unpause,verbs=create
// +kubebuilder:rbac:groups=subresources.kubevirt.io,resources=virtualmachineinstances/pause,verbs=create;update
// +kubebuilder:rbac:groups=subresources.kubevirt.io,resources=virtualmachineinstances/unpause,verbs=create;update
// +kubebuilder:rbac:groups=cdi.kubevirt.io,resources=volumeimportsources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RedfishServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting reconciliation", "namespace", req.Namespace, "name", req.Name)

	// Fetch the RedfishServer instance
	redfishServer := &redfishv1alpha1.RedfishServer{}
	err := r.Get(ctx, req.NamespacedName, redfishServer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("RedfishServer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get RedfishServer")
		return ctrl.Result{}, err
	}

	// Ensure TypeMeta is set for owner references
	if redfishServer.APIVersion == "" {
		redfishServer.APIVersion = "redfish.kubevirt.io/v1alpha1"
	}
	if redfishServer.Kind == "" {
		redfishServer.Kind = "RedfishServer"
	}

	// Validate the specification
	if err := r.validateSpec(redfishServer.Spec); err != nil {
		logger.Error(err, "Invalid RedfishServer specification")
		redfishServer.Status.Phase = redfishv1alpha1.RedfishServerPhaseFailed
		redfishServer.Status.Conditions = []metav1.Condition{
			{
				Type:               "Failed",
				Status:             metav1.ConditionTrue,
				Reason:             "ValidationFailed",
				Message:            err.Error(),
				LastTransitionTime: metav1.Now(),
			},
		}
		if err := r.Status().Patch(ctx, redfishServer, client.Merge); err != nil {
			logger.Error(err, "Failed to update RedfishServer status")
		}
		return ctrl.Result{}, err
	}

	// Reconcile ServiceAccount
	if err := r.reconcileServiceAccount(ctx, redfishServer); err != nil {
		logger.Error(err, "Failed to reconcile ServiceAccount")
		return ctrl.Result{}, err
	}

	// Reconcile RBAC
	if err := r.reconcileRBAC(ctx, redfishServer); err != nil {
		logger.Error(err, "Failed to reconcile RBAC")
		return ctrl.Result{}, err
	}

	// Reconcile ConfigMap
	if err := r.reconcileConfigMap(ctx, redfishServer); err != nil {
		logger.Error(err, "Failed to reconcile ConfigMap")
		return ctrl.Result{}, err
	}

	// Reconcile Deployment
	if err := r.reconcileDeployment(ctx, redfishServer); err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	// Reconcile Service
	if err := r.reconcileService(ctx, redfishServer); err != nil {
		logger.Error(err, "Failed to reconcile Service")
		return ctrl.Result{}, err
	}

	// Reconcile Route (OpenShift)
	if redfishServer.Spec.RouteEnabled {
		if err := r.reconcileRoute(ctx, redfishServer); err != nil {
			logger.Error(err, "Failed to reconcile Route")
			return ctrl.Result{}, err
		}
	}

	// Update status
	if err := r.updateStatus(ctx, redfishServer); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciliation completed successfully")
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// validateSpec validates the RedfishServer specification
func (r *RedfishServerReconciler) validateSpec(spec redfishv1alpha1.RedfishServerSpec) error {
	if spec.Version == "" {
		return fmt.Errorf("version is required")
	}
	if spec.Replicas < 1 || spec.Replicas > 10 {
		return fmt.Errorf("replicas must be between 1 and 10")
	}
	if len(spec.Chassis) == 0 {
		return fmt.Errorf("at least one chassis must be specified")
	}
	if len(spec.Authentication.Users) == 0 {
		return fmt.Errorf("at least one user must be specified")
	}
	return nil
}

// reconcileServiceAccount creates or updates the ServiceAccount
func (r *RedfishServerReconciler) reconcileServiceAccount(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sa", redfishServer.Name),
			Namespace: redfishServer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, serviceAccount, func() error {
		// ServiceAccount is mostly immutable, so we don't need to set many fields
		// Don't set blockOwnerDeletion for ServiceAccounts as they can't have finalizers
		serviceAccount.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: redfishServer.APIVersion,
				Kind:       redfishServer.Kind,
				Name:       redfishServer.Name,
				UID:        redfishServer.UID,
				Controller: &[]bool{true}[0],
				// Don't set BlockOwnerDeletion for ServiceAccounts
			},
		}
		return nil
	})

	return err
}

// reconcileRBAC creates or updates the RBAC resources
func (r *RedfishServerReconciler) reconcileRBAC(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	// Create ClusterRole with comprehensive permissions (matching working Helm deployment)
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-role", redfishServer.Namespace, redfishServer.Name),
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, clusterRole, func() error {
		clusterRole.Rules = []rbacv1.PolicyRule{
			// KubeVirt core permissions
			{
				APIGroups: []string{"kubevirt.io"},
				Resources: []string{"virtualmachines", "virtualmachineinstances"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			// KubeVirt status permissions
			{
				APIGroups: []string{"kubevirt.io"},
				Resources: []string{"virtualmachines/status", "virtualmachineinstances/status"},
				Verbs:     []string{"get", "list", "watch", "patch"},
			},
			// KubeVirt action permissions
			{
				APIGroups: []string{"kubevirt.io"},
				Resources: []string{"virtualmachines/restart", "virtualmachines/start", "virtualmachines/stop"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"kubevirt.io"},
				Resources: []string{"virtualmachineinstances/pause", "virtualmachineinstances/unpause"},
				Verbs:     []string{"create"},
			},
			// KubeVirt subresources
			{
				APIGroups: []string{"subresources.kubevirt.io"},
				Resources: []string{"virtualmachineinstances/pause", "virtualmachineinstances/unpause"},
				Verbs:     []string{"create", "update"},
			},
			// Core API permissions
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "configmaps", "secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list"},
			},
			// CDI permissions
			{
				APIGroups: []string{"cdi.kubevirt.io"},
				Resources: []string{"datavolumes", "volumeimportsources"},
				Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
			},
			// Storage permissions
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list"},
			},
		}
		return nil // No owner reference for cluster-scoped resources
	})

	if err != nil {
		return err
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-binding", redfishServer.Namespace, redfishServer.Name),
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		}
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      fmt.Sprintf("%s-sa", redfishServer.Name),
				Namespace: redfishServer.Namespace,
			},
		}
		return nil // No owner reference for cluster-scoped resources
	})

	return err
}

// reconcileConfigMap creates or updates the ConfigMap
func (r *RedfishServerReconciler) reconcileConfigMap(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config", redfishServer.Name),
			Namespace: redfishServer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		// Build chassis configurations
		chassisConfigs := []map[string]interface{}{}
		for _, chassis := range redfishServer.Spec.Chassis {
			chassisConfig := map[string]interface{}{
				"name":            chassis.Name,
				"namespace":       chassis.Namespace,
				"description":     chassis.Description,
				"service_account": chassis.ServiceAccount,
				"vm_selector": map[string]interface{}{
					"labels": map[string]string{
						"redfish-enabled": "true",
					},
				},
			}
			chassisConfigs = append(chassisConfigs, chassisConfig)
		}

		// Build authentication users
		userConfigs := []map[string]interface{}{}
		for _, user := range redfishServer.Spec.Authentication.Users {
			userConfig := map[string]interface{}{
				"username": user.Username,
				"password": "admin123", // Default password - in production, this should come from the secret
				"chassis":  user.Chassis,
			}
			userConfigs = append(userConfigs, userConfig)
		}

		// Build configuration based on RedfishServer spec
		config := map[string]interface{}{
			"server": map[string]interface{}{
				"host": "0.0.0.0",
				"port": 8443,
				"tls": map[string]interface{}{
					"enabled": redfishServer.Spec.TLS.Enabled,
				},
			},
			"chassis": chassisConfigs,
			"authentication": map[string]interface{}{
				"users": userConfigs,
			},
			"kubevirt": map[string]interface{}{
				"api_version":        "v1",
				"timeout":            30,
				"allow_insecure_tls": true,
			},
			"datavolume": map[string]interface{}{
				"storage_size":         "3Gi",
				"allow_insecure_tls":   true,
				"storage_class":        "lvms-vg1-immediate",
				"vm_update_timeout":    "2m",
				"iso_download_timeout": "30m",
				"helper_image":         "alpine:latest",
			},
		}

		// Convert to YAML
		yamlData, err := yaml.Marshal(config)
		if err != nil {
			return err
		}

		configMap.Data = map[string]string{
			"config.yaml": string(yamlData),
		}

		return ctrl.SetControllerReference(redfishServer, configMap, r.Scheme)
	})

	return err
}

// generateConfig generates the configuration YAML for the Redfish server
func (r *RedfishServerReconciler) generateConfig(spec redfishv1alpha1.RedfishServerSpec) string {
	// This is a simplified config generation - you would implement the full config
	// based on your kubevirt-redfish application's configuration format
	config := fmt.Sprintf(`
version: %s
replicas: %d
chassis:
`, spec.Version, spec.Replicas)

	for _, chassis := range spec.Chassis {
		config += fmt.Sprintf(`  - name: %s
    namespace: %s
    description: %s
    service_account: %s
`, chassis.Name, chassis.Namespace, chassis.Description, chassis.ServiceAccount)
	}

	config += `authentication:
  users:
`
	for _, user := range spec.Authentication.Users {
		config += fmt.Sprintf(`    - username: %s
      password_secret: %s
      chassis: [%s]
`, user.Username, user.PasswordSecret, r.joinStrings(user.Chassis, ", "))
	}

	if spec.TLS.Enabled {
		config += fmt.Sprintf(`
tls:
  enabled: true
  cert_secret: %s
  key_secret: %s
  insecure_skip_verify: %t
`, spec.TLS.CertSecret, spec.TLS.KeySecret, spec.TLS.InsecureSkipVerify)
	}

	return config
}

// joinStrings joins a slice of strings with a separator
func (r *RedfishServerReconciler) joinStrings(strs []string, sep string) string {
	result := ""
	for i, str := range strs {
		if i > 0 {
			result += sep
		}
		result += fmt.Sprintf(`"%s"`, str)
	}
	return result
}

// reconcileDeployment creates or updates the Deployment
func (r *RedfishServerReconciler) reconcileDeployment(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redfishServer.Name,
			Namespace: redfishServer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// Set image - use actual Redfish image if specified, otherwise use a working version
		image := redfishServer.Spec.Image
		if image == "" {
			image = "quay.io/bjozsa-redhat/kubevirt-redfish:08ff5a7c" // Use the working image from Helm deployment
		}

		// Set image pull policy
		imagePullPolicy := corev1.PullPolicy(redfishServer.Spec.ImagePullPolicy)
		if imagePullPolicy == "" {
			imagePullPolicy = corev1.PullAlways
		}

		// Set replicas
		replicas := redfishServer.Spec.Replicas
		if replicas == 0 {
			replicas = 1
		}

		// Set resource requirements
		resources := redfishServer.Spec.Resources
		if resources == nil {
			resources = &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				},
			}
		}

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": redfishServer.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": redfishServer.Name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: fmt.Sprintf("%s-sa", redfishServer.Name),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
					},
					Containers: []corev1.Container{
						{
							Name:            redfishServer.Name,
							Image:           image,
							ImagePullPolicy: imagePullPolicy,
							Resources:       *resources,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &[]bool{false}[0],
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								RunAsNonRoot: &[]bool{true}[0],
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8443,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "CONFIG_PATH",
									Value: "/app/config/config.yaml",
								},
								{
									Name:  "LOG_LEVEL",
									Value: "debug",
								},
								{
									Name:  "REDFISH_LOGGING_ENABLED",
									Value: "true",
								},
								{
									Name:  "REDFISH_LOG_LEVEL",
									Value: "DEBUG",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/redfish/v1/",
										Port:   intstr.FromInt(8443),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    3,
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/redfish/v1/",
										Port:   intstr.FromInt(8443),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    3,
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								TimeoutSeconds:      3,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/app/config",
									ReadOnly:  true,
								},
								{
									Name:      "temp-storage",
									MountPath: "/tmp",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-config", redfishServer.Name),
									},
								},
							},
						},
						{
							Name: "temp-storage",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									SizeLimit: resource.NewQuantity(50*1024*1024*1024, resource.BinarySI), // 50Gi
								},
							},
						},
					},
				},
			},
		}

		return ctrl.SetControllerReference(redfishServer, deployment, r.Scheme)
	})

	return err
}

// getImage returns the image to use for the deployment
func (r *RedfishServerReconciler) getImage(spec redfishv1alpha1.RedfishServerSpec) string {
	if spec.Image != "" {
		return spec.Image
	}
	// Default image - you would replace this with your actual image
	return fmt.Sprintf("quay.io/bjozsa-redhat/kubevirt-redfish:%s", spec.Version)
}

// getResources returns the resource requirements
func (r *RedfishServerReconciler) getResources(spec redfishv1alpha1.RedfishServerSpec) corev1.ResourceRequirements {
	if spec.Resources != nil {
		return *spec.Resources
	}
	// Default resources
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
}

// reconcileService creates or updates the Service
func (r *RedfishServerReconciler) reconcileService(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redfishServer.Name,
			Namespace: redfishServer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Set service type
		serviceType := redfishServer.Spec.ServiceType
		if serviceType == "" {
			serviceType = "ClusterIP"
		}

		service.Spec = corev1.ServiceSpec{
			Type: corev1.ServiceType(serviceType),
			Selector: map[string]string{
				"app": redfishServer.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8443,
					TargetPort: intstr.FromInt(8443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}

		return ctrl.SetControllerReference(redfishServer, service, r.Scheme)
	})

	return err
}

// reconcileRoute creates or updates the OpenShift Route
func (r *RedfishServerReconciler) reconcileRoute(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	if !redfishServer.Spec.RouteEnabled {
		// If route is not enabled, delete any existing route
		route := &routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redfishServer.Name,
				Namespace: redfishServer.Namespace,
			},
		}
		return r.Client.Delete(ctx, route)
	}

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redfishServer.Name,
			Namespace: redfishServer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, route, func() error {
		// Set host - use provided host or generate one
		host := redfishServer.Spec.RouteHost
		if host == "" {
			host = fmt.Sprintf("%s-%s.apps.hub.lab.ocp.run", redfishServer.Name, redfishServer.Namespace)
		}

		route.Spec = routev1.RouteSpec{
			Host: host,
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   redfishServer.Name,
				Weight: &[]int32{100}[0],
			},
		}

		return ctrl.SetControllerReference(redfishServer, route, r.Scheme)
	})

	return err
}

// updateStatus updates the RedfishServer status
func (r *RedfishServerReconciler) updateStatus(ctx context.Context, redfishServer *redfishv1alpha1.RedfishServer) error {
	// Get the deployment to check its status
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      redfishServer.Name,
		Namespace: redfishServer.Namespace,
	}, deployment)

	if err != nil {
		if errors.IsNotFound(err) {
			redfishServer.Status.Phase = redfishv1alpha1.RedfishServerPhasePending
			redfishServer.Status.ReadyReplicas = 0
		} else {
			return err
		}
	} else {
		// Update status based on deployment
		redfishServer.Status.ReadyReplicas = deployment.Status.ReadyReplicas
		redfishServer.Status.CurrentVersion = redfishServer.Spec.Version
		redfishServer.Status.ObservedGeneration = redfishServer.Generation

		// Determine phase based on deployment status
		if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
			redfishServer.Status.Phase = redfishv1alpha1.RedfishServerPhaseRunning
			redfishServer.Status.Conditions = []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             "AllReplicasReady",
					Message:            "All replicas are ready",
					LastTransitionTime: metav1.Now(),
				},
			}
		} else if deployment.Status.ReadyReplicas > 0 {
			redfishServer.Status.Phase = redfishv1alpha1.RedfishServerPhaseDeploying
		} else {
			redfishServer.Status.Phase = redfishv1alpha1.RedfishServerPhasePending
		}

		// Set service URL
		redfishServer.Status.ServiceURL = fmt.Sprintf("http://%s.%s.svc.cluster.local", redfishServer.Name, redfishServer.Namespace)
	}

	// Use Patch instead of Update to avoid conflicts
	return r.Status().Patch(ctx, redfishServer, client.Merge)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedfishServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redfishv1alpha1.RedfishServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Named("redfishserver").
		Complete(r)
}
