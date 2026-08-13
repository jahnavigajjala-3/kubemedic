import type { HealthReport } from '../types/healthreport';
import { HealthBadge } from './HealthBadge';

interface HealthReportTableProps {
  reports: HealthReport[];
  onSelect: (report: HealthReport) => void;
}

function formatTime(iso: string): string {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleString(undefined, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function HealthReportTable({ reports, onSelect }: HealthReportTableProps) {
  return (
    <div className="table-wrapper">
      <table className="report-table">
        <thead>
          <tr>
            <th>Namespace</th>
            <th>Pod</th>
            <th>Phase</th>
            <th>Diagnosis</th>
            <th>Severity</th>
            <th>Restarts</th>
            <th>Last Updated</th>
          </tr>
        </thead>
        <tbody>
          {reports.map((r) => (
            <tr
              key={`${r.status.namespace}/${r.status.podName}`}
              className="table-row"
              onClick={() => onSelect(r)}
            >
              <td>{r.status.namespace}</td>
              <td className="pod-name">{r.status.podName}</td>
              <td>{r.status.phase}</td>
              <td>{r.status.diagnosis}</td>
              <td>
                <HealthBadge severity={r.status.severity} />
              </td>
              <td>{r.status.restartCount}</td>
              <td>{formatTime(r.status.lastUpdated)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
