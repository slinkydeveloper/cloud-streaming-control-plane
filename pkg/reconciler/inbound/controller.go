/*
Copyright 2020 The Knative Authors

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

package inbound

import (
	context "context"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	configmap "knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"

	"knative.dev/streaming/pkg/client/injection/informers/streaming/v1alpha1/inbound"
	"knative.dev/streaming/pkg/client/injection/informers/streaming/v1alpha1/stream"
	v1alpha1inbound "knative.dev/streaming/pkg/client/injection/reconciler/streaming/v1alpha1/inbound"
)

// NewController creates a Reconciler for Inbound and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	inboundInformer := inbound.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	svcInformer := service.Get(ctx)
	streamsInformer := stream.Get(ctx)
	kubeClient := kubeclient.Get(ctx)

	r := &Reconciler{
		deploymentInformer: deploymentInformer,
		svcInformer:        svcInformer,
		streamsInformer:    streamsInformer,
		kubeClient:         kubeClient,
	}
	impl := v1alpha1inbound.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	inboundInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	svcInformer.Informer().AddEventHandler(controller.HandleAll(impl.EnqueueControllerOf))
	deploymentInformer.Informer().AddEventHandler(controller.HandleAll(impl.EnqueueControllerOf))

	return impl
}
