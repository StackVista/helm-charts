{{/*
=============================================================================
Victoria Metrics Affinity Configuration
These templates provide complete affinity configurations for Victoria Metrics instances.
=============================================================================
*/}}

{{/*
Get complete Victoria Metrics 0 affinity configuration for HA profiles.
Returns a complete affinity block with podAntiAffinity targeting victoria-metrics-1 for HA profiles.

Usage in values.yaml:
victoria-metrics-0:
  server:
    affinity: {{ include "common.sizing.victoria-metrics-0.affinityConfig" . | nindent 6 }}

Returns (for HA profiles):
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/name: victoria-metrics-1
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.victoria-metrics-0.affinityConfig" -}}
{{- include "common.sizing.victoria-metrics-0.affinity" . -}}
{{- end -}}

{{/*
Get complete Victoria Metrics 1 affinity configuration for HA profiles.
Returns a complete affinity block with podAntiAffinity targeting victoria-metrics-0 for HA profiles.

Usage in values.yaml:
victoria-metrics-1:
  server:
    affinity: {{ include "common.sizing.victoria-metrics-1.affinityConfig" . | nindent 6 }}

Returns (for HA profiles):
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/name: victoria-metrics-0
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.victoria-metrics-1.affinityConfig" -}}
{{- include "common.sizing.victoria-metrics-1.affinity" . -}}
{{- end -}}
