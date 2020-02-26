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

// Code generated by injection-gen. DO NOT EDIT.

package inbound

import (
	context "context"

	corev1 "k8s.io/api/core/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	scheme "k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	record "k8s.io/client-go/tools/record"
	client "knative.dev/pkg/client/injection/kube/client"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
	versionedscheme "knative.dev/streaming/pkg/client/clientset/versioned/scheme"
	injectionclient "knative.dev/streaming/pkg/client/injection/client"
	inbound "knative.dev/streaming/pkg/client/injection/informers/streaming/v1alpha1/inbound"
)

const (
	defaultControllerAgentName = "inbound-controller"
	defaultFinalizerName       = "inbounds.streaming.knative.dev"
	defaultQueueName           = "inbounds"
)

// NewImpl returns a controller.Impl that handles queuing and feeding work from
// the queue through an implementation of controller.Reconciler, delegating to
// the provided Interface and optional Finalizer methods. OptionsFn is used to return
// controller.Options to be used but the internal reconciler.
func NewImpl(ctx context.Context, r Interface, optionsFns ...controller.OptionsFn) *controller.Impl {
	logger := logging.FromContext(ctx)

	// Check the options function input. It should be 0 or 1.
	if len(optionsFns) > 1 {
		logger.Fatalf("up to one options function is supported, found %d", len(optionsFns))
	}

	inboundInformer := inbound.Get(ctx)

	recorder := controller.GetEventRecorder(ctx)
	if recorder == nil {
		// Create event broadcaster
		logger.Debug("Creating event broadcaster")
		eventBroadcaster := record.NewBroadcaster()
		watches := []watch.Interface{
			eventBroadcaster.StartLogging(logger.Named("event-broadcaster").Infof),
			eventBroadcaster.StartRecordingToSink(
				&v1.EventSinkImpl{Interface: client.Get(ctx).CoreV1().Events("")}),
		}
		recorder = eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: defaultControllerAgentName})
		go func() {
			<-ctx.Done()
			for _, w := range watches {
				w.Stop()
			}
		}()
	}

	rec := &reconcilerImpl{
		Client:     injectionclient.Get(ctx),
		Lister:     inboundInformer.Lister(),
		Recorder:   recorder,
		reconciler: r,
	}
	impl := controller.NewImpl(rec, logger, defaultQueueName)

	// Pass impl to the options. Save any optional results.
	for _, fn := range optionsFns {
		opts := fn(impl)
		if opts.ConfigStore != nil {
			rec.configStore = opts.ConfigStore
		}
	}

	return impl
}

func init() {
	versionedscheme.AddToScheme(scheme.Scheme)
}
