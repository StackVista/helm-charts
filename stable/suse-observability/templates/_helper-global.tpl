{{/*
=============================================================================
SUSE Observability Global Configuration Helpers
These helpers process the global.suseObservability section for simplified configuration.
=============================================================================
*/}}

{{/*
Check if global.suseObservability mode is enabled.

Usage:
{{ if include "suse-observability.global.enabled" . }}
*/}}
{{- define "suse-observability.global.enabled" -}}
{{- include "suse-observability.sizing-profile.enabled" . -}}
{{- end -}}

{{/*
Get the effective license key.
Prefers global.suseObservability.license if enabled, falls back to stackstate.license.key.

Usage:
{{ include "suse-observability.global.license" . }}
*/}}
{{- define "suse-observability.global.license" -}}
{{- if and (include "suse-observability.global.enabled" .) .Values.global.suseObservability.license -}}
{{ .Values.global.suseObservability.license }}
{{- else if .Values.stackstate.license.key -}}
{{ .Values.stackstate.license.key }}
{{- end -}}
{{- end -}}

{{/*
Get the effective base URL.
Prefers global.suseObservability.baseUrl if enabled, falls back to stackstate.baseUrl.

Usage:
{{ include "suse-observability.global.baseUrl" . }}
*/}}
{{- define "suse-observability.global.baseUrl" -}}
{{- if and (include "suse-observability.global.enabled" .) .Values.global.suseObservability.baseUrl -}}
{{ .Values.global.suseObservability.baseUrl }}
{{- else if .Values.stackstate.baseUrl -}}
{{ .Values.stackstate.baseUrl }}
{{- end -}}
{{- end -}}

{{/*
Get the effective HBase deployment mode.
Returns "Mono" for non-HA profiles when global sizing is enabled, otherwise uses .Values.hbase.deployment.mode

Usage:
{{ include "suse-observability.hbase.deploymentMode" . }}
*/}}
{{- define "suse-observability.hbase.deploymentMode" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $mode := include "common.sizing.hbase.deployment.mode" . | trim -}}
{{- if $mode -}}
{{- $mode -}}
{{- else -}}
{{- .Values.hbase.deployment.mode | default "Distributed" -}}
{{- end -}}
{{- else -}}
{{- .Values.hbase.deployment.mode | default "Distributed" -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective admin password.
Prefers global.suseObservability.adminPassword if enabled, falls back to stackstate.authentication.adminPassword.

Usage:
{{ include "suse-observability.global.adminPassword" . }}
*/}}
{{- define "suse-observability.global.adminPassword" -}}
{{- if and (include "suse-observability.global.enabled" .) .Values.global.suseObservability.adminPassword -}}
{{ .Values.global.suseObservability.adminPassword }}
{{- else if .Values.stackstate.authentication.adminPassword -}}
{{ .Values.stackstate.authentication.adminPassword }}
{{- end -}}
{{- end -}}

{{/*
Get the effective image registry.
Prefers global.imageRegistry, falls back to global.suseObservability.imageRegistry if in global mode.
Note: This is currently not exposed in global.suseObservability, users should use global.imageRegistry directly.

Usage:
{{ include "suse-observability.global.imageRegistry" . }}
*/}}
{{- define "suse-observability.global.imageRegistry" -}}
{{- if .Values.global.imageRegistry -}}
{{ .Values.global.imageRegistry }}
{{- end -}}
{{- end -}}

{{/*
Get the effective receiver API key.
Prefers global.suseObservability.receiverApiKey if enabled, falls back to:
  1. stackstate.apiKey.key (used by suse-observability-values chart)
  2. global.receiverApiKey (deprecated)
  3. stackstate.receiver.apiKey

Usage:
{{ include "suse-observability.global.receiverApiKey" . }}
*/}}
{{- define "suse-observability.global.receiverApiKey" -}}
{{- if and (include "suse-observability.global.enabled" .) .Values.global.suseObservability.receiverApiKey -}}
{{ .Values.global.suseObservability.receiverApiKey }}
{{- else if .Values.stackstate.apiKey.key -}}
{{ .Values.stackstate.apiKey.key }}
{{- else if .Values.global.receiverApiKey -}}
{{ .Values.global.receiverApiKey }}
{{- else if .Values.stackstate.receiver.apiKey -}}
{{ .Values.stackstate.receiver.apiKey }}
{{- end -}}
{{- end -}}

