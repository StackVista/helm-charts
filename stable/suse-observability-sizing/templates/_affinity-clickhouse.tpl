{{/*
=============================================================================
ClickHouse Affinity Configuration
These templates provide complete affinity configurations for ClickHouse.
=============================================================================
*/}}

{{/*
Get complete ClickHouse affinity configuration for HA profiles.
Returns a complete affinity block with podAntiAffinity for HA profiles, empty for non-HA.

Usage in values.yaml:
clickhouse:
  affinity: {{ include "common.sizing.clickhouse.affinityConfig" . | nindent 4 }}

Returns (for HA profiles):
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/component: clickhouse
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.clickhouse.affinityConfig" -}}
{{- include "common.sizing.clickhouse.affinity" . -}}
{{- end -}}
