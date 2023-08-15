#!/bin/bash

set -exo pipefail

: "${CLOUD_PROVIDER:?CLOUD_PROVIDER is not set}"
: "${BUCKET:?BUCKET is not set}"
: "${PREFIX:?PREFIX is not set}"

case ${CLOUD_PROVIDER} in
    "aws")
        : "${AWS_REGION:?AWS_REGION is not set}"
        ;;
    "gcp")
        ;;
    *)
        echo "CLOUD_PROVIDER must be either 'aws' or 'gcp'"
        exit 1
        ;;
esac

IMAGE_LIST=/tmp/image_list-$(date +%Y%m%d)

# Collect image list from the pods
kubectl get pods -A -o json | jq --arg namespace_exclude_by_name_regex "${IGNORE_NAMESPACE_REGEX}" --arg pod_exclude_by_name_regex "${IGNORE_RESOURCE_NAME_REGEX}"  -r '.items[] | if $namespace_exclude_by_name_regex != "" then select(.metadata.namespace | test($namespace_exclude_by_name_regex) | not) else . end | if $pod_exclude_by_name_regex != "" then select(.metadata.name |test($pod_exclude_by_name_regex)|not) else . end | .spec.containers[].image, .spec.initContainers[]?.image' | sed 's#@.*$##g'  > "${IMAGE_LIST}_pods"

# Collect image list from the application controllers
kubectl get deployments,statefulsets,daemonsets -A -o json | jq --arg namespace_exclude_by_name_regex "${IGNORE_NAMESPACE_REGEX}" --arg apps_exclude_by_name_regex "${IGNORE_RESOURCE_NAME_REGEX}"  -r '.items[] | if $namespace_exclude_by_name_regex != "" then select(.metadata.namespace | test($namespace_exclude_by_name_regex) | not) else . end | if $apps_exclude_by_name_regex != "" then select(.metadata.name |test($apps_exclude_by_name_regex)|not) else . end | .spec.template.spec.containers[].image, .spec.template.spec.initContainers[]?.image' | sed 's#@.*$##g'  > "${IMAGE_LIST}_apps"


# Collect image list from the jobs
kubectl get jobs -A -o json | jq --arg namespace_exclude_by_name_regex "${IGNORE_NAMESPACE_REGEX}" --arg jobs_exclude_by_name_regex "${IGNORE_RESOURCE_NAME_REGEX}"  -r '.items[] | if $namespace_exclude_by_name_regex != "" then select(.metadata.namespace | test($namespace_exclude_by_name_regex) | not) else . end | if $jobs_exclude_by_name_regex != "" then select(.metadata.name |test($jobs_exclude_by_name_regex)|not) else . end |.spec.template.spec.containers[].image, .spec.template.spec.initContainers[]?.image' | sed 's#@.*$##g'  > "${IMAGE_LIST}_jobs"

# Collect image list from the cronjobs
kubectl get cronjobs -A -o json | jq --arg namespace_exclude_by_name_regex "${IGNORE_NAMESPACE_REGEX}" --arg jobs_exclude_by_name_regex "${IGNORE_RESOURCE_NAME_REGEX}"  -r '.items[] | if $namespace_exclude_by_name_regex != "" then select(.metadata.namespace | test($namespace_exclude_by_name_regex) | not) else . end | if $jobs_exclude_by_name_regex != "" then select(.metadata.name |test($jobs_exclude_by_name_regex)|not) else . end |.spec.jobTemplate.spec.template.spec.containers[].image, .spec.jobTemplate.spec.template.spec.initContainers[]?.image' | sed 's#@.*$##g'  > "${IMAGE_LIST}_cronjobs"

sort "${IMAGE_LIST}_pods" "${IMAGE_LIST}_apps" "${IMAGE_LIST}_jobs" "${IMAGE_LIST}_jobs" | uniq > "${IMAGE_LIST}"

case ${CLOUD_PROVIDER} in
    "aws")
        aws s3 cp "${IMAGE_LIST}" "s3://${BUCKET}${PREFIX}$(basename "${IMAGE_LIST}")"
        ;;
    "gcp")
        gsutil cp "${IMAGE_LIST}" "gs://${BUCKET}${PREFIX}$(basename "${IMAGE_LIST}")"
        ;;
esac
