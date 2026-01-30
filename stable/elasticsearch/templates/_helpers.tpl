{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "elasticsearch.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "elasticsearch.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "elasticsearch.uname" -}}
{{- if empty .Values.fullnameOverride -}}
{{- if empty .Values.nameOverride -}}
{{ .Values.clusterName }}-{{ .Values.nodeGroup }}
{{- else -}}
{{ .Values.nameOverride }}-{{ .Values.nodeGroup }}
{{- end -}}
{{- else -}}
{{ .Values.fullnameOverride }}
{{- end -}}
{{- end -}}

{{/*
Return the image registry
*/}}
{{- define "elasticsearch.imageRegistry" -}}
  {{- if .Values.global }}
    {{- .Values.global.imageRegistry | default .Values.imageRegistry -}}
  {{- else -}}
    {{- .Values.imageRegistry -}}
  {{- end -}}
{{- end -}}

{{/*
Generate certificates when the secret doesn't exist
*/}}
{{- define "elasticsearch.gen-certs" -}}
{{- $certs := lookup "v1" "Secret" .Release.Namespace ( printf "%s-certs" (include "elasticsearch.uname" . ) ) -}}
{{- if $certs -}}
tls.crt: {{ index $certs.data "tls.crt" }}
tls.key: {{ index $certs.data "tls.key" }}
ca.crt: {{ index $certs.data "ca.crt" }}
{{- else -}}
{{- $altNames := list ( include "elasticsearch.masterService" . ) ( printf "%s.%s" (include "elasticsearch.masterService" .) .Release.Namespace ) ( printf "%s.%s.svc" (include "elasticsearch.masterService" .) .Release.Namespace ) -}}
{{- $ca := genCA "elasticsearch-ca" 365 -}}
{{- $cert := genSignedCert ( include "elasticsearch.masterService" . ) nil $altNames 365 $ca -}}
tls.crt: {{ $cert.Cert | toString | b64enc }}
tls.key: {{ $cert.Key | toString | b64enc }}
ca.crt: {{ $ca.Cert | toString | b64enc }}
{{- end -}}
{{- end -}}

{{- define "elasticsearch.masterService" -}}
{{- if empty .Values.masterService -}}
{{- if empty .Values.fullnameOverride -}}
{{- if empty .Values.nameOverride -}}
{{ .Values.clusterName }}-master
{{- else -}}
{{ .Values.nameOverride }}-master
{{- end -}}
{{- else -}}
{{ .Values.fullnameOverride }}
{{- end -}}
{{- else -}}
{{ .Values.masterService }}
{{- end -}}
{{- end -}}

{{- define "elasticsearch.endpoints" -}}
{{- $sizingReplicas := include "common.sizing.elasticsearch.replicas" . | trim -}}
{{- $replicas := 0 -}}
{{- if $sizingReplicas -}}
  {{- $replicas = int $sizingReplicas -}}
{{- else -}}
  {{- $replicas = int (toString (.Values.replicas)) -}}
{{- end -}}
{{- $uname := printf "%s-%s" .Values.clusterName .Values.nodeGroup }}
  {{- range $i, $e := untilStep 0 $replicas 1 -}}
{{ $uname }}-{{ $i }},
  {{- end -}}
{{- end -}}

{{- define "elasticsearch.roles" -}}
{{- range $.Values.roles -}}
{{ . }},
{{- end -}}
{{- end -}}

{{- define "elasticsearch.esMajorVersion" -}}
{{- if .Values.esMajorVersion -}}
{{ .Values.esMajorVersion }}
{{- else -}}
{{- $version := int (index (.Values.imageTag | splitList ".") 0) -}}
  {{- if and (contains "elasticsearch/elasticsearch" .Values.imageRepository) (not (eq $version 0)) -}}
{{ $version }}
  {{- else -}}
