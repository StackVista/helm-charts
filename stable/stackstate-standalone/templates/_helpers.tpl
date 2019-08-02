{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "stackstate-standalone.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "stackstate-standalone.fullname" -}}
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
{{- define "stackstate-standalone.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "stackstate-standalone.labels" -}}
app.kubernetes.io/name: {{ include "stackstate-standalone.name" . }}
helm.sh/chart: {{ include "stackstate-standalone.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
GitLab annotations
*/}}
{{- define "gitlab.annotations" -}}
{{- if .Values.gitlab.app }}
app.gitlab.com/app: {{ .Values.gitlab.app | quote }}
{{- end }}
{{- if .Values.gitlab.env }}
app.gitlab.com/env: {{ .Values.gitlab.env | quote }}
{{- end }}
{{- end -}}

{{/*
Demo data lifecycle hook
*/}}
{{- define "stackstate-standalone.demoData.lifecycle" -}}
{{- if .Values.stackstate.demoData.enabled }}
lifecycle:
  postStart:
    exec:
      command: ["/bin/bash", "/sts/stackstate/bin/branch-deploy.sh"]
{{- end }}
{{- end -}}

{{/*
StackState required environment variables
*/}}
{{- define "stackstate-standalone.requiredEnvVars" }}
- name: API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "stackstate-standalone.fullname" . }}
      key: sts-receiver-api-key
- name: LICENSE_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "stackstate-standalone.fullname" . }}
      key: sts-license-key
- name: RECEIVER_BASE_URL
  value: {{ .Values.stackstate.receiver.baseUrl | quote }}
{{- end }}

{{/*
Checksum annotations
*/}}
{{- define "stackstate-standalone.checksum-configs" }}
checksum/config: {{ include (print $.Template.BasePath "/sts-standalone-secret.yaml") . | sha256sum }}
{{- end }}
