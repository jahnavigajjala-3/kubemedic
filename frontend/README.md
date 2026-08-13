# KubeMedic Dashboard

React frontend for the KubeMedic Kubernetes Pod Health Monitor.

Displays HealthReport data from the KubeMedic operator as a real-time observability dashboard with summary cards, filterable pod table, and detailed pod health views.

## Prerequisites

- Node.js 18+
- npm

## Setup

```bash
cd frontend
npm install
npm run dev
```

The dev server starts at `http://localhost:5173`.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VITE_USE_MOCK_API` | `true` | Set to `false` to use the real REST API |
| `VITE_API_BASE_URL` | (empty) | Base URL for the KubeMedic API (e.g. `http://localhost:8080`) |

## Mock Mode

By default the dashboard runs with mock data so it can be developed and demonstrated without the REST API. Set `VITE_USE_MOCK_API=false` in `.env` to switch to the real API.

## Expected REST API Endpoints

The frontend expects these endpoints (not yet implemented in the backend):

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/healthz` | Health check |
| `GET` | `/api/healthreports` | List all HealthReports (returns `{ items: HealthReport[] }`) |
| `GET` | `/api/healthreports/{namespace}/{podName}` | Get a single HealthReport |

The response shape should match the Kubernetes HealthReport CRD JSON serialization.

## Switching to the Real API

1. Set `VITE_USE_MOCK_API=false` in `.env`
2. Set `VITE_API_BASE_URL` to the API server address
3. Restart the dev server

## Build

```bash
npm run build
```

Output is in `dist/`.
