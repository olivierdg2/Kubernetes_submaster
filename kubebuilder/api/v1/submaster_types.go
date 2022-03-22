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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SubmasterSpec defines the desired state of Submaster
type SubmasterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Containerized bool   `json:"existing,omitempty"`
	IP string              `json:"ip,omitempty"`
	Config string          `json:"config,omitempty"`
}

// SubmasterStatus defines the observed state of Submaster
type SubmasterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status corev1.PodPhase `json:"status,omitempty"`
	IP     string          `json:"ip,omitempty"`
	Containerized string   `json:"existing,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.status",name="STATUS",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.ip",name="IP",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.nodes",name="Nodes",type="integer"

// Submaster is the Schema for the submasters API
type Submaster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubmasterSpec   `json:"spec,omitempty"`
	Status SubmasterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubmasterList contains a list of Submaster
type SubmasterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Submaster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Submaster{}, &SubmasterList{})
}
