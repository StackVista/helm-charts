{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "aws-amicleaner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "aws-amicleaner.fullname" -}}
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
{{- define "aws-amicleaner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "aws-amicleaner.labels" -}}
app.kubernetes.io/name: {{ include "aws-amicleaner.name" . }}
helm.sh/chart: {{ include "aws-amicleaner.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
StackState required environment variables
*/}}
{{- define "aws-amicleaner.env-vars" -}}
- name: AWS_CONFIG_FILE
  value: {{ .Values.aws.mountPath }}/config
  {{- if .Values.aws.credentialsFile }}
- name: AWS_SHARED_CREDENTIALS_FILE
  value: {{ .Values.aws.mountPath }}/credentials
  {{- end }}
{{- end -}}

{{/*
Checksum annotations
*/}}
{{- define "aws-amicleaner.checksum-configs" -}}
checksum/config: {{ include (print $.Template.BasePath "/sts-secret.yaml") . | sha256sum }}
{{- end -}}