{{/*
Check if pull secret should be created from global.suseObservability.

Usage:
{{ if include "suse-observability.global.hasPullSecret" . }}
*/}}
{{- define "suse-observability.global.hasPullSecret" -}}
{{- if and (include "suse-observability.global.enabled" .) .Values.global.suseObservability.pullSecret.username .Values.global.suseObservability.pullSecret.password -}}
true
{{- end -}}
{{- end -}}

{{/*
Get the effective server.split value based on sizing profile.
If explicitly set in stackstate.features.server.split, use that.
Otherwise, determine from profile: non-HA profiles (trial, 10-nonha, etc.) = false, HA profiles (150-ha, etc.) = true

Usage:
{{ include "suse-observability.global.serverSplit" . }}
*/}}
{{- define "suse-observability.global.serverSplit" -}}
{{- $explicitValue := "" -}}
{{- if and .Values.stackstate .Values.stackstate.features .Values.stackstate.features.server -}}
  {{- if hasKey .Values.stackstate.features.server "split" -}}
    {{- $explicitValue = .Values.stackstate.features.server.split -}}
  {{- end -}}
{{- end -}}
{{- if ne $explicitValue "" -}}
{{- $explicitValue -}}
{{- else if include "suse-observability.global.enabled" . -}}
  {{- include "common.sizing.stackstate.server.split" . -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective victoria-metrics-1.enabled value based on sizing profile.
When global sizing is enabled, use sizing profile logic.
Otherwise, use explicitly set victoria-metrics-1.enabled value.

Usage:
{{ include "suse-observability.global.victoriaMetrics1Enabled" . }}
*/}}
{{- define "suse-observability.global.victoriaMetrics1Enabled" -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- include "common.sizing.victoria-metrics-1.enabled" . -}}
{{- else if hasKey .Values "victoria-metrics-1" -}}
  {{- if hasKey (index .Values "victoria-metrics-1") "enabled" -}}
    {{- index .Values "victoria-metrics-1" "enabled" -}}
  {{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective receiver.split.enabled value based on sizing profile.
If explicitly set in stackstate.components.receiver.split.enabled, use that.
Otherwise, determine from profile: non-HA profiles = false, HA profiles = true

Usage:
{{ include "suse-observability.global.receiverSplitEnabled" . }}
*/}}
{{- define "suse-observability.global.receiverSplitEnabled" -}}
{{- $explicitValue := "" -}}
{{- if and .Values.stackstate .Values.stackstate.components .Values.stackstate.components.receiver .Values.stackstate.components.receiver.split -}}
  {{- if hasKey .Values.stackstate.components.receiver.split "enabled" -}}
    {{- $explicitValue = .Values.stackstate.components.receiver.split.enabled -}}
  {{- end -}}
{{- end -}}
{{- if ne $explicitValue "" -}}
{{- $explicitValue -}}
{{- else if include "suse-observability.global.enabled" . -}}
  {{- include "common.sizing.stackstate.receiver.split.enabled" . -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective receiver.retention value based on sizing profile.
When global sizing is enabled, use sizing profile logic.
Otherwise, use explicitly set stackstate.components.receiver.retention value.

Usage:
{{ include "suse-observability.global.receiverRetention" . }}
*/}}
{{- define "suse-observability.global.receiverRetention" -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- $retention := include "common.sizing.stackstate.receiver.retention" . | trim -}}
  {{- if $retention -}}
    {{- $retention -}}
  {{- else -}}
    {{- .Values.stackstate.components.receiver.retention -}}
  {{- end -}}
{{- else -}}
  {{- .Values.stackstate.components.receiver.retention -}}
{{- end -}}
{{- end -}}

