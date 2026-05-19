{{/*
=============================================================================
OpenTelemetry Collector Affinity Configuration
These templates provide complete affinity configurations for OpenTelemetry Collector.
=============================================================================
*/}}

{{/*
Get complete OpenTelemetry Collector affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles. Empty for non-HA profiles without nodeAffinity.

Usage in values.yaml:
opentelemetry-collector:
  affinity: {{ include "common.sizing.opentelemetry-collector.affinityConfig" . | nindent 4 }}

Returns (for HA profiles with nodeAffinity):
  nodeAffinity:
    <from global.suseObservability.affinity.nodeAffinity>
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/component: statefulset-collector
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.opentelemetry-collector.affinityConfig" -}}
{{- $result := dict -}}
{{- /* Add nodeAffinity from global config */ -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- /* Add podAntiAffinity from sizing profile */ -}}
{{- $podAntiAffinity := include "common.sizing.opentelemetry-collector.affinity" . | trim -}}
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
