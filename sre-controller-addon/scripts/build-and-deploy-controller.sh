#!/usr/bin/env bash
set -euo pipefail

docker build -t sreservice-controller:v1 ./controller
k3d image import sreservice-controller:v1 -c sre-lab
kubectl apply -f crd/sreservice-crd.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/controller-deployment.yaml
kubectl rollout status deployment/sreservice-controller -n sre-lab --timeout=180s
kubectl apply -f samples/orders-sreservice.yaml
kubectl get sreservices -n sre-lab
kubectl get configmap -n sre-lab | grep sreservice || true
