{{/*
Expand the name of the chart.
*/}}
{{- define "suse-observability-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "suse-observability-operator.fullname" -}}
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
{{- define "suse-observability-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "suse-observability-operator.labels" -}}
helm.sh/chart: {{ include "suse-observability-operator.chart" . }}
{{ include "suse-observability-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "suse-observability-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "suse-observability-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Returns the image registry
{{ include "common.image.registry" ( dict "image" . "context" $) }}
*/}}
{{- define "common.image.registry" -}}
  {{- if .context.Values.global }}
    {{- .image.registry | default .context.Values.global.imageRegistry -}}
  {{- else -}}
    {{- .image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Returns an affinity configuration, it use global affinity and then overrides it with the more specific for given resource
{{include "merge.affinity"  ( dict "affinity" .Values.clickhouse.affinity "context" $) }}
*/}}
{{- define "merge.affinity" -}}
affinity:
  nodeAffinity: {{ .affinity.nodeAffinity | default .context.Values.global.affinity.nodeAffinity | toYaml | nindent 4 }}
  podAffinity: {{ .affinity.podAffinity | default .context.Values.global.affinity.podAffinity | toYaml | nindent 4 }}
  podAntiAffinity: {{ .affinity.podAntiAffinity | default .context.Values.global.affinity.podAntiAffinity | toYaml | nindent 4 }}
{{- end -}}

{{/*
Return the image
{{ include "common.image" ( dict "image" . "context" $) }}
*/}}
{{- define "common.image" -}}
  {{ include "common.image.registry" $ }}/{{ .image.repository }}:{{ .image.tag }}
{{- end -}}

{{- /*
  Adding a trailing slash to a value if it is not empty.
*/ -}}
{{- define "ensureTrailingSlashIfNotEmpty" -}}
  {{- if . -}}
    {{- printf "%s/" (. | trimSuffix "/") -}}
  {{- else -}}
    {{- "" -}}
  {{- end -}}
{{- end -}}

{{/*
Return ttlSecondsAfterFinished. We make this a very high value for argo so failures cannot be silently ignored.
*/}}
{{- define "stackstate.job.ttlSecondsAfterFinished" -}}
{{- if .Values.deployment.compatibleWithArgoCD }}86400{{- else }}600{{- end -}}
{{- end -}}
