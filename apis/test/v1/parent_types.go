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
type ParentPhase string

const (
	ParentPhaseCreating ParentPhase = "Creating"
	ParentPhaseRunning  ParentPhase = "Running"
	ParentPhaseUpdating ParentPhase = "Updating"
	ParentPhaseDeleting ParentPhase = "Deleting"
)

// ParentSpec defines the desired state of Parent
type ParentSpec struct {
}

// ParentStatus defines the observed state of Parent
type ParentStatus struct {
	Phase ParentPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Parent is the Schema for the parents API
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Parent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ParentSpec   `json:"spec,omitempty"`
	Status ParentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ParentList contains a list of Parent
type ParentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Parent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Parent{}, &ParentList{})
}
