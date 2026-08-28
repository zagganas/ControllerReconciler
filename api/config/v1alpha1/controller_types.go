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
// +groupName=config.zagganas.com
package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ControllerComponentType struct {
	// ApiVersion, Kind: https://pkg.go.dev/go.f110.dev/kubeproto/go/apis/metav1#TypeMeta
	metav1.TypeMeta `json:",inline"`
	// Name, Namespace, etc: https://pkg.go.dev/go.f110.dev/kubeproto/go/apis/metav1#ObjectMeta
	Metadata metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec string `json:"spec"`
}

// ControllerSpec defines the desired state of Controller
type ControllerSpec struct {
	GVK metav1.GroupVersionKind `json:"gvk"`

	// The template of the composite resource
	// +kubebuilder:validation:items:XEmbeddedResource
	Components []ControllerComponentType `json:"components"`
}

// ControllerStatus defines the observed state of Controller.
type ControllerStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Controller resource.
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
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Controller is the Schema for the controllers API
type Controller struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Controller
	// +required
	Spec ControllerSpec `json:"spec"`

	// status defines the observed state of Controller
	// +optional
	Status ControllerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ControllerList contains a list of Controller
type ControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Controller `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Controller{}, &ControllerList{})
		return nil
	})
}

func (c Controller) GetSpecBytes() ([]byte, error) {

	specStr, err := json.Marshal(c.Spec)
	if err != nil {
		return []byte(""), err
	}

	return specStr, nil
}
