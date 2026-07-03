#!/usr/bin/env bash
set -euo pipefail
kubectl delete -f samples/orders-sreservice.yaml --ignore-not-found
kubectl delete -f k8s/controller-deployment.yaml --ignore-not-found
kubectl delete -f k8s/rbac.yaml --ignore-not-found
kubectl delete -f crd/sreservice-crd.yaml --ignore-not-found
