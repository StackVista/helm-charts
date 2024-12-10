#!/bin/bash

set -ex

echo "Removing router mode configmap"

# shellcheck disable=SC2140
kubectl delete configmap -n "{{ .Release.Namespace }}" "{{ template "common.fullname.short" . }}-router-automatic" || true

echo "Configmap removed"
