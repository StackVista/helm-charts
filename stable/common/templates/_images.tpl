{{/*
Return the proper Docker Image Registry Secret Names evaluating values as templates
{{ include "common.image.pullSecret.name" ( dict "images" (list .Values.path.to.the.image1, .Values.path.to.the.image2) "context" $) }}
*/}}
{{- define "common.image.pullSecret.name" -}}
  {{- $pullSecrets := list }}
  {{- $context := .context }}

  {{- if $context.Values.global }}
    {{- range $context.Values.global.imagePullSecrets -}}
      {{- $pullSecrets = append $pullSecrets (include "common.tplvalue.render" (dict "value" . "context" $context)) -}}
    {{- end -}}
  {{- end -}}
  {{- range $context.Values.imagePullSecrets -}}
    {{- $pullSecrets = append $pullSecrets (include "common.tplvalue.render" (dict "value" .name "context" $context)) -}}
  {{- end -}}
  {{- range .images -}}
    {{- if .pullSecretName -}}
      {{- $pullSecrets = append $pullSecrets (include "common.tplvalue.render" (dict "value" .pullSecretName "context" $context)) -}}
    {{- else if (or .pullSecretUsername .pullSecretDockerConfigJson) -}}
      {{- $pullSecrets = append $pullSecrets ((list (include "common.fullname.short" $context ) "pull-secret") | join "-")  -}}
    {{- end -}}
  {{- end -}}

  {{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
    {{- range $pullSecrets | uniq }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Return the image registry
{{ include "common.image.registry" ( dict "image" . "context" $) }}
When global.suseObservability is enabled (global mode), defaults to registry.rancher.com
to match the suse-observability-values chart behavior.
*/}}
{{- define "common.image.registry" -}}
  {{- if .context.Values.global }}
    {{- if .context.Values.global.imageRegistry -}}
      {{- .context.Values.global.imageRegistry -}}
    {{- else if and .context.Values.global.suseObservability (or .context.Values.global.suseObservability.license .context.Values.global.suseObservability.baseUrl .context.Values.global.suseObservability.sizing.profile) -}}
      {{- /* Global mode is enabled, default to registry.rancher.com */ -}}
      registry.rancher.com
    {{- else -}}
      {{- .image.registry -}}
    {{- end -}}
  {{- else -}}
    {{- .image.registry -}}
  {{- end -}}
{{- end -}}
