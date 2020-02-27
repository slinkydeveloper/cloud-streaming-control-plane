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

var inboundCondSet = apis.NewLivingConditionSet(
	InboundConditionReady,
	InboundConditionStreamReady,
)

const (
	// InboundConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	InboundConditionReady = apis.ConditionReady

	InboundConditionStreamReady apis.ConditionType = "InboundConditionStreamReady"
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (is *Inbound) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Inbound")
}

func (ass *InboundStatus) InitializeConditions() {
	inboundCondSet.Manage(ass).InitializeConditions()
}

func (ass *InboundStatus) MarkStreamNotFound(name string) {
	inboundCondSet.Manage(ass).MarkFalse(
		InboundConditionStreamReady,
		"StreamNotFound",
		"Stream %s not found: please create it and then reapply this resource", name)
}

func (ass *InboundStatus) MarkStreamReady() {
	inboundCondSet.Manage(ass).MarkTrue(InboundConditionStreamReady)
}

func (ass *InboundStatus) MarkServiceFailed(name string, err error) {
	inboundCondSet.Manage(ass).MarkFalse(
		InboundConditionReady,
		"ServiceFailed",
		"Service %q failed: %v", name, err)
}

func (ass *InboundStatus) MarkDeploymentFailed(name string, err error) {
	inboundCondSet.Manage(ass).MarkFalse(
		InboundConditionReady,
		"DeploymentFailed",
		"Deployment %q wasn't found: %v", name, err)
}

func (ass *InboundStatus) MarkReady() {
	inboundCondSet.Manage(ass).MarkTrue(InboundConditionReady)
}
