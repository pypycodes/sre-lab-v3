#!/usr/bin/env bash
set -euo pipefail

k3d cluster delete sre-lab >/dev/null 2>&1 || true
k3d cluster create sre-lab --agents 2 --port "8080:30080@loadbalancer"
kubectl get nodes
