# Runbook - High Latency

## Alert
`OrdersHighLatencyP95`

## Impact
Users may experience slow responses from the Orders service.

## Prometheus Checks

```promql
sli:orders_latency_p95:rate5m
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{namespace="sre-lab"}[5m])) by (le))
sum(rate(http_requests_total{namespace="sre-lab"}[5m]))
```

## Kubernetes Checks

```bash
kubectl top pods -n sre-lab
kubectl describe deployment orders-service -n sre-lab
kubectl get hpa -n sre-lab
```

## Mitigation

1. Check if `/slow` endpoint load test is running.
2. Check CPU/memory saturation.
3. Scale replicas if needed.
4. Review recent chaos experiments.

```bash
kubectl scale deployment/orders-service -n sre-lab --replicas=5
```

## Recovery Validation

P95 latency should return below 300ms.
