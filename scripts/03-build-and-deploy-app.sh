#!/usr/bin/env bash
set -euo pipefail

docker build -t orders-service:v1 ./app
k3d image import orders-service:v1 -c sre-lab
kubectl apply -f k8s/
kubectl rollout status deployment/orders-service -n sre-lab --timeout=180s
kubectl get pods -n sre-lab
