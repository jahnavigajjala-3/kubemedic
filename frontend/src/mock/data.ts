import type { HealthReport } from '../types/healthreport';

/**
 * Mock HealthReport data for development without the REST API.
 * These match the actual KubeMedic HealthReport CRD structure.
 * Remove or disable once the real API is available.
 */
export const mockHealthReports: HealthReport[] = [
  {
    metadata: { name: 'api-7f8d9b-health', namespace: 'default' },
    spec: { podName: 'api-7f8d9b' },
    status: {
      namespace: 'default',
      podName: 'api-7f8d9b',
      phase: 'Running',
      diagnosis: 'Healthy',
      message: 'Pod is running normally.',
      restartCount: 0,
      lastUpdated: new Date().toISOString(),
      recommendation: '',
      severity: 'Healthy',
      conditions: [
        {
          type: 'Ready',
          status: 'True',
          lastTransitionTime: new Date().toISOString(),
          reason: 'PodReady',
          message: 'All containers are ready.',
        },
      ],
    },
  },
  {
    metadata: { name: 'worker-5d2f1a-health', namespace: 'default' },
    spec: { podName: 'worker-5d2f1a' },
    status: {
      namespace: 'default',
      podName: 'worker-5d2f1a',
      phase: 'Running',
      diagnosis: 'Healthy',
      message: 'Pod is running normally.',
      restartCount: 0,
      lastUpdated: new Date().toISOString(),
      recommendation: '',
      severity: 'Healthy',
      conditions: [
        {
          type: 'Ready',
          status: 'True',
          lastTransitionTime: new Date().toISOString(),
          reason: 'PodReady',
          message: 'All containers are ready.',
        },
      ],
    },
  },
  {
    metadata: { name: 'test-oom-health', namespace: 'default' },
    spec: { podName: 'test-oom' },
    status: {
      namespace: 'default',
      podName: 'test-oom',
      phase: 'Running',
      diagnosis: 'OOMKilled',
      message: 'The container was terminated because it exceeded its memory limit.',
      restartCount: 3,
      lastUpdated: new Date(Date.now() - 300_000).toISOString(),
      recommendation:
        'Increase the container memory limit or investigate memory pressure and optimize memory usage.',
      severity: 'Critical',
      conditions: [
        {
          type: 'Ready',
          status: 'False',
          lastTransitionTime: new Date(Date.now() - 300_000).toISOString(),
          reason: 'OOMKilled',
          message: 'The container was terminated because it exceeded its memory limit.',
        },
      ],
    },
  },
  {
    metadata: { name: 'redis-0-health', namespace: 'default' },
    spec: { podName: 'redis-0' },
    status: {
      namespace: 'default',
      podName: 'redis-0',
      phase: 'Running',
      diagnosis: 'CrashLoopBackOff',
      message: 'Container is in CrashLoopBackOff state.',
      restartCount: 12,
      lastUpdated: new Date(Date.now() - 60_000).toISOString(),
      recommendation:
        'Check container logs for crash reason. Verify configuration and resource limits.',
      severity: 'Critical',
      conditions: [
        {
          type: 'Ready',
          status: 'False',
          lastTransitionTime: new Date(Date.now() - 60_000).toISOString(),
          reason: 'CrashLoopBackOff',
          message: 'Container is restarting repeatedly.',
        },
      ],
    },
  },
  {
    metadata: { name: 'nginx-deploy-health', namespace: 'kube-system' },
    spec: { podName: 'nginx-deploy-8c4b2' },
    status: {
      namespace: 'kube-system',
      podName: 'nginx-deploy-8c4b2',
      phase: 'Pending',
      diagnosis: 'ImagePullBackOff',
      message: 'Failed to pull image "nginx:invalid-tag".',
      restartCount: 0,
      lastUpdated: new Date(Date.now() - 120_000).toISOString(),
      recommendation:
        'Verify the image name and tag. Ensure the image exists in the registry and credentials are configured.',
      severity: 'Warning',
      conditions: [
        {
          type: 'Ready',
          status: 'False',
          lastTransitionTime: new Date(Date.now() - 120_000).toISOString(),
          reason: 'ImagePullBackOff',
          message: 'Back-off pulling image "nginx:invalid-tag".',
        },
        {
          type: 'ContainersReady',
          status: 'False',
          lastTransitionTime: new Date(Date.now() - 120_000).toISOString(),
          reason: 'ContainersNotReady',
          message: 'containers with unready status: [nginx]',
        },
      ],
    },
  },
  {
    metadata: { name: 'scheduler-health', namespace: 'kube-system' },
    spec: { podName: 'scheduler-abc12' },
    status: {
      namespace: 'kube-system',
      podName: 'scheduler-abc12',
      phase: 'Running',
      diagnosis: 'Healthy',
      message: 'Pod is running normally.',
      restartCount: 1,
      lastUpdated: new Date(Date.now() - 600_000).toISOString(),
      recommendation: '',
      severity: 'Healthy',
      conditions: [
        {
          type: 'Ready',
          status: 'True',
          lastTransitionTime: new Date(Date.now() - 600_000).toISOString(),
          reason: 'PodReady',
          message: 'All containers are ready.',
        },
      ],
    },
  },
  {
    metadata: { name: 'etcd-health', namespace: 'monitoring' },
    spec: { podName: 'etcd-monitor-0' },
    status: {
      namespace: 'monitoring',
      podName: 'etcd-monitor-0',
      phase: 'Running',
      diagnosis: 'HighRestartCount',
      message: 'Container has restarted 5 times.',
      restartCount: 5,
      lastUpdated: new Date(Date.now() - 900_000).toISOString(),
      recommendation: 'Investigate why the container keeps restarting. Check resource limits and liveness probes.',
      severity: 'High',
      conditions: [
        {
          type: 'Ready',
          status: 'True',
          lastTransitionTime: new Date(Date.now() - 900_000).toISOString(),
          reason: 'PodReady',
          message: 'Pod is ready but has high restart count.',
        },
      ],
    },
  },
];
