import type { HealthReport } from '../types/healthreport';
import { HealthBadge } from './HealthBadge';

interface PodDetailsProps {
  report: HealthReport;
  onClose: () => void;
}

function formatTime(iso: string): string {
  if (!iso) return '—';
  return new Date(iso).toLocaleString(undefined, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function PodDetails({ report, onClose }: PodDetailsProps) {
  const s = report.status;

  return (
    <div className="detail-overlay" onClick={onClose}>
      <div className="detail-panel" onClick={(e) => e.stopPropagation()}>
        <div className="detail-header">
          <div>
            <h2 className="detail-title">{s.podName}</h2>
            <span className="detail-ns">{s.namespace}</span>
          </div>
          <button className="btn-close" onClick={onClose}>
            ✕
          </button>
        </div>

        <div className="detail-grid">
          <div className="detail-field">
            <label>Severity</label>
            <HealthBadge severity={s.severity} />
          </div>
          <div className="detail-field">
            <label>Phase</label>
            <span>{s.phase}</span>
          </div>
          <div className="detail-field">
            <label>Diagnosis</label>
            <span>{s.diagnosis}</span>
          </div>
          <div className="detail-field">
            <label>Restart Count</label>
            <span>{s.restartCount}</span>
          </div>
          <div className="detail-field">
            <label>Last Updated</label>
            <span>{formatTime(s.lastUpdated)}</span>
          </div>
        </div>

        {s.message && (
          <div className="detail-section">
            <label>Message</label>
            <p>{s.message}</p>
          </div>
        )}

        {s.recommendation && (
          <div className="detail-section">
            <label>Recommendation</label>
            <p>{s.recommendation}</p>
          </div>
        )}

        {s.conditions && s.conditions.length > 0 && (
          <div className="detail-section">
            <label>Conditions</label>
            <div className="conditions-list">
              {s.conditions.map((c, i) => (
                <div key={i} className="condition-card">
                  <div className="condition-header">
                    <span className="condition-type">{c.type}</span>
                    <span className={`status-dot ${c.status === 'True' ? 'connected' : 'disconnected'}`} />
                    <span>{c.status}</span>
                  </div>
                  {c.reason && (
                    <div className="condition-field">
                      <label>Reason</label>
                      <span>{c.reason}</span>
                    </div>
                  )}
                  {c.message && (
                    <div className="condition-field">
                      <label>Message</label>
                      <span>{c.message}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
