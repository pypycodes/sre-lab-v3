# SREService Controller Add-on

This add-on teaches how a Kubernetes **controller/operator** works.

## What this controller does

When you create a custom resource like:

```yaml
apiVersion: demo.ct.com/v1
kind: SREService
metadata:
  name: orders
  namespace: sre-lab
spec:
  owner: pavan
  team: platform
  availabilitySLO: "99.9"
  latencySLO: "300ms"
```

The controller watches it and automatically creates/updates this ConfigMap:

```text
sreservice-orders-config
```

The ConfigMap contains the SRE metadata from the custom resource.

This demonstrates the controller loop:

```text
Watch custom resource → Reconcile desired state → Create/update Kubernetes resource
```

## Install

```bash
kubectl apply -f crd/sreservice-crd.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/controller-deployment.yaml
kubectl apply -f samples/orders-sreservice.yaml
```

## Verify

```bash
kubectl get sreservices -n sre-lab
kubectl get configmap -n sre-lab | grep sreservice
kubectl get configmap sreservice-orders-config -n sre-lab -o yaml
kubectl logs -n sre-lab deploy/sreservice-controller -f
```

## Cleanup

```bash
kubectl delete -f samples/orders-sreservice.yaml
kubectl delete -f k8s/controller-deployment.yaml
kubectl delete -f k8s/rbac.yaml
kubectl delete -f crd/sreservice-crd.yaml
```
