# Runbook - High Error Rate

## Alert
`OrdersHighErrorRate` or `OrdersErrorBudgetFastBurn`

## Impact
Users may experience failed API requests against the Orders service.

## First 5 Minutes

```bash
kubectl get pods -n sre-lab
kubectl logs -n sre-lab deploy/orders-service --tail=100
kubectl get events -n sre-lab --sort-by=.lastTimestamp
```

## Prometheus Checks

```promql
sli:orders_error_rate:ratio_rate5m
slo:orders_availability:burn_rate5m
sum(rate(http_requests_total{namespace="sre-lab", status=~"5.."}[5m]))
```

## Mitigation

1. Confirm if errors are caused by `/error` traffic or real application failures.
2. Stop error load test if running.
3. Restart deployment only if pods are unhealthy.
4. Scale service if saturation is observed.

```bash
kubectl rollout restart deployment/orders-service -n sre-lab
kubectl scale deployment/orders-service -n sre-lab --replicas=3
```

## Recovery Validation

```promql
sli:orders_availability:ratio_rate5m
sli:orders_error_rate:ratio_rate5m
```

## Capture

- Detection time
- Acknowledgement time
- Recovery time
- Root cause
- Preventive action
