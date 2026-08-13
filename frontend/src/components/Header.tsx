interface HeaderProps {
  apiConnected: boolean;
  onRefresh: () => void;
  useMock: boolean;
}

export function Header({ apiConnected, onRefresh, useMock }: HeaderProps) {
  return (
    <header className="header">
      <div className="header-left">
        <h1 className="header-title">KubeMedic</h1>
        <span className="header-subtitle">Kubernetes Pod Health Monitor</span>
      </div>
      <div className="header-right">
        {useMock && <span className="badge badge-mock">Mock Data</span>}
        <span className={`status-dot ${apiConnected ? 'connected' : 'disconnected'}`} />
        <span className="status-text">{apiConnected ? 'Connected' : 'API unavailable'}</span>
        <button className="btn-refresh" onClick={onRefresh} title="Refresh">
          ↻ Refresh
        </button>
      </div>
    </header>
  );
}
