import type { HealthReport, HealthReportListResponse } from '../types/healthreport';
import { mockHealthReports } from '../mock/data';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';
const USE_MOCK = import.meta.env.VITE_USE_MOCK_API !== 'false';

async function fetchJson<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json() as Promise<T>;
}

function normalizeReport(item: any, defaultNamespace?: string, defaultPodName?: string): HealthReport {
  if (item && item.status) {
    return item as HealthReport;
  }
  const ns = item?.namespace || defaultNamespace || '';
  const pod = item?.podName || defaultPodName || '';
  return {
    metadata: {
      name: item?.podName ? `${item.podName}-health` : pod,
      namespace: ns,
    },
    spec: { podName: pod },
    status: {
      namespace: ns,
      podName: pod,
      phase: item?.phase || 'Unknown',
      diagnosis: item?.diagnosis || 'Unknown',
      message: item?.message || '',
      restartCount: item?.restartCount ?? 0,
      lastUpdated: item?.lastUpdated || new Date().toISOString(),
      recommendation: item?.recommendation || '',
      severity: item?.severity || 'Info',
      conditions: item?.conditions || [],
    },
  };
}

export async function getHealthReports(): Promise<HealthReport[]> {
  if (USE_MOCK) {
    await new Promise((r) => setTimeout(r, 400));
    return mockHealthReports;
  }
  const data = await fetchJson<HealthReportListResponse | { items: any[] }>('/api/healthreports');
  return (data.items || []).map((item: any) => normalizeReport(item));
}

export async function getHealthReport(
  namespace: string,
  podName: string,
): Promise<HealthReport | undefined> {
  if (USE_MOCK) {
    await new Promise((r) => setTimeout(r, 200));
    return mockHealthReports.find(
      (r) => r.status.namespace === namespace && r.status.podName === podName,
    );
  }
  const data = await fetchJson<any>(`/api/healthreports/${namespace}/${podName}`);
  return normalizeReport(data, namespace, podName);
}

/**
 * Check if the API is reachable.
 * In mock mode this always returns true.
 */
export async function checkApiHealth(): Promise<boolean> {
  if (USE_MOCK) return true;
  try {
    const res = await fetch(`${API_BASE_URL}/api/healthz`);
    if (res.ok) return true;
    const fallbackRes = await fetch(`${API_BASE_URL}/api/healthreports`);
    return fallbackRes.ok;
  } catch {
    return false;
  }
}

export { USE_MOCK };
