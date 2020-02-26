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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var streamCondSet = apis.NewLivingConditionSet(
	StreamConditionReady,
	StreamConditionConfigReady,
)

const (
	// StreamConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	StreamConditionReady = apis.ConditionReady

	StreamConditionConfigReady apis.ConditionType = "ConfigurationReady"
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (is *Stream) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Stream")
}

func (streamStatus *StreamStatus) InitializeConditions() {
	streamCondSet.Manage(streamStatus).InitializeConditions()
}

func (streamStatus *StreamStatus) MarkNotReady(reason, messageFormat string, messageA ...interface{}) {
	streamCondSet.Manage(streamStatus).MarkFalse(StreamConditionReady, reason, messageFormat, messageA...)
}

func (streamStatus *StreamStatus) MarkReady() {
	streamCondSet.Manage(streamStatus).MarkTrue(StreamConditionReady)
}

func (streamStatus *StreamStatus) MarkConfigTrue() {
	streamCondSet.Manage(streamStatus).MarkTrue(StreamConditionConfigReady)
}

func (streamStatus *StreamStatus) MarkConfigFailed(reason, messageFormat string, messageA ...interface{}) {
	streamCondSet.Manage(streamStatus).MarkFalse(StreamConditionConfigReady, reason, messageFormat, messageA...)
}
