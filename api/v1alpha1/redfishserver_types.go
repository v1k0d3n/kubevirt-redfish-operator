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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RedfishServerPhase represents the current phase of the RedfishServer
type RedfishServerPhase string

const (
	// RedfishServerPhasePending indicates the RedfishServer is pending deployment
	RedfishServerPhasePending RedfishServerPhase = "Pending"
	// RedfishServerPhaseDeploying indicates the RedfishServer is being deployed
	RedfishServerPhaseDeploying RedfishServerPhase = "Deploying"
	// RedfishServerPhaseRunning indicates the RedfishServer is running successfully
	RedfishServerPhaseRunning RedfishServerPhase = "Running"
	// RedfishServerPhaseFailed indicates the RedfishServer has failed
	RedfishServerPhaseFailed RedfishServerPhase = "Failed"
	// RedfishServerPhaseUpdating indicates the RedfishServer is being updated
	RedfishServerPhaseUpdating RedfishServerPhase = "Updating"
)

// RedfishChassisSpec defines the configuration for a Redfish chassis
type RedfishChassisSpec struct {
	// Name is the unique identifier for the chassis
	Name string `json:"name"`
	// Namespace is the Kubernetes namespace where the chassis operates
	Namespace string `json:"namespace"`
	// Description provides a human-readable description of the chassis
	Description string `json:"description,omitempty"`
	// ServiceAccount is the service account to use for this chassis
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// VMSelector defines the labels to select VMs for this chassis
	VMSelector map[string]string `json:"vmSelector,omitempty"`
}

// RedfishUserSpec defines the configuration for a Redfish user
type RedfishUserSpec struct {
	// Username is the unique username for authentication
	Username string `json:"username"`
	// PasswordSecret is the name of the secret containing the user's password
	PasswordSecret string `json:"passwordSecret"`
	// Chassis is a list of chassis names this user has access to
	Chassis []string `json:"chassis"`
}

// RedfishAuthSpec defines the authentication configuration
type RedfishAuthSpec struct {
	// Users is a list of users with their authentication details
	Users []RedfishUserSpec `json:"users"`
}

// RedfishTLSSpec defines the TLS configuration
type RedfishTLSSpec struct {
	// Enabled indicates whether TLS is enabled
	Enabled bool `json:"enabled,omitempty"`
	// CertSecret is the name of the secret containing the TLS certificate
	CertSecret string `json:"certSecret,omitempty"`
	// KeySecret is the name of the secret containing the TLS private key
	KeySecret string `json:"keySecret,omitempty"`
	// InsecureSkipVerify allows insecure TLS connections (for development)
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// RedfishMonitoringSpec defines the monitoring configuration
type RedfishMonitoringSpec struct {
	// Enabled indicates whether monitoring is enabled
	Enabled bool `json:"enabled,omitempty"`
	// ServiceMonitor creates a Prometheus ServiceMonitor
	ServiceMonitor bool `json:"serviceMonitor,omitempty"`
	// MetricsPort is the port for metrics endpoint
	MetricsPort int32 `json:"metricsPort,omitempty"`
}

// DataVolumeSpec defines DataVolume configuration for virtual media operations
type DataVolumeSpec struct {
	// Storage size for ISO images
	StorageSize string `json:"storageSize,omitempty"`
	// Allow insecure TLS for ISO downloads
	AllowInsecureTLS bool `json:"allowInsecureTLS,omitempty"`
	// Storage class for PVC creation
	StorageClass string `json:"storageClass,omitempty"`
	// Timeout for VM update operations
	VMUpdateTimeout string `json:"vmUpdateTimeout,omitempty"`
	// Timeout for ISO download operations
	ISODownloadTimeout string `json:"isoDownloadTimeout,omitempty"`
	// Helper image for ISO copy operations
	HelperImage string `json:"helperImage,omitempty"`
}

// VirtualMediaSpec defines virtual media configuration
type VirtualMediaSpec struct {
	// DataVolume configuration for ISO operations
	DataVolume *DataVolumeSpec `json:"datavolume,omitempty"`
}

// RedfishServerSpec defines the desired state of RedfishServer
type RedfishServerSpec struct {
	// Version specifies the version of the Redfish server to deploy
	Version string `json:"version"`
	// Replicas specifies the number of replicas to run
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Replicas int32 `json:"replicas,omitempty"`
	// Chassis defines the chassis configurations
	Chassis []RedfishChassisSpec `json:"chassis"`
	// Authentication defines the authentication configuration
	Authentication RedfishAuthSpec `json:"authentication"`
	// TLS defines the TLS configuration
	TLS RedfishTLSSpec `json:"tls,omitempty"`
	// Monitoring defines the monitoring configuration
	Monitoring RedfishMonitoringSpec `json:"monitoring,omitempty"`
	// VirtualMedia defines virtual media configuration
	VirtualMedia VirtualMediaSpec `json:"virtualMedia,omitempty"`
	// Image specifies the container image to use
	Image string `json:"image,omitempty"`
	// ImagePullPolicy specifies the image pull policy
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
	// Resources specifies the resource requirements
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// ServiceType specifies the type of service to create
	ServiceType string `json:"serviceType,omitempty"`
	// RouteEnabled enables OpenShift route creation
	RouteEnabled bool `json:"routeEnabled,omitempty"`
	// RouteHost specifies the host for the OpenShift route
	RouteHost string `json:"routeHost,omitempty"`
}

// RedfishServerStatus defines the observed state of RedfishServer
type RedfishServerStatus struct {
	// Conditions represents the latest available observations of a RedfishServer's current state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Phase represents the current phase of the RedfishServer
	Phase RedfishServerPhase `json:"phase,omitempty"`
	// ReadyReplicas is the number of ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// CurrentVersion is the currently deployed version
	CurrentVersion string `json:"currentVersion,omitempty"`
	// ServiceURL is the URL where the service is accessible
	ServiceURL string `json:"serviceURL,omitempty"`
	// RouteURL is the URL where the route is accessible (OpenShift)
	RouteURL string `json:"routeURL,omitempty"`
	// ObservedGeneration is the most recent generation observed by the controller
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.currentVersion"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// RedfishServer is the Schema for the redfishservers API
type RedfishServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedfishServerSpec   `json:"spec,omitempty"`
	Status RedfishServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedfishServerList contains a list of RedfishServer
type RedfishServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedfishServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedfishServer{}, &RedfishServerList{})
}