8
  {{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for statefulset.
*/}}
{{- define "elasticsearch.statefulset.apiVersion" -}}
{{- if semverCompare "<1.9-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "apps/v1beta2" -}}
{{- else -}}
{{- print "apps/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for ingress.
*/}}
{{- define "elasticsearch.ingress.apiVersion" -}}
{{- if semverCompare "<1.14-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "extensions/v1beta1" -}}
{{- else if semverCompare "<1.19-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "networking.k8s.io/v1beta1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Recommended labels
*/}}
{{- define "elasticsearch.labels.recommended" -}}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/name: {{ template "elasticsearch.name" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "elasticsearch.labels.selector" -}}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/name: {{ template "elasticsearch.name" . }}
{{- end -}}


{{/*
Common labels
*/}}
{{- define "elasticsearch.labels.common" -}}
{{- range $key, $value := .Values.commonLabels }}
{{ $key }}: {{ $value | quote }}
{{- end -}}
{{- end -}}

{{/*
Merge global and local common labels.
Precedence order: .Values.commonLabels (higher) -> .Values.global.commonLabels (lower)
*/}}
{{- define "elasticsearch.commonLabels" -}}
{{- $labels := dict }}
{{- if .Values.commonLabels }}
{{- $labels = .Values.commonLabels }}
{{- end }}
{{- if .Values.global.commonLabels }}
{{- $labels = merge $labels .Values.global.commonLabels }}
{{- end }}
{{- range $key, $value := $labels }}
{{ $key }}: {{ $value | quote }}
{{- end -}}
{{- end -}}

{{/*
Return the proper Elasticsearch pod labels
Merges elasticsearch.commonLabels with podLabels, with podLabels taking precedence
Precedence order: .Values.podLabels (highest) -> .Values.commonLabels (middle) -> .Values.global.commonLabels (lowest)
*/}}
{{- define "elasticsearch.podLabels" -}}
{{- $labels := dict }}
{{- if .Values.podLabels }}
{{- $labels = .Values.podLabels }}
{{- end }}
{{- if .Values.commonLabels }}
{{- $labels = merge $labels .Values.commonLabels }}
{{- end }}
{{- if .Values.global.commonLabels }}
{{- $labels = merge $labels .Values.global.commonLabels }}
{{- end }}
{{- range $key, $value := $labels }}
{{ $key }}: {{ $value | quote }}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains a template.
Usage:
{{ include "elasticsearch.tplvalue.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "elasticsearch.tplvalue.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names evaluating values as templates
{{ include "elasticsearch.image.pullSecret.name" ( dict "images" (list .Values.path.to.the.image1, .Values.path.to.the.image2) "context" $) }}
*/}}
{{- define "elasticsearch.image.pullSecret.name" -}}
  {{- $pullSecrets := list }}
  {{- $context := .context }}

  {{- if and $context.Values.global $context.Values.global.suseObservability $context.Values.global.suseObservability.pullSecret $context.Values.global.suseObservability.pullSecret.username $context.Values.global.suseObservability.pullSecret.password -}}
    {{- $pullSecrets = append $pullSecrets "suse-observability-pull-secret" -}}
  {{- else -}}
    {{- if $context.Values.global }}
      {{- range $context.Values.global.imagePullSecrets -}}
        {{/* Is plain array of strings, compatible with all bitnami charts */}}
        {{- $pullSecrets = append $pullSecrets (include "elasticsearch.tplvalue.render" (dict "value" . "context" $context)) -}}
      {{- end -}}
    {{- end -}}
    {{- range $context.Values.imagePullSecrets -}}
      {{- $pullSecrets = append $pullSecrets (include "elasticsearch.tplvalue.render" (dict "value" .name "context" $context)) -}}
    {{- end -}}
    {{- range .images -}}
      {{- if .pullSecretName -}}
        {{- $pullSecrets = append $pullSecrets (include "elasticsearch.tplvalue.render" (dict "value" .pullSecretName "context" $context)) -}}
      {{- else if (or .pullSecretUsername .pullSecretDockerConfigJson) -}}
        {{- $pullSecrets = append $pullSecrets ((list (include "elasticsearch.uname" $context ) "pull-secret") | join "-")  -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}

  {{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
    {{- range $pullSecrets | uniq }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}
