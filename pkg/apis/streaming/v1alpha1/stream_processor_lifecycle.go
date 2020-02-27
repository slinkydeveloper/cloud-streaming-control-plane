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

const (
	// StreamProcessorConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	StreamProcessorConditionReady = apis.ConditionReady

	StreamProcessorConditionStreamsReady apis.ConditionType = "StreamProcessorConditionStreamsReady"
)

var streamProcessorCondSet = apis.NewLivingConditionSet(
	StreamConditionConfigReady,
	StreamProcessorConditionStreamsReady,
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (as *StreamProcessor) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("StreamProcessor")
}

func (ass *StreamProcessorStatus) InitializeConditions() {
	streamProcessorCondSet.Manage(ass).InitializeConditions()
}

func (ass *StreamProcessorStatus) MarkStreamNotFound(name string) {
	streamProcessorCondSet.Manage(ass).MarkFalse(
		StreamProcessorConditionStreamsReady,
		"StreamNotFound",
		"Stream %s not found: please create it and then reapply this resource", name)
}

func (ass *StreamProcessorStatus) MarkStreamsReady() {
	streamProcessorCondSet.Manage(ass).MarkTrue(StreamProcessorConditionStreamsReady)
}

func (ass *StreamProcessorStatus) MarkDeploymentFailed(name string, err error) {
	inboundCondSet.Manage(ass).MarkFalse(
		InboundConditionReady,
		"DeploymentFailed",
		"Deployment %q wasn't found: %v", name, err)
}

func (ass *StreamProcessorStatus) MarkReady() {
	streamProcessorCondSet.Manage(ass).MarkTrue(StreamProcessorConditionReady)
}
