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


---

# Implementation Guide - SRE Lab V3

## Goal

Create a complete GitHub-ready SRE platform that demonstrates:

1. Application instrumentation
2. Kubernetes deployment
3. Prometheus scraping
4. Grafana visualization
5. SLI and SLO calculation
6. Error budget burn-rate alerts
7. k6 traffic generation
8. Chaos engineering
9. Incident management and postmortems
10. Optional Dynatrace OTLP integration

---

## Lab 1 - Prepare WSL

```bash
sudo apt update
sudo apt install -y curl wget jq git make docker.io
sudo usermod -aG docker $USER
```

Restart WSL after adding your user to the Docker group:

```powershell
wsl --shutdown
```

Then reopen Ubuntu and verify:

```bash
docker version
```

---

## Lab 2 - Create k3d Cluster

```bash
k3d cluster create sre-lab \
  --agents 2 \
  --port "8080:30080@loadbalancer"
```

Verify:

```bash
kubectl get nodes
```

---

## Lab 3 - Install Monitoring

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack -n monitoring
```

Verify:

```bash
kubectl get pods -n monitoring
```

---

## Lab 4 - Build and Deploy Orders API

```bash
docker build -t orders-service:v1 ./app
k3d image import orders-service:v1 -c sre-lab
kubectl apply -f k8s/
```

Verify:

```bash
kubectl get pods -n sre-lab
kubectl get svc -n sre-lab
```

---

## Lab 5 - Validate Metrics

```bash
kubectl port-forward svc/orders-service -n sre-lab 8080:8080
curl http://localhost:8080/orders
curl http://localhost:8080/error
curl http://localhost:8080/slow
curl http://localhost:8080/metrics
```

Expected metrics:

```text
http_requests_total
http_request_duration_seconds_bucket
orders_created_total
orders_failed_total
inventory_lookup_total
```

---

## Lab 6 - Apply SLO Rules

```bash
kubectl apply -f monitoring/prometheus-slo-rules.yaml
```

PromQL examples:

### Availability

```promql
sli:orders_availability:ratio_rate5m
```

### Error rate

```promql
sli:orders_error_rate:ratio_rate5m
```

### P95 latency

```promql
sli:orders_latency_p95:rate5m
```

### Availability burn rate

```promql
slo:orders_availability:burn_rate5m
```

---

## Lab 7 - Run Load Tests

```bash
k6 run loadtests/normal.js
k6 run loadtests/error-spike.js
k6 run loadtests/latency-spike.js
k6 run loadtests/stress.js
```

---

## Lab 8 - Import Grafana Dashboard

Open Grafana:

```bash
kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80
```

Import dashboard JSON from:

```text
grafana/dashboards/sre-dashboard.json
```

---

## Lab 9 - Run Incident Drill

1. Start normal load.
2. Run error spike.
3. Watch error rate and availability drop.
4. Confirm alert firing in Prometheus/Alertmanager.
5. Use runbook `runbooks/high-error-rate.md`.
6. Record MTTD, MTTA, MTTR.
7. Complete `runbooks/postmortem-template.md`.

---

## Lab 10 - Run Chaos Experiments

Install Chaos Mesh if required, then apply experiments from `chaos/`.

```bash
kubectl apply -f chaos/pod-kill.yaml
kubectl apply -f chaos/network-delay.yaml
kubectl apply -f chaos/cpu-stress.yaml
```

> Only run chaos experiments against the `sre-lab` namespace.

---

## Lab 11 - Optional Dynatrace Integration

Create a Kubernetes secret with your Dynatrace OTLP endpoint and token. Do not commit real tokens.

Example placeholder:

```bash
kubectl create secret generic dynatrace-otel \
  -n observability \
  --from-literal=DT_ENDPOINT="https://YOUR_ENV.live.dynatrace.com/api/v2/otlp" \
  --from-literal=DT_API_TOKEN="YOUR_TOKEN"
```

Then update `otel/collector-config.yaml` to enable the Dynatrace exporter.

---

## Suggested Interview/Portfolio Story

This repository demonstrates how an SRE team can move from basic monitoring to reliability engineering using:

- user-centric SLIs
- business-aligned SLOs
- error budgets
- burn-rate alerting
- automated dashboards
- runbooks
- chaos testing
- incident reviews
