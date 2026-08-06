{{/*
Expand the name of the chart.
*/}}
{{- define "codeartifact-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "codeartifact-proxy.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "codeartifact-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "codeartifact-proxy.labels" -}}
helm.sh/chart: {{ include "codeartifact-proxy.chart" . }}
{{ include "codeartifact-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "codeartifact-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "codeartifact-proxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Name of the StatefulSet's cache volumeClaimTemplate. Renaming it orphans every existing PVC (and
so every warm cache), because a StatefulSet's volumeClaimTemplates are immutable after creation.
*/}}
{{- define "codeartifact-proxy.cacheVolumeName" -}}
nginx-cache
{{- end }}

{{/*
Resolve a codeartifact.endpoints.<key>, failing at template time when it is unset or still a
REPLACE-WITH-TERRAFORM-OUTPUT placeholder — either renders as valid YAML and would only fail once
the pod is serving traffic.
*/}}
{{- define "codeartifact-proxy.endpoint" -}}
{{- $key := .key -}}
{{- $value := index .root.Values.codeartifact.endpoints $key -}}
{{- $terraformOutput := index (dict "mavenReleases" "codeartifact_packages_maven_endpoint" "mavenSnapshots" "codeartifact_packages_snapshot_maven_endpoint" "pypi" "codeartifact_packages_pypi_endpoint") $key -}}
{{- if or (empty $value) (contains "REPLACE-WITH-TERRAFORM-OUTPUT" $value) -}}
{{- fail (printf "codeartifact.endpoints.%s must be set to the terraform-infra output `%s` (STAC-25404); it is currently unset or still a placeholder" $key $terraformOutput) -}}
{{- end -}}
{{- $value -}}
{{- end }}

{{/*
Build a digest-pinned image reference, failing at template time on an unset or `sha256:0000…`
placeholder digest rather than at image pull.
*/}}
{{- define "codeartifact-proxy.digest" -}}
{{- $image := .image -}}
{{- if or (empty $image.digest) (hasPrefix "sha256:0000" $image.digest) -}}
{{- fail (printf "%s must be set to the digest published for %s/%s in the SUSE Application Collection; it is currently unset or still a placeholder" .path $image.registry $image.repository) -}}
{{- end -}}
{{- printf "%s/%s@%s" $image.registry $image.repository $image.digest -}}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "codeartifact-proxy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "codeartifact-proxy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
