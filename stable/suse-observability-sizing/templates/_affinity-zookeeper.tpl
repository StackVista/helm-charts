{{/*
=============================================================================
Zookeeper Affinity Configuration
These templates provide complete affinity configurations for Zookeeper.
=============================================================================
*/}}

{{/*
Get complete Zookeeper affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles. Empty for non-HA profiles without nodeAffinity.

Usage in values.yaml:
zookeeper:
  affinity: {{ include "common.sizing.zookeeper.affinityConfig" . | nindent 4 }}

Returns (for HA profiles with nodeAffinity):
  nodeAffinity:
    <from global.suseObservability.affinity.nodeAffinity>
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/component: zookeeper
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.zookeeper.affinityConfig" -}}
{{- $result := dict -}}
{{- /* Add nodeAffinity from global config */ -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- /* Add podAntiAffinity from sizing profile */ -}}
{{- $podAntiAffinity := include "common.sizing.zookeeper.affinity" . | trim -}}
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
