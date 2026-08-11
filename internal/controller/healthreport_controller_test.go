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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	observabilityv1alpha1 "github.com/jahnavigajjala-3/kubemedic/api/v1alpha1"
)

var _ = Describe("HealthReport Controller", func() {
	const namespace = "default"
	ctx := context.Background()

	createPod := func(name string, statuses ...corev1.ContainerStatus) *corev1.Pod {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       types.UID(name + "-uid"),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main", Image: "busybox:latest"}},
			},
			Status: corev1.PodStatus{
				Phase:             corev1.PodRunning,
				ContainerStatuses: statuses,
			},
		}
		Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		pod.Status = corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: statuses,
		}
		Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
		return pod
	}

	cleanupPodAndReport := func(name string) {
		report := &observabilityv1alpha1.HealthReport{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, report); err == nil {
			Expect(k8sClient.Delete(ctx, report)).To(Succeed())
		}

		pod := &corev1.Pod{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, pod); err == nil {
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
		}
	}

	It("creates a HealthReport for an unhealthy Pod", func() {
		podName := "oom-pod-1"
		pod := createPod(podName, corev1.ContainerStatus{
			Name:  "main",
			Ready: false,
			State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", Message: "Container ran out of memory", ExitCode: 137}},
		})
		defer cleanupPodAndReport(podName)

		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		report := &observabilityv1alpha1.HealthReport{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, report)).To(Succeed())
		Expect(report.Spec.PodName).To(Equal(podName))
		Expect(report.Status.Diagnosis).To(Equal("OOMKilled"))
		Expect(report.Status.Severity).To(Equal("Critical"))
		Expect(report.Status.Recommendation).To(ContainSubstring("memory"))
		Expect(report.Status.Conditions).To(HaveLen(1))
		Expect(report.Status.Conditions[0].Type).To(Equal("Ready"))
		Expect(report.Status.Conditions[0].Status).To(Equal(metav1.ConditionFalse))
		Expect(report.OwnerReferences).To(HaveLen(1))
		Expect(report.OwnerReferences[0].Kind).To(Equal("Pod"))
		Expect(report.OwnerReferences[0].Name).To(Equal(podName))
		_ = pod
	})

	It("creates a HealthReport for a healthy Pod", func() {
		podName := "healthy-pod-1"
		createPod(podName, corev1.ContainerStatus{
			Name:  "main",
			Ready: true,
			State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
		})
		defer cleanupPodAndReport(podName)

		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		report := &observabilityv1alpha1.HealthReport{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, report)).To(Succeed())
		Expect(report.Status.Diagnosis).To(Equal("Healthy"))
		Expect(report.Status.Severity).To(Equal("Info"))
		Expect(report.Status.Conditions).To(HaveLen(1))
		Expect(report.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
	})

	It("updates an existing HealthReport when Pod health changes", func() {
		podName := "update-pod-1"
		pod := createPod(podName, corev1.ContainerStatus{
			Name:  "main",
			Ready: true,
			State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
		})
		defer cleanupPodAndReport(podName)

		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			current := &corev1.Pod{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, current); err != nil {
				return err
			}
			current.Status = corev1.PodStatus{
				Phase: corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{{
					Name:  "main",
					Ready: false,
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off restarting failed container"}},
				}},
			}
			return k8sClient.Status().Update(ctx, current)
		}).Should(Succeed())

		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		report := &observabilityv1alpha1.HealthReport{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, report)).To(Succeed())
		Expect(report.Status.Diagnosis).To(Equal("CrashLoopBackOff"))
		Expect(report.Status.Severity).To(Equal("Warning"))
		Expect(report.Status.Recommendation).To(ContainSubstring("logs"))
		_ = pod
	})

	It("does not create duplicate HealthReports for repeated reconciliation", func() {
		podName := "duplicate-pod-1"
		createPod(podName, corev1.ContainerStatus{
			Name:  "main",
			Ready: false,
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "Back-off pulling image"}},
		})
		defer cleanupPodAndReport(podName)

		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		list := &observabilityv1alpha1.HealthReportList{}
		Expect(k8sClient.List(ctx, list, ctrlclient.InNamespace(namespace))).To(Succeed())
		items := 0
		for _, item := range list.Items {
			if item.Name == podName {
				items++
			}
		}
		Expect(items).To(Equal(1))
	})

	It("handles a missing Pod gracefully", func() {
		podName := "missing-pod-1"
		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		report := &observabilityv1alpha1.HealthReport{}
		err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, report)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("sets a deterministic report name and sane pod identity", func() {
		podName := "identity-pod-1"
		createPod(podName, corev1.ContainerStatus{
			Name:  "main",
			Ready: false,
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ErrImagePull", Message: "Failed to pull image"}},
		})
		defer cleanupPodAndReport(podName)

		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		report := &observabilityv1alpha1.HealthReport{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, report)).To(Succeed())
		Expect(report.Name).To(Equal(podName))
		Expect(report.Spec.PodName).To(Equal(podName))
		Expect(report.Status.PodName).To(Equal(podName))
		Expect(report.Status.Namespace).To(Equal(namespace))
	})

	It("creates a HealthReport for a pod with multiple containers", func() {
		podName := fmt.Sprintf("multi-pod-%d", GinkgoRandomSeed())
		createPod(podName,
			corev1.ContainerStatus{Name: "healthy", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
			corev1.ContainerStatus{Name: "bad", Ready: false, State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off restarting failed container"}}},
		)
		defer cleanupPodAndReport(podName)

		reconciler := &HealthReportReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: podName}})
		Expect(err).NotTo(HaveOccurred())

		report := &observabilityv1alpha1.HealthReport{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, report)).To(Succeed())
		Expect(report.Status.Diagnosis).To(Equal("CrashLoopBackOff"))
		Expect(report.Status.Severity).To(Equal("Warning"))
		Expect(report.Status.Message).To(ContainSubstring("restarting"))
	})
})