{{/*
Validate required fields when global.suseObservability mode is enabled.
Call this early in the chart rendering to fail fast.

Usage:
{{ include "suse-observability.global.validate" . }}
*/}}
{{- define "suse-observability.global.validate" -}}
{{- /* Validate global.suseObservability configuration when auto-detected */ -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- $license := .Values.global.suseObservability.license | required "global.suseObservability.license is required when using global.suseObservability configuration" -}}
  {{- $baseUrl := .Values.global.suseObservability.baseUrl | required "global.suseObservability.baseUrl is required when using global.suseObservability configuration" -}}
  {{- include "stackstate.values.validateUrl" (dict "value" $baseUrl "errorMessage" "Please provide your SUSE Observability base URL in 'global.suseObservability.baseUrl'") -}}

  {{- /* Validate sizing profile using library chart validation */ -}}
  {{- include "common.sizing.profiles.validate" . -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective affinity configuration for a component.
Merges affinity from multiple sources in priority order:
1. Component-specific affinity (highest priority)
2. Global affinity (global.suseObservability.affinity)
3. Profile-based affinity (for HA profiles)

Usage:
{{ include "suse-observability.global.affinity" (dict "componentAffinity" .Values.stackstate.components.api.affinity "allAffinity" .Values.stackstate.components.all.affinity "context" .) }}

Parameters:
- componentAffinity: Component-specific affinity override
- allAffinity: All-components affinity override
- context: The root context (.)
*/}}
{{- define "suse-observability.global.affinity" -}}
{{- $componentAffinity := .componentAffinity -}}
{{- $allAffinity := .allAffinity -}}
{{- $context := .context -}}
{{- $result := dict -}}

{{- /* Step 1: Merge global.suseObservability.affinity (for application components) */ -}}
{{- if include "suse-observability.global.enabled" $context -}}
  {{- if $context.Values.global.suseObservability.affinity -}}
    {{- /* Merge nodeAffinity */ -}}
    {{- if $context.Values.global.suseObservability.affinity.nodeAffinity -}}
      {{- $_ := set $result "nodeAffinity" $context.Values.global.suseObservability.affinity.nodeAffinity -}}
    {{- end -}}

    {{- /* Merge podAffinity */ -}}
    {{- if $context.Values.global.suseObservability.affinity.podAffinity -}}
      {{- $_ := set $result "podAffinity" $context.Values.global.suseObservability.affinity.podAffinity -}}
    {{- end -}}

    {{- /* Merge podAntiAffinity if explicitly provided in proper Kubernetes format */ -}}
    {{- if $context.Values.global.suseObservability.affinity.podAntiAffinity -}}
      {{- $podAntiAffinity := $context.Values.global.suseObservability.affinity.podAntiAffinity -}}
      {{- /* Check if it's a list (proper Kubernetes format), not a boolean (simplified format) */ -}}
      {{- $hasValidRequired := and (hasKey $podAntiAffinity "requiredDuringSchedulingIgnoredDuringExecution") (kindOf $podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution | eq "slice") -}}
      {{- $hasValidPreferred := and (hasKey $podAntiAffinity "preferredDuringSchedulingIgnoredDuringExecution") (kindOf $podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution | eq "slice") -}}
      {{- if or $hasValidRequired $hasValidPreferred -}}
        {{- $_ := set $result "podAntiAffinity" $podAntiAffinity -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- /* Step 3: Merge stackstate.components.all.affinity */ -}}
{{- if $allAffinity -}}
  {{- range $key, $value := $allAffinity -}}
    {{- $_ := set $result $key $value -}}
  {{- end -}}
{{- end -}}

{{- /* Step 4: Merge component-specific affinity (highest priority) */ -}}
{{- if $componentAffinity -}}
  {{- range $key, $value := $componentAffinity -}}
    {{- $_ := set $result $key $value -}}
  {{- end -}}
{{- end -}}

