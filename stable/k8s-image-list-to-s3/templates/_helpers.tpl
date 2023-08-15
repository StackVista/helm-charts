{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-image-list-to-s3.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "k8s-image-list-to-s3.fullname" -}}
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
{{- define "k8s-image-list-to-s3.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "k8s-image-list-to-s3.labels" -}}
app.kubernetes.io/name: {{ include "k8s-image-list-to-s3.name" . }}
helm.sh/chart: {{ include "k8s-image-list-to-s3.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Required environment variables
*/}}
{{- define "k8s-image-list-to-s3.env-vars" -}}
{{- if .Values.reports.aws.enabled -}}
- name: CLOUD_PROVIDER
  value: "aws"
- name: BUCKET
  value: {{ required ".Values.reports.aws.bucket.name has to be set" .Values.reports.aws.bucket.name }}
- name: AWS_REGION
  value: {{ required ".Values.reports.aws.bucket.region has to be set" .Values.reports.aws.bucket.region }}
- name: PREFIX
  value: {{ .Values.reports.aws.bucket.prefix }}
{{- else if .Values.reports.gcp.enabled -}}
- name: CLOUD_PROVIDER
  value: "gcp"
- name: BUCKET
  value: {{ required ".Values.reports.gcp.bucket.name has to be set" .Values.reports.gcp.bucket.name }}
- name: PREFIX
  value: {{ .Values.reports.gcp.bucket.prefix }}
{{- end }}
- name: IGNORE_NAMESPACE_REGEX
  value: {{ .Values.scan.ignoreNamespaceRegex | quote }}
- name: IGNORE_RESOURCE_NAME_REGEX
  value: {{ .Values.scan.ignoreResourceNameRegex | quote }}
{{- end -}}
