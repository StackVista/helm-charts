{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "python-amicleaner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "python-amicleaner.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "python-amicleaner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "python-amicleaner.labels" -}}
app.kubernetes.io/name: {{ include "python-amicleaner.name" . }}
helm.sh/chart: {{ include "python-amicleaner.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
StackState required environment variables
*/}}
{{- define "python-amicleaner.env-vars" -}}
- name: AWS_CONFIG_FILE
  value: {{ .Values.aws.mountPath }}/config
- name: AWS_SHARED_CREDENTIALS_FILE
  value: {{ .Values.aws.mountPath }}/credentials
{{- end -}}

{{/*
Checksum annotations
*/}}
{{- define "python-amicleaner.checksum-configs" -}}
checksum/config: {{ include (print $.Template.BasePath "/sts-secret.yaml") . | sha256sum }}
{{- end -}}
