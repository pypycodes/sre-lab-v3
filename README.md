# SRE Lab V3 - GitHub Ready Implementation

Production-style hands-on SRE lab runnable inside **WSL Ubuntu** using:

- Go Orders API
- Docker
- k3d / Kubernetes
- kube-prometheus-stack
- Prometheus metrics
- Grafana dashboards
- Alertmanager + SLO burn-rate alerts
- OpenTelemetry Collector
- Dynatrace OTLP forwarding option
- k6 load generation
- Chaos Mesh experiments
- Incident runbooks and postmortem template

> This project is intentionally designed as a local SRE playground for learning SLIs, SLOs, error budgets, burn-rate alerting, incident response, and chaos engineering.

---

## 1. Repository Structure

```text
sre-lab-v3/
├── app/
│   ├── cmd/server/main.go
│   ├── go.mod
│   └── Dockerfile
├── k8s/
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── servicemonitor.yaml
├── monitoring/
│   ├── prometheus-slo-rules.yaml
│   └── alertmanager-config.yaml
├── otel/
│   └── collector-config.yaml
├── grafana/
│   └── dashboards/sre-dashboard.json
├── loadtests/
│   ├── normal.js
│   ├── error-spike.js
│   ├── latency-spike.js
│   └── stress.js
├── chaos/
│   ├── pod-kill.yaml
│   ├── network-delay.yaml
│   └── cpu-stress.yaml
├── runbooks/
│   ├── high-error-rate.md
│   ├── high-latency.md
│   └── postmortem-template.md
├── scripts/
│   ├── 00-install-tools.sh
│   ├── 01-create-cluster.sh
│   ├── 02-install-monitoring.sh
│   ├── 03-build-and-deploy-app.sh
│   ├── 04-port-forward.sh
│   └── 99-cleanup.sh
└── docs/
    └── IMPLEMENTATION_GUIDE.md
```

---

## 2. Quick Start

Run from WSL Ubuntu:

```bash
chmod +x scripts/*.sh
./scripts/00-install-tools.sh
./scripts/01-create-cluster.sh
./scripts/02-install-monitoring.sh
./scripts/03-build-and-deploy-app.sh
```

Open separate terminals for port-forwarding:

```bash
./scripts/04-port-forward.sh
```

Access:

```text
Orders API:  http://localhost:8080/orders
Metrics:     http://localhost:8080/metrics
Grafana:     http://localhost:3000
Prometheus:  http://localhost:9090
Alertmanager:http://localhost:9093
```

Grafana credentials:

```bash
kubectl get secret monitoring-grafana -n monitoring -o jsonpath='{.data.admin-password}' | base64 -d; echo
```

Username:

```text
admin
```

---

## 3. Generate Load

Install k6 if not already installed, then run:

```bash
k6 run loadtests/normal.js
k6 run loadtests/error-spike.js
k6 run loadtests/latency-spike.js
k6 run loadtests/stress.js
```

---

## 4. Core SRE Learning Goals

### SLIs

- Availability
- Error rate
- Latency P95/P99
- Traffic / throughput

### SLOs

- Availability >= 99.9%
- Error rate < 0.1%
- 95% requests under 300ms

### Error Budget

For 99.9% availability:

```text
Allowed failure = 0.1%
```

### Burn Rate

Burn rate shows how quickly the error budget is being consumed.

---

## 5. Important Safety Notes

- This is a local lab only.
- Do not commit real Dynatrace tokens.
- Do not run stress tests against production systems.
- Chaos experiments should only run against this local lab namespace.
