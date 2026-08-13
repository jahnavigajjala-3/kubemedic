import type { Severity } from '../types/healthreport';

const severityClass: Record<string, string> = {
  Critical: 'severity-critical',
  High: 'severity-high',
  Warning: 'severity-warning',
  Healthy: 'severity-healthy',
};

export function HealthBadge({ severity }: { severity: Severity }) {
  const cls = severityClass[severity] ?? 'severity-unknown';
  return <span className={`badge ${cls}`}>{severity}</span>;
}
