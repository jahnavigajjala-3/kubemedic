import type { HealthReport } from '../types/healthreport';

interface SummaryCardsProps {
  reports: HealthReport[];
}

export function SummaryCards({ reports }: SummaryCardsProps) {
  const total = reports.length;
  const healthy = reports.filter((r) => r.status.severity === 'Healthy').length;
  const warning = reports.filter((r) => r.status.severity === 'Warning').length;
  const critical = reports.filter(
    (r) => r.status.severity === 'Critical' || r.status.severity === 'High',
  ).length;

  const cards = [
    { label: 'Total Pods', value: total, cls: 'card-total' },
    { label: 'Healthy', value: healthy, cls: 'card-healthy' },
    { label: 'Warning', value: warning, cls: 'card-warning' },
    { label: 'Critical', value: critical, cls: 'card-critical' },
  ];

  return (
    <div className="summary-cards">
      {cards.map((c) => (
        <div key={c.label} className={`summary-card ${c.cls}`}>
          <span className="card-label">{c.label}</span>
          <span className="card-value">{c.value}</span>
        </div>
      ))}
    </div>
  );
}
