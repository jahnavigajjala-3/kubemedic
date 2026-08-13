import { useCallback, useEffect, useRef, useState } from 'react';
import type { HealthReport } from '../types/healthreport';
import { getHealthReports, checkApiHealth } from '../services/api';

const REFRESH_INTERVAL_MS = 10_000;

interface UseHealthReportsResult {
  reports: HealthReport[];
  loading: boolean;
  error: string | null;
  apiConnected: boolean;
  refresh: () => void;
}

export function useHealthReports(): UseHealthReportsResult {
  const [reports, setReports] = useState<HealthReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [apiConnected, setApiConnected] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const connected = await checkApiHealth();
      setApiConnected(connected);
      const data = await getHealthReports();
      setReports(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      setApiConnected(false);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    intervalRef.current = setInterval(fetchData, REFRESH_INTERVAL_MS);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [fetchData]);

  const refresh = useCallback(() => {
    setLoading(true);
    fetchData();
  }, [fetchData]);

  return { reports, loading, error, apiConnected, refresh };
}
