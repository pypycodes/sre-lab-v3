#!/usr/bin/env bash
set -euo pipefail

k3d cluster delete sre-lab || true
docker image rm orders-service:v1 || true
docker system df

echo "Cleanup complete. Run docker system prune if you want to reclaim more space."
