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
	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	reconciler "knative.dev/pkg/reconciler"

	v1alpha1 "knative.dev/streaming/pkg/apis/streaming/v1alpha1"
	inbound "knative.dev/streaming/pkg/client/injection/reconciler/streaming/v1alpha1/inbound"
	"knative.dev/streaming/pkg/kafka"
)

const (
	INBOUND_IMAGE = "docker.io/slinkydeveloper/inbound-kafka"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason InboundReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "InboundReconciled", "Inbound reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements controller.Reconciler for Inbound resources.
type Reconciler struct {
	svcInformer        corev1informers.ServiceInformer
	deploymentInformer appsv1informers.DeploymentInformer
	kubeClient         kubernetes.Interface
}

// Check that our Reconciler implements Interface
var _ inbound.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, inbound *v1alpha1.Inbound) reconciler.Event {
	logger := logging.FromContext(ctx)
	inbound.Status.InitializeConditions()

	deployment, err := r.reconcileDeployment(ctx, createDeployment(inbound))
	if err != nil {
		logger.Error("Error while reconciling the deployment", zap.Error(err))
		inbound.Status.MarkDeploymentFailed(deployment.Name, err)
		return err
	}

	svc, err := r.reconcileSvc(ctx, createService(inbound))
	if err != nil {
		logger.Error("Error while reconciling the service", zap.Error(err))
		inbound.Status.MarkServiceFailed(svc.Name, err)
		return err
	}

	inbound.Status.Address = &duckv1.Addressable{
		URL: &apis.URL{
			Scheme: "http",
			Host:   network.GetServiceHostname(svc.Name, svc.Namespace),
		},
	}
	inbound.Status.MarkReady()

	inbound.Status.ObservedGeneration = inbound.Generation
	return newReconciledNormal(inbound.Namespace, inbound.Name)
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

func (r *Reconciler) reconcileSvc(ctx context.Context, expected *corev1.Service) (*corev1.Service, error) {
	current, err := r.svcInformer.Lister().Services(expected.Namespace).Get(expected.Name)
	if apierrs.IsNotFound(err) {
		current, err = r.kubeClient.CoreV1().Services(expected.Namespace).Create(expected)
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
		current, err = r.kubeClient.CoreV1().Services(current.Namespace).Update(desired)
		if err != nil {
			return nil, err
		}
	}
	return current, nil
}

func deploymentName(inbound *v1alpha1.Inbound) string {
	return fmt.Sprintf("inbound-deployment-%s", inbound.Name)
}

func svcName(inbound *v1alpha1.Inbound) string {
	return fmt.Sprintf("inbound-svc-%s", inbound.Name)
}

func selectorName(inbound *v1alpha1.Inbound) map[string]string {
	return map[string]string{"streaming-inbound": inbound.Name}
}

func generateEnvs(inbound *v1alpha1.InboundSpec) []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	envs = append(envs, corev1.EnvVar{
		Name:  "NAME",
		Value: inbound.StreamName,
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "PORT",
		Value: "8080",
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "KAFKA_BROKERS",
		Value: kafka.HARDCODED_BOOTSTRAP_SERVER,
	})

	if inbound.Key != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "KEY_METADATA",
			Value: inbound.Key,
		})
	}

	return envs
}

func createDeployment(inbound *v1alpha1.Inbound) *appsv1.Deployment {
	name := deploymentName(inbound)
	labels := selectorName(inbound)
	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(inbound),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "inbound",
						Image: INBOUND_IMAGE,
						Env:   generateEnvs(&inbound.Spec),
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}

func createService(inbound *v1alpha1.Inbound) *corev1.Service {
	name := svcName(inbound)
	selector := selectorName(inbound)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: selector,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(inbound),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}
}
