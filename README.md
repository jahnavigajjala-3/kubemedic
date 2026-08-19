# KubeMedic 🏥⚡

**KubeMedic** is an automated Kubernetes operator and diagnostic tool designed to monitor pod health, detect failures in real time, generate actionable diagnostic reports, and visualize cluster status through an interactive Web Dashboard and REST API.

---

## 🌟 Key Features

* **Automated Pod Diagnosis**: Continuously watches all Pods across cluster namespaces and diagnoses failure states like `OOMKilled`, `CrashLoopBackOff`, `ImagePullBackOff`, `ErrImagePull`, `HighRestartCount`, `Pending`, `Failed`, and `Unhealthy`.
* **HealthReport Custom Resource (CRD)**: Automatically creates and reconciles `HealthReport` objects (`observability.kubemedic.io/v1alpha1`) linked via Kubernetes OwnerReferences for automatic garbage collection.
* **Actionable Recommendations**: Provides human-readable diagnosis messages and specific remediation recommendations for quick troubleshooting.
* **Built-in REST API**: High-performance HTTP API server running alongside the operator on port `:8080` (providing `/api/healthreports`, `/api/healthreports/{namespace}/{podName}`, and `/api/healthz`).
* **Modern Web Dashboard**: A fast React + TypeScript + Vite frontend featuring real-time status updates, search & namespace filtering, severity cards (`Critical`, `High`, `Warning`, `Healthy`), and an interactive Pod Details drawer.
* **Flexible Monitoring**: Supports local development, single-command bundle deployment, and remote cluster monitoring via kubeconfig.

---

## 🏗️ Architecture Overview

```
 ┌─────────────────────────────────────────────────────────┐
 │                   Kubernetes Cluster                    │
 │                                                         │
 │  ┌──────────┐      Watches       ┌───────────────────┐  │
 │  │   Pod    │ <───────────────── │ KubeMedic         │  │
 │  └──────────┘                    │ Controller        │  │
 │       │                          └─────────┬─────────┘  │
 │       │ Reconciles                         │            │
 │       ▼                                    ▼            │
 │  ┌───────────────────────────────────────────────┐      │
 │  │ HealthReport CRD (observability.kubemedic.io) │      │
 │  └───────────────────────┬───────────────────────┘      │
 └──────────────────────────┼──────────────────────────────┘
                            │ Reads
                            ▼
              ┌───────────────────────────┐
              │ KubeMedic REST API (:8080)│
              └─────────────┬─────────────┘
                            │ Serves JSON
                            ▼
              ┌───────────────────────────┐
              │ Web Dashboard (React/Vite)│
              └───────────────────────────┘
```

---

## 🚀 Quickstart (Local Development)

### Prerequisites
* **Go**: v1.24+
* **Node.js**: v18+
* **kubectl**: v1.26+
* **Docker / Kind** (optional for local cluster testing)

### Step 1: Run the Backend Operator & REST API
```bash
# Run the controller and REST API server locally against your active kubeconfig
go run ./cmd/main.go --api-bind-address=:8080
```

### Step 2: Run the Web Dashboard
In a separate terminal window:
```bash
cd frontend
npm install
npm run dev
```
Open **`http://localhost:5173`** in your browser to view the KubeMedic Dashboard.

---

## 📊 Viewing Health Reports

You can inspect Pod health reports using three convenient methods:

### 1. Using `kubectl` (Cluster CRDs)
```bash
# List all HealthReports across all namespaces
kubectl get healthreports -A

# Short form
kubectl get hr -n default

# Get detailed diagnostic details for a specific pod
kubectl get healthreport <pod-name> -n <namespace> -o yaml
```

### 2. Using the REST API
```bash
# Get all health reports
curl http://localhost:8080/api/healthreports

# Get health report for a specific pod
curl http://localhost:8080/api/healthreports/default/nginx

# API Health Check
curl http://localhost:8080/api/healthz
```

### 3. Using the Web Dashboard
Visit `http://localhost:5173` to interactively filter, search, and view detailed recommendations for any pod in your cluster.

---

## 📦 Deployment to a Remote Cluster

### Option 1: Standard Kubebuilder / Kustomize Deployment

1. **Set your remote cluster context**:
   ```bash
   export KUBECONFIG=/path/to/remote-cluster.kubeconfig
   ```

2. **Build and push the container image**:
   ```bash
   export IMG=your-dockerhub-username/kubemedic:v1.0.0
   make docker-build docker-push IMG=$IMG
   ```

3. **Install CRDs and Deploy the Manager**:
   ```bash
   make install
   make deploy IMG=$IMG
   ```

4. **Verify the Deployment**:
   ```bash
   kubectl get pods -n kubemedic-system
   ```

---

### Option 2: Single-File YAML Bundle

Generate a single `dist/install.yaml` manifest that contains all CRDs, RBAC roles, and Deployment configs:

```bash
# 1. Build the installer manifest
export IMG=your-dockerhub-username/kubemedic:v1.0.0
make build-installer IMG=$IMG

# 2. Apply to any cluster
kubectl apply -f dist/install.yaml
```

---

## 🌐 Monitoring Another Cluster

To monitor a different cluster from your local machine:
```bash
# Set KUBECONFIG to the target cluster
export KUBECONFIG=/path/to/target-cluster.kubeconfig

# Install CRDs on target cluster
make install

# Run KubeMedic
go run ./cmd/main.go --api-bind-address=:8080
```

---

## 🧪 Testing & Quality Assurance

```bash
# Run unit tests (uses envtest: real K8s API + etcd)
make test

# Run code linters
make lint

# Run e2e tests (requires Kind)
make test-e2e
```

---

## 📁 Repository Structure

```
├── api/v1alpha1/           # HealthReport CRD API schema definitions
├── cmd/main.go              # Manager entrypoint (Registers controllers & REST API)
├── config/                  # Kustomize deployment manifests & CRDs
├── frontend/                # React + TypeScript + Vite Dashboard
│   ├── src/components/      # UI components (Header, Table, PodDetails, Filters)
│   ├── src/services/        # API service and response normalization
│   └── src/types/           # TypeScript interfaces
├── internal/api/            # REST API server & HTTP handlers
├── internal/controller/     # HealthReport reconciler and pod diagnosis logic
├── Makefile                 # Build, test, and deployment automation
└── README.md                # Project documentation
```

---

## 📄 License

Copyright 2026. Licensed under the Apache License, Version 2.0.
