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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JobOperatorSpec defines the desired state of JobOperator.
type JobOperatorSpec struct {
	// Replicas is the number of replicas to deploy
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Replicas *int32 `json:"replicas,omitempty"`

	// Image is the container image to deploy
	Image string `json:"image"`

	// Port is the port the container listens on
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`
}

// JobOperatorStatus defines the observed state of JobOperator.
type JobOperatorStatus struct {
	// Ready replicas count
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total replicas count
	Replicas int32 `json:"replicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=jo

// JobOperator is the Schema for the joboperators API.
type JobOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobOperatorSpec   `json:"spec,omitempty"`
	Status JobOperatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JobOperatorList contains a list of JobOperator.
type JobOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JobOperator{}, &JobOperatorList{})
}
