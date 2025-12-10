{{/*
=============================================================================
Elasticsearch Affinity Configuration
These templates provide complete affinity configurations for Elasticsearch.
Note: Elasticsearch chart uses antiAffinity and antiAffinityTopologyKey fields
instead of the standard affinity field.
=============================================================================
*/}}

{{/*
Get Elasticsearch antiAffinity strategy for HA profiles.
Returns "hard" for HA profiles, empty for non-HA profiles.

Usage in values.yaml:
elasticsearch:
  antiAffinity: {{ include "common.sizing.elasticsearch.antiAffinityConfig" . | quote }}

Returns: "hard" or ""
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityConfig" -}}
{{- include "common.sizing.elasticsearch.antiAffinity" . -}}
{{- end -}}

{{/*
Get Elasticsearch antiAffinityTopologyKey for HA profiles.
Returns "kubernetes.io/hostname" for HA profiles, empty for non-HA profiles.

Usage in values.yaml:
elasticsearch:
  antiAffinityTopologyKey: {{ include "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" . | quote }}

Returns: "kubernetes.io/hostname" or ""
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" -}}
{{- include "common.sizing.elasticsearch.antiAffinityTopologyKey" . -}}
{{- end -}}
