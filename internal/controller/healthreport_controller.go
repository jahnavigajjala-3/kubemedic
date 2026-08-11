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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	observabilityv1alpha1 "github.com/jahnavigajjala-3/kubemedic/api/v1alpha1"
)

// HealthReportReconciler reconciles a HealthReport object.
type HealthReportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=observability.kubemedic.io,resources=healthreports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.kubemedic.io,resources=healthreports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.kubemedic.io,resources=healthreports/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile reconciles a Pod into a HealthReport.
func (r *HealthReportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("Pod not found; skipping reconciliation", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	diagnosis := diagnosePod(&pod)
	desiredStatus := buildHealthReportStatus(&pod, diagnosis)
	reportName := pod.Name

	reportKey := types.NamespacedName{Namespace: pod.Namespace, Name: reportName}
	report := &observabilityv1alpha1.HealthReport{}
	if err := r.Get(ctx, reportKey, report); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		report = &observabilityv1alpha1.HealthReport{
			ObjectMeta: metav1.ObjectMeta{
				Name:      reportName,
				Namespace: pod.Namespace,
			},
			Spec: observabilityv1alpha1.HealthReportSpec{PodName: pod.Name},
		}
		if err := controllerutil.SetControllerReference(&pod, report, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, report); err != nil {
			return ctrl.Result{}, err
		}
		report.Status = desiredStatus
		if err := r.Status().Update(ctx, report); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Created HealthReport for pod", "pod", pod.Name, "namespace", pod.Namespace, "diagnosis", diagnosis.Reason)
		return ctrl.Result{}, nil
	}

	if report.Spec.PodName != pod.Name {
		report.Spec.PodName = pod.Name
		if err := r.Update(ctx, report); err != nil {
			return ctrl.Result{}, err
		}
	}
	if !hasOwnerReference(report.OwnerReferences, &pod) {
		if err := controllerutil.SetControllerReference(&pod, report, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, report); err != nil {
			return ctrl.Result{}, err
		}
	}

	if !healthReportStatusEquivalent(report.Status, desiredStatus) {
		statusToWrite := desiredStatus
		statusToWrite.LastUpdated = metav1.Now()
		if len(statusToWrite.Conditions) > 0 {
			if len(report.Status.Conditions) > 0 && report.Status.Conditions[0].Type == statusToWrite.Conditions[0].Type && report.Status.Conditions[0].Status == statusToWrite.Conditions[0].Status && report.Status.Conditions[0].Reason == statusToWrite.Conditions[0].Reason {
				statusToWrite.Conditions[0].LastTransitionTime = report.Status.Conditions[0].LastTransitionTime
			} else {
				statusToWrite.Conditions[0].LastTransitionTime = statusToWrite.LastUpdated
			}
		}
		report.Status = statusToWrite
		if err := r.Status().Update(ctx, report); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Updated HealthReport status for pod", "pod", pod.Name, "namespace", pod.Namespace, "diagnosis", diagnosis.Reason)
	}

	return ctrl.Result{}, nil
}

func diagnosePod(pod *corev1.Pod) *Diagnosis {
	var best *Diagnosis
	for _, container := range pod.Status.InitContainerStatuses {
		if diagnosis := diagnoseContainer(container); diagnosis != nil && isMoreSevere(diagnosis, best) {
			best = diagnosis
		}
	}
	for _, container := range pod.Status.ContainerStatuses {
		if diagnosis := diagnoseContainer(container); diagnosis != nil && isMoreSevere(diagnosis, best) {
			best = diagnosis
		}
	}
	if best != nil {
		return best
	}

	switch pod.Status.Phase {
	case corev1.PodFailed:
		return &Diagnosis{
			Reason:         "Failed",
			Message:        "The pod has failed and one or more containers terminated unexpectedly.",
			Recommendation: "Review the pod events and container logs to determine why the workload failed.",
			Severity:       "Warning",
		}
	case corev1.PodSucceeded:
		return &Diagnosis{
			Reason:         "Healthy",
			Message:        "The pod completed successfully.",
			Recommendation: "No action required.",
			Severity:       "Info",
		}
	default:
		return &Diagnosis{
			Reason:         "Healthy",
			Message:        "All containers are healthy and the pod is running normally.",
			Recommendation: "No action required.",
			Severity:       "Info",
		}
	}
}

func buildHealthReportStatus(pod *corev1.Pod, diagnosis *Diagnosis) observabilityv1alpha1.HealthReportStatus {
	status := observabilityv1alpha1.HealthReportStatus{
		Namespace:      pod.Namespace,
		PodName:        pod.Name,
		Phase:          string(pod.Status.Phase),
		Diagnosis:      diagnosis.Reason,
		Message:        diagnosis.Message,
		Recommendation: diagnosis.Recommendation,
		Severity:       diagnosis.Severity,
	}

	restartCount := int32(0)
	for _, container := range pod.Status.InitContainerStatuses {
		if container.RestartCount > restartCount {
			restartCount = container.RestartCount
		}
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.RestartCount > restartCount {
			restartCount = container.RestartCount
		}
	}
	status.RestartCount = restartCount
	status.LastUpdated = metav1.Now()

	readyStatus := metav1.ConditionTrue
	if diagnosis.Severity != "Info" {
		readyStatus = metav1.ConditionFalse
	}
	status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             readyStatus,
		ObservedGeneration: 0,
		Reason:             diagnosis.Reason,
		Message:            diagnosis.Message,
		LastTransitionTime: status.LastUpdated,
	}}
	return status
}

func isMoreSevere(candidate, current *Diagnosis) bool {
	if current == nil {
		return true
	}
	return severityRank(candidate.Severity) > severityRank(current.Severity)
}

func severityRank(level string) int {
	switch level {
	case "Critical":
		return 3
	case "Warning":
		return 2
	case "Info":
		return 1
	default:
		return 0
	}
}

func hasOwnerReference(owners []metav1.OwnerReference, pod *corev1.Pod) bool {
	for _, owner := range owners {
		if owner.APIVersion == "v1" && owner.Kind == "Pod" && owner.Name == pod.Name && owner.UID == pod.UID {
			return true
		}
	}
	return false
}

func healthReportStatusEquivalent(a, b observabilityv1alpha1.HealthReportStatus) bool {
	left := a
	right := b
	left.LastUpdated = metav1.Time{}
	right.LastUpdated = metav1.Time{}
	if len(left.Conditions) > 0 && len(right.Conditions) > 0 {
		left.Conditions[0].LastTransitionTime = metav1.Time{}
		right.Conditions[0].LastTransitionTime = metav1.Time{}
	}
	return reflect.DeepEqual(left, right)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthReportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Named("healthreport").
		Complete(r)
}
