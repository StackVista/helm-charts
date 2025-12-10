{{/*
=============================================================================
Kafka Affinity Configuration
These templates provide complete affinity configurations for Kafka.
=============================================================================
*/}}

{{/*
Get complete Kafka affinity configuration for HA profiles.
Returns a complete affinity block with podAntiAffinity for HA profiles, empty for non-HA.

Usage in values.yaml:
kafka:
  affinity: {{ include "common.sizing.kafka.affinityConfig" . | nindent 4 }}

Returns (for HA profiles):
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/component: kafka
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.kafka.affinityConfig" -}}
{{- include "common.sizing.kafka.affinity" . -}}
{{- end -}}
