/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Stream struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the Stream (from the client).
	// +optional
	Spec StreamSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the Stream (from the controller).
	// +optional
	Status StreamStatus `json:"status,omitempty"`
}

// Check that Stream can be validated and defaulted.
var _ apis.Validatable = (*Stream)(nil)
var _ apis.Defaultable = (*Stream)(nil)
var _ kmeta.OwnerRefable = (*Stream)(nil)

// StreamBindingSpec holds the desired state of the Stream (from the client).
type StreamSpec struct{}

// StreamStatus communicates the observed state of the Stream (from the controller).
type StreamStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StreamList is a list of Stream resources
type StreamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Stream `json:"items"`
}
