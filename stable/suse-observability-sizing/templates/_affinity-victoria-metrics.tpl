{{/*
=============================================================================
Victoria Metrics Affinity Configuration
These templates provide complete affinity configurations for Victoria Metrics
instances.
=============================================================================
*/}}

{{/*
Get complete Victoria Metrics 0 affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity targeting victoria-metrics-1 for HA profiles.

Usage in values.yaml:
victoria-metrics-0:
  server:
    affinity: {{ include "common.sizing.victoria-metrics-0.affinityConfig" . | nindent 6 }}

Returns (for HA profiles with nodeAffinity):
  nodeAffinity:
    <from global.suseObservability.affinity.nodeAffinity>
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/name: victoria-metrics-1
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.victoria-metrics-0.affinityConfig" -}}
{{- $result := dict -}}
{{- /* Add nodeAffinity from global config */ -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- /* Add podAntiAffinity from sizing profile */ -}}
{{- $podAntiAffinity := include "common.sizing.victoria-metrics-0.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- /* Output result */ -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete Victoria Metrics 1 affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity targeting victoria-metrics-0 for HA profiles.

Usage in values.yaml:
victoria-metrics-1:
  server:
    affinity: {{ include "common.sizing.victoria-metrics-1.affinityConfig" . | nindent 6 }}

Returns (for HA profiles with nodeAffinity):
  nodeAffinity:
    <from global.suseObservability.affinity.nodeAffinity>
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/name: victoria-metrics-0
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.victoria-metrics-1.affinityConfig" -}}
{{- $result := dict -}}
{{- /* Add nodeAffinity from global config */ -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- /* Add podAntiAffinity from sizing profile */ -}}
{{- $podAntiAffinity := include "common.sizing.victoria-metrics-1.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- /* Output result */ -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}
