/*
Copyright 2021.

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

// +kubebuilder:validation:Enum=Creating;Running;Updating;Deleting
// +kubebuilder:default=Creating
type ChildPhase string

const (
	ChildPhaseCreating ChildPhase = "Creating"
	ChildPhaseRunning  ChildPhase = "Running"
	ChildPhaseUpdating ChildPhase = "Updating"
	ChildPhaseDeleting ChildPhase = "Deleting"
)

// ChildSpec defines the desired state of Child
type ChildSpec struct {
}

// ChildStatus defines the observed state of Child
type ChildStatus struct {
	Phase ChildPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Child is the Schema for the children API
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Child struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChildSpec   `json:"spec,omitempty"`
	Status ChildStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChildList contains a list of Child
type ChildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Child `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Child{}, &ChildList{})
}
