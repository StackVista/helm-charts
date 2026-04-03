{{/*
This file was copied from https://gitlab.com/stackvista/devops/helm-charts/-/blob/master/stable/suse-observability-sizing/templates/_sizing-affinity.tpl?ref_type=heads
*/}}


{{/*
AFFINITY TEMPLATES
=============================================================================
*/}}

{{/*
Get the sizing profile from either global.suseObservability.sizing.profile or global.sizing.profile.
This helper enables backwards compatibility with both configuration paths.

Usage: {{ include "common.sizing.global.profile" . }}
*/}}
{{- define "common.sizing.global.profile" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- .Values.global.suseObservability.sizing.profile -}}
{{- else if and .Values.global .Values.global.sizing .Values.global.sizing.profile -}}
{{- .Values.global.sizing.profile -}}
{{- end -}}
{{- end -}}

{{/*
Get the podAntiAffinity topologyKey from global.suseObservability config.
Returns custom topologyKey or defaults to "kubernetes.io/hostname".

Usage: {{ include "common.sizing.global.topologyKey" . }}
*/}}
{{- define "common.sizing.global.topologyKey" -}}
{{- $default := "kubernetes.io/hostname" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.podAntiAffinity .Values.global.suseObservability.affinity.podAntiAffinity.topologyKey -}}
{{- .Values.global.suseObservability.affinity.podAntiAffinity.topologyKey -}}
{{- else -}}
{{- $default -}}
{{- end -}}
{{- end -}}

{{/*
Check if hard (required) pod anti-affinity should be used.
Returns "true" for hard, "false" for soft. Default is "true".

Usage: {{ include "common.sizing.global.podAntiAffinityRequired" . }}
*/}}
{{- define "common.sizing.global.podAntiAffinityRequired" -}}
{{- $default := "true" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.podAntiAffinity -}}
  {{- if hasKey .Values.global.suseObservability.affinity.podAntiAffinity "requiredDuringSchedulingIgnoredDuringExecution" -}}
    {{- if .Values.global.suseObservability.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution -}}
true
    {{- else -}}
false
    {{- end -}}
  {{- else -}}
{{- $default -}}
  {{- end -}}
{{- else -}}
{{- $default -}}
{{- end -}}
{{- end -}}

{{/*
Get the global nodeAffinity from global.suseObservability config.
Returns empty if not configured.

Usage: {{ include "common.sizing.global.nodeAffinity" . }}
*/}}
{{- define "common.sizing.global.nodeAffinity" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.nodeAffinity -}}
{{- toYaml .Values.global.suseObservability.affinity.nodeAffinity -}}
{{- end -}}
{{- end -}}

{{/*
Helper to generate podAntiAffinity configuration
Supports both hard (required) and soft (preferred) anti-affinity based on global config.
Also reads custom topologyKey from global.suseObservability.affinity.podAntiAffinity.topologyKey.
Checks for profile in both global.suseObservability.sizing.profile and global.sizing.profile.

Usage: {{ include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kafka") "topologyKey" "kubernetes.io/hostname" "context" .) }}
*/}}
{{- define "common.sizing.podAntiAffinity" -}}
{{- $labels := .labels -}}
{{- $context := .context -}}
{{- /* Get topologyKey: explicit param > global config > default */ -}}
{{- $globalTopologyKey := include "common.sizing.global.topologyKey" $context | trim -}}
{{- $topologyKey := .topologyKey | default $globalTopologyKey | default "kubernetes.io/hostname" -}}
{{- /* Check if hard or soft anti-affinity */ -}}
{{- $isRequired := eq (include "common.sizing.global.podAntiAffinityRequired" $context | trim) "true" -}}
{{- /* Get profile from either path */ -}}
{{- $profile := include "common.sizing.global.profile" $context | trim -}}
{{- if and $profile (hasSuffix "-ha" $profile) }}
podAntiAffinity:
{{- if $isRequired }}
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchLabels:
{{- range $key, $value := $labels }}
        {{ $key }}: {{ $value }}
{{- end }}
    topologyKey: {{ $topologyKey }}
{{- else }}
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchLabels:
{{- range $key, $value := $labels }}
          {{ $key }}: {{ $value }}
{{- end }}
      topologyKey: {{ $topologyKey }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Get kubernetes-rbac affinity
Usage: {{ include "common.sizing.anomaly-detection.affinity" . }}
*/}}
{{- define "common.sizing.anomaly-detection.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "worker") "context" .) }}
{{- end }}
