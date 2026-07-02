#!/usr/bin/env bash
set -euo pipefail

echo "Starting port-forwards. Stop with Ctrl+C."

echo "Orders API: http://localhost:8080"
echo "Grafana:    http://localhost:3000"
echo "Prometheus: http://localhost:9090"
echo "Alertmgr:   http://localhost:9093"

kubectl port-forward svc/orders-service -n sre-lab 8080:8080 &
kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80 &
kubectl port-forward svc/monitoring-kube-prometheus-prometheus -n monitoring 9090:9090 &
kubectl port-forward svc/monitoring-kube-prometheus-alertmanager -n monitoring 9093:9093 &
wait
