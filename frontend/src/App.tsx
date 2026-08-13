import { useMemo, useState } from 'react';
import { useHealthReports } from './hooks/useHealthReports';
import { USE_MOCK } from './services/api';
import { Header } from './components/Header';
import { SummaryCards } from './components/SummaryCards';
import { Filters } from './components/Filters';
import { HealthReportTable } from './components/HealthReportTable';
import { PodDetails } from './components/PodDetails';
import { LoadingState } from './components/LoadingState';
import { ErrorState } from './components/ErrorState';
import { EmptyState } from './components/EmptyState';
import type { HealthReport } from './types/healthreport';

export default function App() {
  const { reports, loading, error, apiConnected, refresh } = useHealthReports();
  const [selected, setSelected] = useState<HealthReport | null>(null);
  const [search, setSearch] = useState('');
  const [nsFilter, setNsFilter] = useState('');
  const [sevFilter, setSevFilter] = useState('');
  const [diagFilter, setDiagFilter] = useState('');

  const namespaces = useMemo(
    () => [...new Set(reports.map((r) => r.status.namespace))].sort(),
    [reports],
  );
  const severities = useMemo(
    () => [...new Set(reports.map((r) => r.status.severity))].sort(),
    [reports],
  );
  const diagnoses = useMemo(
    () => [...new Set(reports.map((r) => r.status.diagnosis))].sort(),
    [reports],
  );

  const filtered = useMemo(() => {
    return reports.filter((r) => {
      if (search && !r.status.podName.toLowerCase().includes(search.toLowerCase())) return false;
      if (nsFilter && r.status.namespace !== nsFilter) return false;
      if (sevFilter && r.status.severity !== sevFilter) return false;
      if (diagFilter && r.status.diagnosis !== diagFilter) return false;
      return true;
    });
  }, [reports, search, nsFilter, sevFilter, diagFilter]);

  return (
    <div className="app">
      <Header apiConnected={apiConnected} onRefresh={refresh} useMock={USE_MOCK} />

      {loading ? (
        <LoadingState />
      ) : error && !USE_MOCK ? (
        <ErrorState message={error} onRetry={refresh} />
      ) : reports.length === 0 ? (
        <EmptyState />
      ) : (
        <main className="main">
          <SummaryCards reports={reports} />
          <Filters
            search={search}
            onSearchChange={setSearch}
            namespace={nsFilter}
            onNamespaceChange={setNsFilter}
            severity={sevFilter}
            onSeverityChange={setSevFilter}
            diagnosis={diagFilter}
            onDiagnosisChange={setDiagFilter}
            namespaces={namespaces}
            severities={severities}
            diagnoses={diagnoses}
          />
          {filtered.length === 0 ? (
            <div className="state-container">
              <p>No pods match the current filters.</p>
            </div>
          ) : (
            <HealthReportTable reports={filtered} onSelect={setSelected} />
          )}
        </main>
      )}

      {selected && <PodDetails report={selected} onClose={() => setSelected(null)} />}
    </div>
  );
}
