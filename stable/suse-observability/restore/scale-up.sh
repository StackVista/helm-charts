#!/usr/bin/env bash
set -Eeuo pipefail

echo "=== Scaling up deployments for pods that connect to StackGraph"
for deployment in $(kubectl get deployment --selector=stackstate.com/connects-to-stackgraph=true --output=name)
do
    replicas=$(kubectl get "${deployment}" --output=jsonpath="{.metadata.labels.stackstate\.com/replicas}")
    kubectl scale --replicas="${replicas}" "${deployment}"
done
