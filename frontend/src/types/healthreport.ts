export interface Condition {
  type: string;
  status: string;
  lastTransitionTime: string;
  reason: string;
  message: string;
  observedGeneration?: number;
}

export interface HealthReportStatus {
  namespace: string;
  podName: string;
  phase: string;
  diagnosis: string;
  message: string;
  restartCount: number;
  lastUpdated: string;
  recommendation: string;
  severity: string;
  conditions: Condition[];
}

export interface HealthReportSpec {
  podName: string;
}

export interface HealthReport {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp?: string;
    uid?: string;
  };
  spec: HealthReportSpec;
  status: HealthReportStatus;
}

export interface HealthReportListResponse {
  items: HealthReport[];
}

export type Severity = 'Critical' | 'High' | 'Warning' | 'Healthy' | string;
