package controller

import (
	corev1 "k8s.io/api/core/v1"
)

type Diagnosis struct {
	Reason         string
	Message        string
	Recommendation string
	Severity       string
}

func diagnoseContainer(container corev1.ContainerStatus) *Diagnosis {
	if container.State.Waiting != nil {
		waiting := container.State.Waiting
		switch waiting.Reason {
		case "CrashLoopBackOff":
			return &Diagnosis{
				Reason:         waiting.Reason,
				Message:        messageOrDefault(waiting.Message, "Container is restarting repeatedly and has not become healthy."),
				Recommendation: "Inspect the application logs and fix the root cause of the crash before the container can recover.",
				Severity:       "Warning",
			}
		case "ImagePullBackOff":
			return &Diagnosis{
				Reason:         waiting.Reason,
				Message:        messageOrDefault(waiting.Message, "The container image cannot be pulled and the kubelet is backing off retries."),
				Recommendation: "Verify the image name, tag, registry credentials, and network access for the container image.",
				Severity:       "Warning",
			}
		case "ErrImagePull":
			return &Diagnosis{
				Reason:         waiting.Reason,
				Message:        messageOrDefault(waiting.Message, "The container image pull failed."),
				Recommendation: "Check the image reference, registry permissions, and the pull error details in the pod events.",
				Severity:       "Warning",
			}
		case "CreateContainerConfigError":
			return &Diagnosis{
				Reason:         waiting.Reason,
				Message:        messageOrDefault(waiting.Message, "The kubelet could not create the container using the provided configuration."),
				Recommendation: "Review the pod configuration, environment variables, and mounted secrets for invalid values.",
				Severity:       "Warning",
			}
		case "CreateContainerError":
			return &Diagnosis{
				Reason:         waiting.Reason,
				Message:        messageOrDefault(waiting.Message, "The container could not be created."),
				Recommendation: "Check the pod events and validate the container configuration, image, and runtime constraints.",
				Severity:       "Warning",
			}
		case "InvalidImageName":
			return &Diagnosis{
				Reason:         waiting.Reason,
				Message:        messageOrDefault(waiting.Message, "The container image name is invalid."),
				Recommendation: "Fix the image reference and ensure that it is valid for the configured registry and tag.",
				Severity:       "Warning",
			}
		default:
			if waiting.Reason != "PodInitializing" && waiting.Reason != "ContainerCreating" && waiting.Reason != "Running" && waiting.Reason != "Completed" {
				return &Diagnosis{
					Reason:         waiting.Reason,
					Message:        messageOrDefault(waiting.Message, "Container is waiting in a non-ready state."),
					Recommendation: "Inspect pod events and container logs to identify why the container is waiting.",
					Severity:       "Info",
				}
			}
		}
	}

	if container.State.Terminated != nil {
		terminated := container.State.Terminated
		switch terminated.Reason {
		case "Completed":
			return nil
		case "OOMKilled":
			return &Diagnosis{
				Reason:         terminated.Reason,
				Message:        messageOrDefault(terminated.Message, "The container was terminated because it exceeded its memory limit."),
				Recommendation: "Increase the container memory limit or investigate memory pressure and optimize memory usage.",
				Severity:       "Critical",
			}
		case "ContainerStatusUnknown":
			return &Diagnosis{
				Reason:         terminated.Reason,
				Message:        messageOrDefault(terminated.Message, "The container state is unknown and may be stuck or failed."),
				Recommendation: "Check the kubelet and container runtime logs, and verify the container is still functioning as expected.",
				Severity:       "Warning",
			}
		default:
			if terminated.ExitCode != 0 {
				return &Diagnosis{
					Reason:         terminated.Reason,
					Message:        messageOrDefault(terminated.Message, "The container exited with a non-zero exit code."),
					Recommendation: "Review the application logs and fix the cause of the unexpected exit code.",
					Severity:       "Warning",
				}
			}
		}
	}

	if container.State.Running != nil {
		return nil
	}

	return nil
}

func messageOrDefault(message, defaultMessage string) string {
	if message != "" {
		return message
	}
	return defaultMessage
}
