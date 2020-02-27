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

package streamprocessor

import (
	context "context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/kmeta"
	reconciler "knative.dev/pkg/reconciler"

	v1alpha1 "knative.dev/streaming/pkg/apis/streaming/v1alpha1"
	streamprocessor "knative.dev/streaming/pkg/client/injection/reconciler/streaming/v1alpha1/streamprocessor"
	"knative.dev/streaming/pkg/kafka"
)

const (
	STREAM_PROCESSOR_IMAGE = "docker.io/slinkydeveloper/engine-kafka"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason StreamProcessorReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "StreamProcessorReconciled", "StreamProcessor reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements controller.Reconciler for StreamProcessor resources.
type Reconciler struct {
	svcInformer        corev1informers.ServiceInformer
	deploymentInformer appsv1informers.DeploymentInformer
	kubeClient         kubernetes.Interface
}

// Check that our Reconciler implements Interface
var _ streamprocessor.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, streamProcessor *v1alpha1.StreamProcessor) reconciler.Event {
	streamProcessor.Status.InitializeConditions()

	patchStreamProcessor(streamProcessor)

	_, err := r.reconcileDeployment(ctx, createDeployment(streamProcessor))
	if err != nil {
		return err
	}

	streamProcessor.Status.MarkReady()

	streamProcessor.Status.ObservedGeneration = streamProcessor.Generation
	return newReconciledNormal(streamProcessor.Namespace, streamProcessor.Name)
}

func (r *Reconciler) reconcileDeployment(ctx context.Context, expected *appsv1.Deployment) (*appsv1.Deployment, error) {
	current, err := r.deploymentInformer.Lister().Deployments(expected.Namespace).Get(expected.Name)
	if apierrs.IsNotFound(err) {
		current, err = r.kubeClient.AppsV1().Deployments(expected.Namespace).Create(expected)
		if err != nil {
			return nil, err
		}
		return current, nil
	} else if err != nil {
		return nil, err
	}

	if !equality.Semantic.DeepDerivative(expected.Spec, current.Spec) {
		// Don't modify the informers copy.
		desired := current.DeepCopy()
		desired.Spec = expected.Spec
		current, err = r.kubeClient.AppsV1().Deployments(current.Namespace).Update(desired)
		if err != nil {
			return nil, err
		}
	}
	return current, nil
}

func patchStreamProcessor(streamProcessor *v1alpha1.StreamProcessor) {
	container := &streamProcessor.Spec.Container

	found := false
	for _, env := range container.Env {
		if env.Name == "UNIX_DOMAIN_SOCKET" {
			env.Value = "/data/function"
			found = true
			break
		}
	}
	if found == false {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "UNIX_DOMAIN_SOCKET",
			Value: "/data/function",
		})
	}

	found = false
	for _, vm := range container.VolumeMounts {
		if vm.Name == "uds-data" {
			vm.MountPath = "/data"
			found = true
			break
		}
	}
	if found == false {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "uds-data",
			MountPath: "/data",
		})
	}
}

func deploymentName(streamProcessor *v1alpha1.StreamProcessor) string {
	return fmt.Sprintf("streaming-node-deployment-%s", streamProcessor.Name)
}

func appId(streamProcessor *v1alpha1.StreamProcessor) string {
	return fmt.Sprintf("stream-processor-%s", streamProcessor.Name)
}

func generateStreamBindingEnvValue(streamBindings []v1alpha1.StreamBindingSpec) string {
	res := []string{}

	for _, sb := range streamBindings {
		if sb.ParameterName != "" && sb.Key != "" {
			res = append(res, fmt.Sprintf("%s:%s:%s", sb.Name, sb.ParameterName, sb.Key))
		} else if sb.ParameterName != "" {
			res = append(res, fmt.Sprintf("%s:%s", sb.Name, sb.ParameterName))
		} else {
			res = append(res, sb.Name)
		}
	}

	return strings.Join(res, ",")
}

func generateEnvs(streamProcessor *v1alpha1.StreamProcessor) []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	envs = append(envs, corev1.EnvVar{
		Name:  "KAFKA_BROKERS",
		Value: kafka.HARDCODED_BOOTSTRAP_SERVER,
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "APP_ID",
		Value: appId(streamProcessor),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  "INPUT_STREAMS",
		Value: generateStreamBindingEnvValue(streamProcessor.Spec.Input),
	})

	if streamProcessor.Spec.Output != nil && len(streamProcessor.Spec.Output) > 0 {
		envs = append(envs, corev1.EnvVar{
			Name:  "OUTPUT_STREAMS",
			Value: generateStreamBindingEnvValue(streamProcessor.Spec.Output),
		})
	}

	if streamProcessor.Spec.State != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "STATE_STREAM",
			Value: generateStreamBindingEnvValue([]v1alpha1.StreamBindingSpec{*streamProcessor.Spec.State}),
		})
	}

	return envs
}

func createDeployment(streamProcessor *v1alpha1.StreamProcessor) *appsv1.Deployment {
	name := deploymentName(streamProcessor)
	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(streamProcessor),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "stream-processor",
							Image: STREAM_PROCESSOR_IMAGE,
							Env:   generateEnvs(streamProcessor),
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "uds-data",
								MountPath: "/data",
							}},
						},
						streamProcessor.Spec.Container,
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{{
						Name: "uds-data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
				},
			},
		},
	}
}
