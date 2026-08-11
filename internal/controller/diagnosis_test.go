package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Diagnosis engine", func() {
	DescribeTable("diagnoseContainer recognizes unhealthy states",
		func(container corev1.ContainerStatus, expectedReason, expectedSeverity string) {
			diagnosis := diagnoseContainer(container)
			Expect(diagnosis).NotTo(BeNil())
			Expect(diagnosis.Reason).To(Equal(expectedReason))
			Expect(diagnosis.Severity).To(Equal(expectedSeverity))
		},
		Entry("OOMKilled terminated container",
			corev1.ContainerStatus{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", Message: "Container ran out of memory", ExitCode: 137}}},
			"OOMKilled",
			"Critical",
		),
		Entry("CrashLoopBackOff waiting container",
			corev1.ContainerStatus{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off restarting failed container"}}},
			"CrashLoopBackOff",
			"Warning",
		),
		Entry("ImagePullBackOff waiting container",
			corev1.ContainerStatus{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "Back-off pulling image"}}},
			"ImagePullBackOff",
			"Warning",
		),
		Entry("ErrImagePull waiting container",
			corev1.ContainerStatus{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ErrImagePull", Message: "Failed to pull image"}}},
			"ErrImagePull",
			"Warning",
		),
		Entry("terminated container with error exit code",
			corev1.ContainerStatus{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Error", Message: "Container exited unexpectedly", ExitCode: 1}}},
			"Error",
			"Warning",
		),
	)

	It("returns nil for a healthy running container", func() {
		container := corev1.ContainerStatus{State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}}
		Expect(diagnoseContainer(container)).To(BeNil())
	})

	It("uses the most severe diagnosis across multiple containers", func() {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "app", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
					{Name: "oom", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137, Message: "Container ran out of memory"}}},
					{Name: "pull", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "Back-off pulling image"}}},
				},
			},
		}

		diagnosis := diagnosePod(pod)
		Expect(diagnosis).NotTo(BeNil())
		Expect(diagnosis.Reason).To(Equal("OOMKilled"))
		Expect(diagnosis.Severity).To(Equal("Critical"))
	})
})
