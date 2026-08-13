interface ErrorStateProps {
  message: string;
  onRetry: () => void;
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="state-container error-state">
      <p className="error-title">Unable to connect to KubeMedic API</p>
      <p className="error-detail">{message}</p>
      <p className="error-hint">Check that the API server is running.</p>
      <button className="btn-refresh" onClick={onRetry}>
        Retry
      </button>
    </div>
  );
}
