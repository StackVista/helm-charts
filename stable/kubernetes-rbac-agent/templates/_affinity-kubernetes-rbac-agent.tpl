{{/*
=============================================================================
Kubernetes RBAC Affinity Configuration
These templates provide complete affinity configurations for Kubernetes RBAC.
=============================================================================
*/}}

{{/*
Get complete Kubernetes RBAC Agent affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles. Empty for non-HA profiles without nodeAffinity.

Usage in values.yaml:
kubernetes-rbac-agent:
  affinity: {{ include "common.sizing.kubernetes-rbac-agent.affinityConfig" . | nindent 4 }}

Returns (for HA profiles with nodeAffinity):
  nodeAffinity:
    <from global.suseObservability.affinity.nodeAffinity>
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/component: kubernetes-rbac-agent
      topologyKey: kubernetes.io/hostname
*/}}
{{- define "common.sizing.kubernetes-rbac-agent.affinityConfig" -}}
{{- $result := dict -}}
{{- /* Add nodeAffinity from global config */ -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- /* Add podAntiAffinity from sizing profile */ -}}
{{- $podAntiAffinity := include "common.sizing.kubernetes-rbac-agent.affinity" . | trim -}}
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
