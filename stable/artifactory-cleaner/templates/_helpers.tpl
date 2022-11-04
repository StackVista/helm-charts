{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "artifactory-cleaner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "artifactory-cleaner.fullname" -}}
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
{{- define "artifactory-cleaner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "artifactory-cleaner.labels" -}}
app.kubernetes.io/name: {{ include "artifactory-cleaner.name" . }}
helm.sh/chart: {{ include "artifactory-cleaner.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Required environment variables
*/}}
{{- define "artifactory-cleaner.env-vars" -}}
- name: ARTIFACTORY_URL
  value: {{ .Values.artifactory.url }}
- name: ARTIFACTORY_USER
  value: {{ .Values.artifactory.user }}
- name: ARTIFACTORY_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "artifactory-cleaner.fullname" . }}
      key: ARTIFACTORY_PASSWORD
- name: RETENTION_MANIFEST
  value: /scripts/retentions.json
{{- end -}}
