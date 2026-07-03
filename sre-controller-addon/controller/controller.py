import os
import time
import traceback
from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException

GROUP = "demo.ct.com"
VERSION = "v1"
PLURAL = "sreservices"


def load_kube_config():
    try:
        config.load_incluster_config()
        print("Loaded in-cluster Kubernetes config")
    except Exception:
        config.load_kube_config()
        print("Loaded local kubeconfig")


def desired_configmap_body(namespace, name, spec, owner_uid):
    cm_name = f"sreservice-{name}-config"
    labels = {
        "app.kubernetes.io/name": "sreservice-controller",
        "demo.ct.com/sreservice": name,
    }
    return cm_name, {
        "apiVersion": "v1",
        "kind": "ConfigMap",
        "metadata": {
            "name": cm_name,
            "namespace": namespace,
            "labels": labels,
            "ownerReferences": [
                {
                    "apiVersion": f"{GROUP}/{VERSION}",
                    "kind": "SREService",
                    "name": name,
                    "uid": owner_uid,
                    "controller": True,
                    "blockOwnerDeletion": True,
                }
            ],
        },
        "data": {
            "owner": spec.get("owner", ""),
            "team": spec.get("team", ""),
            "availabilitySLO": spec.get("availabilitySLO", ""),
            "latencySLO": spec.get("latencySLO", ""),
            "errorRateSLO": spec.get("errorRateSLO", ""),
        },
    }


def patch_status(custom_api, namespace, name, generation, phase, message, configmap_name):
    body = {
        "status": {
            "observedGeneration": generation,
            "phase": phase,
            "message": message,
            "configMapName": configmap_name,
        }
    }
    try:
        custom_api.patch_namespaced_custom_object_status(
            GROUP, VERSION, namespace, PLURAL, name, body
        )
    except ApiException as e:
        print(f"Failed to patch status for {namespace}/{name}: {e}")


def reconcile(custom_api, core_api, obj):
    metadata = obj.get("metadata", {})
    spec = obj.get("spec", {})
    name = metadata["name"]
    namespace = metadata.get("namespace", "default")
    uid = metadata["uid"]
    generation = metadata.get("generation", 1)

    cm_name, cm_body = desired_configmap_body(namespace, name, spec, uid)

    try:
        existing = core_api.read_namespaced_config_map(cm_name, namespace)
        existing.data = cm_body["data"]
        existing.metadata.labels = cm_body["metadata"]["labels"]
        core_api.replace_namespaced_config_map(cm_name, namespace, existing)
        print(f"Updated ConfigMap {namespace}/{cm_name}")
    except ApiException as e:
        if e.status == 404:
            core_api.create_namespaced_config_map(namespace, cm_body)
            print(f"Created ConfigMap {namespace}/{cm_name}")
        else:
            raise

    patch_status(
        custom_api,
        namespace,
        name,
        generation,
        "Ready",
        "SREService reconciled successfully",
        cm_name,
    )


def delete_owned_configmap(core_api, obj):
    metadata = obj.get("metadata", {})
    name = metadata["name"]
    namespace = metadata.get("namespace", "default")
    cm_name = f"sreservice-{name}-config"
    try:
        core_api.delete_namespaced_config_map(cm_name, namespace)
        print(f"Deleted ConfigMap {namespace}/{cm_name}")
    except ApiException as e:
        if e.status != 404:
            raise


def main():
    load_kube_config()
    custom_api = client.CustomObjectsApi()
    core_api = client.CoreV1Api()

    namespace = os.getenv("WATCH_NAMESPACE", "")
    print(f"Starting SREService controller. WATCH_NAMESPACE='{namespace or 'all namespaces'}'")

    while True:
        try:
            w = watch.Watch()
            if namespace:
                stream = w.stream(
                    custom_api.list_namespaced_custom_object,
                    GROUP,
                    VERSION,
                    namespace,
                    PLURAL,
                    timeout_seconds=60,
                )
            else:
                stream = w.stream(
                    custom_api.list_cluster_custom_object,
                    GROUP,
                    VERSION,
                    PLURAL,
                    timeout_seconds=60,
                )

            for event in stream:
                event_type = event["type"]
                obj = event["object"]
                name = obj.get("metadata", {}).get("name")
                ns = obj.get("metadata", {}).get("namespace", "default")
                print(f"Event {event_type} for {ns}/{name}")

                if event_type in ["ADDED", "MODIFIED"]:
                    reconcile(custom_api, core_api, obj)
                elif event_type == "DELETED":
                    delete_owned_configmap(core_api, obj)

        except Exception:
            print("Controller loop error:")
            traceback.print_exc()
            time.sleep(5)


if __name__ == "__main__":
    main()
