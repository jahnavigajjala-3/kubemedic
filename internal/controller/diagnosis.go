package controller

import corev1 "k8s.io/api/core/v1"

type Diagnosis struct {
	Reason         string
	Message        string
	Recommendation string
}

func diagnoseContainer(container corev1.ContainerStatus) *Diagnosis {

	if container.State.Terminated != nil {
		terminated := container.State.Terminated

		if terminated.Reason == "OOMKilled" {
			return &Diagnosis{
				Reason:         "OOMKilled",
				Message:        terminated.Message,
				Recommendation: "Increase the container memory limit or investigate memory usage.",
			}
		}
	}

	return nil
}
