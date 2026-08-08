/*
Copyright 2026.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// HealthReportReconciler reconciles a HealthReport object
type HealthReportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=observability.kubemedic.io,resources=healthreports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.kubemedic.io,resources=healthreports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.kubemedic.io,resources=healthreports/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main Kubernetes reconciliation loop.
func (r *HealthReportReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	var pod corev1.Pod

	// Get the Pod that triggered this reconciliation.
	err := r.Get(ctx, req.NamespacedName, &pod)

	if err != nil {
		// The Pod may have been deleted between the event and
		// our attempt to retrieve it. That is not an error.
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Basic information about the Pod.
	log := logf.FromContext(ctx)

	log.Info(
		"Pod received",
		"name", pod.Name,
		"namespace", pod.Namespace,
		"phase", pod.Status.Phase,
		"containers", len(pod.Status.ContainerStatuses),
	)

	// Inspect every container in the Pod.
	for _, container := range pod.Status.ContainerStatuses {
		log.Info(
			"Container status",
			"container", container.Name,
			"restartCount", container.RestartCount,
			"ready", container.Ready,
			"waiting", container.State.Waiting != nil,
			"running", container.State.Running != nil,
			"terminated", container.State.Terminated != nil,
		)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthReportReconciler) SetupWithManager(
	mgr ctrl.Manager,
) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Named("healthreport").
		Complete(r)
}
