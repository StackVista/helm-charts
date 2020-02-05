{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "iceman.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "iceman.fullname" -}}
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
{{- define "iceman.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "iceman.labels" -}}
app.kubernetes.io/name: {{ include "iceman.name" . }}
helm.sh/chart: {{ include "iceman.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Environment variables
*/}}
{{- define "iceman.envVars" }}
- name: AWS_DEFAULT_REGION
  value: {{ .Values.iceman.awsRegion | quote }}
- name: ICEMAN_LOG_LEVEL
  value: {{ .Values.iceman.logLevel | quote }}
- name: ICEMAN_S3_BUCKET
  value: {{ required "An entry for .Values.iceman.s3Bucket is required!" .Values.iceman.s3Bucket | quote }}
- name: ICEMAN_STACKSTATE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "iceman.fullname" . }}
      key: iceman-stackstate-password
- name: ICEMAN_STACKSTATE_USERNAME
  valueFrom:
    secretKeyRef:
      name: {{ include "iceman.fullname" . }}
      key: iceman-stackstate-username
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
      name: {{ include "iceman.fullname" $ }}
      key: {{ $key }}
    {{- end }}
  {{- end }}
{{- end }}
