/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// +kubebuilder:object:generate=true
// +groupName=core.dynamic.zagganas.com
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects.
	// This name is used by applyconfiguration generators (e.g. controller-gen).
	SchemeGroupVersion = schema.GroupVersion{Group: "core.dynamic.zagganas.com", Version: "v1alpha1"}

	// GroupVersion is an alias for SchemeGroupVersion, for backward compatibility.
	GroupVersion = SchemeGroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
		metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
		return nil
	})

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DynamicMessageJobSpec defines the desired state of DynamicMessageJob
type DynamicMessageJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// The command to run inside the container
	// +optional
	// +kubebuilder:default:="echo"
	Command *string `json:"command"`
	// The message to display
	// +required
	Message *string `json:"message"`
	// The number of times to retry (default 3)
	// +optional
	// +kubebuilder:default:=3
	// +kubebuilder:validation:Minimum:=1
	RetryLimit int `json:"retryLimit"`
	// The delay (in sec) between backoff retrials (default 30)
	// +optional
	// +kubebuilder:default:=30
	// +kubebuilder:validation:Minimum:=1
	RetryDelay int `json:"retryDelay"`
}

// DynamicMessageJobStatus defines the observed state of DynamicMessageJob.
type DynamicMessageJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the DynamicMessageJob resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// NextRun is the time of the next job run
	// +optional
	NextRun metav1.Time `json:"nextRunTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DynamicMessageJob is the Schema for the messagejobs API
type DynamicMessageJob struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of DynamicMessageJob
	// +required
	Spec DynamicMessageJobSpec `json:"spec"`

	// status defines the observed state of DynamicMessageJob
	// +optional
	Status DynamicMessageJobStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DynamicMessageJobList contains a list of DynamicMessageJob
type DynamicMessageJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []DynamicMessageJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &DynamicMessageJob{}, &DynamicMessageJobList{})
		return nil
	})
}
