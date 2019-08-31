{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "gitlab-steward.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "gitlab-steward.fullname" -}}
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
{{- define "gitlab-steward.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "gitlab-steward.labels" -}}
app.kubernetes.io/name: {{ include "gitlab-steward.name" . }}
helm.sh/chart: {{ include "gitlab-steward.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Environment variables
*/}}
{{- define "gitlab-steward.envVars" }}
- name: STEWARD_DRY_RUN
  value: {{ .Values.steward.dryRun | quote }}
- name: STEWARD_GITLAB_API_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ include "gitlab-steward.fullname" . }}
      key: steward-gitlab-api-token
  {{- if .Values.steward.logLevel }}
- name: STEWARD_LOG_LEVEL
  value: {{ .Values.steward.logLevel | quote }}
  {{- end }}
- name: STEWARD_MAX_DURATION
  value: {{ .Values.steward.maxDuration | quote }}
- name: STEWARD_STACKSTATE_PROJECT
  value: {{ .Values.steward.stackstateProject | quote }}
  {{- if .Values.extraEnv.open }}
    {{- range $key, $value := .Values.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
    {{- end }}
  {{- end }}
  {{- if .Values.extraEnv.secret }}
    {{- range $key, $value := .Values.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ include "gitlab-steward.fullname" $ }}
      key: {{ $key }}
    {{- end }}
  {{- end }}
{{- end }}
