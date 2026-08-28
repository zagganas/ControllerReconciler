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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// StateSpec defines the desired state of State
type StateSpec struct {
	// Add validation
	// +required
	Group *string `json:"group"`
	// +required
	Version *string `json:"version"`
	// +required
	Kind *string `json:"kind"`
	// Denotes whether the controller is active or not
	// +optional
	// +kubebuilder:default:=false
	Active bool `json:"active,omitzero"`
	// Hash defines the hash of the previous version
	// If this has changed, then the operator needs to restart
	// +optional
	// +kubebuilder:default=""
	Hash string `json:"hash,omitzero"`
}

// StateStatus defines the observed state of State.
type StateStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// LastTransitionTime records the time (in Unix timestamp)
	// that the state was last updated.
	// It is used to see if the state is stale after operator restart
	// +optional
	// +kubebuilder:default=0
	LastTransitionTime int64 `json:"lastTransitionTime,omitzero"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// State is the Schema for the controllers API
type State struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of State
	// +required
	Spec StateSpec `json:"spec"`

	// status defines the observed state of State
	// +optional
	Status StateStatus `json:"status"`
}

// +kubebuilder:object:root=true

// StateList contains a list of State
type StateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []State `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &State{}, &StateList{})
		return nil
	})
}
