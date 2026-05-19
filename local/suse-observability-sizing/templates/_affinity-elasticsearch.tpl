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
Returns "hard" or "soft" for HA profiles based on global config, empty for non-HA profiles.

Usage in values.yaml:
elasticsearch:
  antiAffinity: {{ include "common.sizing.elasticsearch.antiAffinityConfig" . | quote }}

Returns: "hard", "soft", or ""
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityConfig" -}}
{{- include "common.sizing.elasticsearch.antiAffinity" . -}}
{{- end -}}

{{/*
Get Elasticsearch antiAffinityTopologyKey for HA profiles.
Returns custom topologyKey from global config or "kubernetes.io/hostname" for HA profiles.

Usage in values.yaml:
elasticsearch:
  antiAffinityTopologyKey: {{ include "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" . | quote }}

Returns: topology key string or ""
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" -}}
{{- include "common.sizing.elasticsearch.antiAffinityTopologyKey" . -}}
{{- end -}}

{{/*
Get Elasticsearch nodeAffinity from global config.
Returns the nodeAffinity configuration if set in global.suseObservability.affinity.nodeAffinity.

Usage in values.yaml:
elasticsearch:
  nodeAffinity: {{ include "common.sizing.elasticsearch.nodeAffinityConfig" . | nindent 4 }}

Returns: nodeAffinity configuration or empty
*/}}
{{- define "common.sizing.elasticsearch.nodeAffinityConfig" -}}
{{- include "common.sizing.global.nodeAffinity" . -}}
{{- end -}}