{{- /* Output the merged affinity */ -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective Kafka replicaCount.
Prefers sizing template if global.suseObservability is enabled, falls back to values.yaml.

Usage:
{{ include "suse-observability.global.kafkaReplicaCount" . }}
*/}}
{{- define "suse-observability.global.kafkaReplicaCount" -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- $replicas := include "common.sizing.kafka.replicaCount" . | trim -}}
  {{- if $replicas -}}
    {{- $replicas -}}
  {{- else -}}
    {{- .Values.kafka.replicaCount -}}
  {{- end -}}
{{- else -}}
  {{- .Values.kafka.replicaCount -}}
{{- end -}}
{{- end -}}

{{/*
Get the effective sync tmpToPVC volumeSize.
Prefers sizing template if global.suseObservability is enabled, falls back to values.yaml.

Usage:
{{ include "suse-observability.global.syncTmpToPVCVolumeSize" . }}
*/}}
{{- define "suse-observability.global.syncTmpToPVCVolumeSize" -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- $volumeSize := include "common.sizing.stackstate.sync.tmpToPVC.volumeSize" . | trim -}}
  {{- if $volumeSize -}}
    {{- $volumeSize -}}
  {{- else if and .Values.stackstate.components.sync.tmpToPVC .Values.stackstate.components.sync.tmpToPVC.volumeSize -}}
    {{- .Values.stackstate.components.sync.tmpToPVC.volumeSize -}}
  {{- end -}}
{{- else if and .Values.stackstate.components.sync.tmpToPVC .Values.stackstate.components.sync.tmpToPVC.volumeSize -}}
  {{- .Values.stackstate.components.sync.tmpToPVC.volumeSize -}}
{{- end -}}
{{- end -}}

{{/*
=============================================================================
INFRASTRUCTURE AFFINITY HELPERS
These helpers provide affinity configuration for infrastructure components
(kafka, clickhouse, zookeeper, elasticsearch, hbase, victoria-metrics).
=============================================================================
*/}}

{{/*
Get the global nodeAffinity configuration for infrastructure components.
Returns the nodeAffinity from global.suseObservability.affinity.nodeAffinity if configured.

Usage:
{{ include "suse-observability.global.infra.nodeAffinity" . }}
*/}}
{{- define "suse-observability.global.infra.nodeAffinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- if and .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.nodeAffinity -}}
{{- toYaml .Values.global.suseObservability.affinity.nodeAffinity -}}
  {{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get the podAntiAffinity topologyKey for infrastructure components.
Returns the custom topologyKey from global.suseObservability.affinity.podAntiAffinity.topologyKey
or defaults to "kubernetes.io/hostname".

Usage:
{{ include "suse-observability.global.infra.podAntiAffinity.topologyKey" . }}
*/}}
{{- define "suse-observability.global.infra.podAntiAffinity.topologyKey" -}}
{{- $default := "kubernetes.io/hostname" -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- if and .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.podAntiAffinity .Values.global.suseObservability.affinity.podAntiAffinity.topologyKey -}}
{{- .Values.global.suseObservability.affinity.podAntiAffinity.topologyKey -}}
  {{- else -}}
{{- $default -}}
  {{- end -}}
{{- else -}}
{{- $default -}}
{{- end -}}
{{- end -}}

{{/*
Check if hard (required) pod anti-affinity should be used for infrastructure components.
Returns "true" if requiredDuringSchedulingIgnoredDuringExecution is enabled (default),
"false" if soft anti-affinity should be used.

Usage:
{{ if eq (include "suse-observability.global.infra.podAntiAffinity.required" .) "true" }}
*/}}
{{- define "suse-observability.global.infra.podAntiAffinity.required" -}}
{{- $default := true -}}
{{- if include "suse-observability.global.enabled" . -}}
  {{- if and .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.podAntiAffinity -}}
    {{- if hasKey .Values.global.suseObservability.affinity.podAntiAffinity "requiredDuringSchedulingIgnoredDuringExecution" -}}
      {{- if .Values.global.suseObservability.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution -}}
true
      {{- else -}}
false
      {{- end -}}
    {{- else -}}
{{- $default -}}
    {{- end -}}
  {{- else -}}
{{- $default -}}
  {{- end -}}
{{- else -}}
{{- $default -}}
{{- end -}}
{{- end -}}
