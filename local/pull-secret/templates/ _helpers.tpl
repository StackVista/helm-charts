{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "pullsecret.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "pullsecret.fullname" -}}
  {{- $base := default (printf "%s-%s" .Release.Name .Chart.Name) .Values.fullnameOverride -}}
  {{- $base | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "pullsecret.fullname.short"}}
  {{- $global := default (dict) .Values.global -}}
  {{- $base := .Chart.Name -}}
  {{- if .Values.fullnameOverride -}}
    {{- $base = .Values.fullnameOverride -}}
  {{- else if ne $base .Release.Name -}}
    {{- $base = (printf "%s-%s" .Release.Name .Chart.Name) -}}
  {{- end -}}
  {{- $base | lower | trunc 63 | trimSuffix "-" -}}
{{- end -}}
