#!/usr/bin/env bash
set -euo pipefail

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
helm repo update
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack -n monitoring
kubectl rollout status deployment/monitoring-grafana -n monitoring --timeout=180s
kubectl apply -f otel/collector-config.yaml
kubectl apply -f monitoring/prometheus-slo-rules.yaml
