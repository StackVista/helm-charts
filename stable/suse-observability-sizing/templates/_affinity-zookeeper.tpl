{{/*
=============================================================================
Zookeeper Affinity Configuration
These templates provide complete affinity configurations for Zookeeper.
=============================================================================
*/}}

{{/*
Get complete Zookeeper affinity configuration for HA profiles.
Returns a complete affinity block with podAntiAffinity for HA profiles, empty for non-HA.

Usage in values.yaml:
zookeeper:
  affinity: {{ include "common.sizing.zookeeper.affinityConfig" . | nindent 4 }}

Returns (for HA profiles):
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/component: zookeeper
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.zookeeper.affinityConfig" -}}
{{- include "common.sizing.zookeeper.affinity" . -}}
{{- end -}}
