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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StreamProcessor struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the StreamProcessor (from the client).
	// +optional
	Spec StreamProcessorSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the StreamProcessor (from the controller).
	// +optional
	Status StreamProcessorStatus `json:"status,omitempty"`
}

// Check that StreamProcessor can be validated and defaulted.
var _ apis.Validatable = (*StreamProcessor)(nil)
var _ apis.Defaultable = (*StreamProcessor)(nil)
var _ kmeta.OwnerRefable = (*StreamProcessor)(nil)

// StreamProcessorSpec holds the desired state of the StreamProcessor (from the client).
type StreamProcessorSpec struct {
	Container corev1.Container `json:"container"`

	Input  []StreamBindingSpec `json:"input"`
	Output []StreamBindingSpec `json:"output"`
	// +optional
	State *StreamBindingSpec `json:"state,omitempty"`
}

type StreamBindingSpec struct {
	Name string `json:"name"`
	// +optional
	ParameterName string `json:"parameterName,omitempty"`
	// +optional
	Key string `json:"key,omitempty"`
}

// StreamProcessorStatus communicates the observed state of the StreamProcessor (from the controller).
type StreamProcessorStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StreamProcessorList is a list of StreamProcessor resources
type StreamProcessorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StreamProcessor `json:"items"`
}
