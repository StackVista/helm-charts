{{/*
 checksum annotations
*/}}
{{- define "anomaly-detection.secret.checksum" -}}
checksum/anomaly-detection-cmdline: {{ print (cat (print .Values.threadWorkers)) | sha256sum }}
checksum/anomaly-detection-config-base: {{ include (print $.Template.BasePath "/anomaly-detection-config-base.yaml") . | sha256sum }}
{{- end -}}

{{- define "anomaly-detection.spotlight.configmap-base" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ template "common.fullname.short" . }}-spotlight-config-base
data:
{{- $currentScope := . }}
{{- range $configFile := tuple "spotlight.yaml" "task_logging.conf" "anomaly_detect.yaml" "anomaly_train.yaml" "anomaly_models.yaml" }}
  {{- with $currentScope }}
    {{- $config := .Files.Get (list "etc/" $configFile | join "") }}
    {{- if .Values.etcoverride }}
      {{ $configFromProperty := index .Values "etcoverride" (print $configFile | replace "." "_") }}
      {{- if $configFromProperty }}
        {{- $config = $configFromProperty }}
      {{- end }}
    {{- end }}
    {{- $configFile | nindent 2 }}: |
      {{- $config | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
StackState URL function
*/}}
{{- define "anomaly-detection.stackstate.instance" -}}
{{ tpl .Values.stackstate.instance . }}
{{- end }}

{{/*
Return the image registry
*/}}
{{- define "anomaly-detection.image.registry" -}}
  {{- if .Values.global }}
    {{- .Values.global.imageRegistry | default .Values.image.registry -}}
  {{- else -}}
    {{- .Values.image.registry -}}
  {{- end -}}
{{- end -}}


{{/*
Renders a value that contains a template.
Usage:
{{ include "anomaly-detection.tplvalue.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "anomaly-detection.tplvalue.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names evaluating values as templates
{{ include "anomaly-detection.image.pullSecret.name" ( dict "images" (list .Values.path.to.the.image1, .Values.path.to.the.image2) "context" $) }}
*/}}
{{- define "anomaly-detection.image.pullSecret.name" -}}
  {{- $pullSecrets := list }}
  {{- $context := .context }}

  {{- if and $context.Values.global $context.Values.global.suseObservability $context.Values.global.suseObservability.pullSecret $context.Values.global.suseObservability.pullSecret.username $context.Values.global.suseObservability.pullSecret.password -}}
    {{- $pullSecrets = append $pullSecrets "suse-observability-pull-secret"  -}}
  {{- else -}}
    {{- if $context.Values.global }}
      {{- range $context.Values.global.imagePullSecrets -}}
        {{- $pullSecrets = append $pullSecrets (include "anomaly-detection.tplvalue.render" (dict "value" . "context" $context)) -}}
      {{- end -}}
    {{- end -}}
    {{- range $context.Values.imagePullSecrets -}}
      {{- $pullSecrets = append $pullSecrets (include "anomaly-detection.tplvalue.render" (dict "value" .name "context" $context)) -}}
    {{- end -}}
    {{- range .images -}}
      {{- if .pullSecretName -}}
        {{- $pullSecrets = append $pullSecrets (include "anomaly-detection.tplvalue.render" (dict "value" .pullSecretName "context" $context)) -}}
      {{- else if (or .pullSecretUsername .pullSecretDockerConfigJson) -}}
        {{- $pullSecrets = append $pullSecrets ((list (include "common.fullname.short" $context ) "pull-secret") | join "-")  -}}
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

{{- define "anomaly-detection.manager.name" -}}
  {{ template "common.fullname.short" . }}-spotlight-manager
{{- end -}}

{{- define "anomaly-detection.worker.name" -}}
  {{ template "common.fullname.short" . }}-spotlight-worker
{{- end -}}

{{/*
Return  the proper Storage Class
{{ include "anomaly-detection.storage.class" ( dict "persistence" .Values.path.to.the.persistence "global" .Values.global) }}
*/}}
{{- define "anomaly-detection.storage.class" -}}

{{- $storageClass := .persistence.storageClass -}}
{{- if not $storageClass -}}
  {{- if (.global).storageClass -}}
      {{- $storageClass = .global.storageClass -}}
  {{- end -}}
{{- end -}}

{{- if $storageClass -}}
  {{- if (eq "-" $storageClass) -}}
      {{- printf "storageClassName: \"\"" -}}
  {{- else }}
      {{- printf "storageClassName: %s" $storageClass -}}
  {{- end -}}
{{- end -}}

{{- end -}}
