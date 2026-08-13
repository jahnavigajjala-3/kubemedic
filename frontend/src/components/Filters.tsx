interface FiltersProps {
  search: string;
  onSearchChange: (v: string) => void;
  namespace: string;
  onNamespaceChange: (v: string) => void;
  severity: string;
  onSeverityChange: (v: string) => void;
  diagnosis: string;
  onDiagnosisChange: (v: string) => void;
  namespaces: string[];
  severities: string[];
  diagnoses: string[];
}

export function Filters({
  search,
  onSearchChange,
  namespace,
  onNamespaceChange,
  severity,
  onSeverityChange,
  diagnosis,
  onDiagnosisChange,
  namespaces,
  severities,
  diagnoses,
}: FiltersProps) {
  return (
    <div className="filters">
      <input
        className="filter-input"
        type="text"
        placeholder="Search pods..."
        value={search}
        onChange={(e) => onSearchChange(e.target.value)}
      />
      <select
        className="filter-select"
        value={namespace}
        onChange={(e) => onNamespaceChange(e.target.value)}
      >
        <option value="">All Namespaces</option>
        {namespaces.map((ns) => (
          <option key={ns} value={ns}>
            {ns}
          </option>
        ))}
      </select>
      <select
        className="filter-select"
        value={severity}
        onChange={(e) => onSeverityChange(e.target.value)}
      >
        <option value="">All Severities</option>
        {severities.map((s) => (
          <option key={s} value={s}>
            {s}
          </option>
        ))}
      </select>
      <select
        className="filter-select"
        value={diagnosis}
        onChange={(e) => onDiagnosisChange(e.target.value)}
      >
        <option value="">All Diagnoses</option>
        {diagnoses.map((d) => (
          <option key={d} value={d}>
            {d}
          </option>
        ))}
      </select>
    </div>
  );
}
