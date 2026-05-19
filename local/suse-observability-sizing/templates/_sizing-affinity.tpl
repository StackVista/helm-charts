{{/*
AFFINITY TEMPLATES
=============================================================================
*/}}

{{/*
Get the sizing profile from either global.suseObservability.sizing.profile or global.sizing.profile.
This helper enables backwards compatibility with both configuration paths.

Usage: {{ include "common.sizing.global.profile" . }}
*/}}
{{- define "common.sizing.global.profile" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- .Values.global.suseObservability.sizing.profile -}}
{{- else if and .Values.global .Values.global.sizing .Values.global.sizing.profile -}}
{{- .Values.global.sizing.profile -}}
{{- end -}}
{{- end -}}

{{/*
Get the podAntiAffinity topologyKey from global.suseObservability config.
Returns custom topologyKey or defaults to "kubernetes.io/hostname".

Usage: {{ include "common.sizing.global.topologyKey" . }}
*/}}
{{- define "common.sizing.global.topologyKey" -}}
{{- $default := "kubernetes.io/hostname" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.podAntiAffinity .Values.global.suseObservability.affinity.podAntiAffinity.topologyKey -}}
{{- .Values.global.suseObservability.affinity.podAntiAffinity.topologyKey -}}
{{- else -}}
{{- $default -}}
{{- end -}}
{{- end -}}

{{/*
Check if hard (required) pod anti-affinity should be used.
Returns "true" for hard, "false" for soft. Default is "true".

Usage: {{ include "common.sizing.global.podAntiAffinityRequired" . }}
*/}}
{{- define "common.sizing.global.podAntiAffinityRequired" -}}
{{- $default := "true" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.podAntiAffinity -}}
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
{{- end -}}

{{/*
Get the global nodeAffinity from global.suseObservability config.
Returns empty if not configured.

Usage: {{ include "common.sizing.global.nodeAffinity" . }}
*/}}
{{- define "common.sizing.global.nodeAffinity" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.affinity .Values.global.suseObservability.affinity.nodeAffinity -}}
{{- toYaml .Values.global.suseObservability.affinity.nodeAffinity -}}
{{- end -}}
{{- end -}}

{{/*
Helper to generate podAntiAffinity configuration
Supports both hard (required) and soft (preferred) anti-affinity based on global config.
Also reads custom topologyKey from global.suseObservability.affinity.podAntiAffinity.topologyKey.
Checks for profile in both global.suseObservability.sizing.profile and global.sizing.profile.

Usage: {{ include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kafka") "topologyKey" "kubernetes.io/hostname" "context" .) }}
*/}}
{{- define "common.sizing.podAntiAffinity" -}}
{{- $labels := .labels -}}
{{- $context := .context -}}
{{- /* Get topologyKey: explicit param > global config > default */ -}}
{{- $globalTopologyKey := include "common.sizing.global.topologyKey" $context | trim -}}
{{- $topologyKey := .topologyKey | default $globalTopologyKey | default "kubernetes.io/hostname" -}}
{{- /* Check if hard or soft anti-affinity */ -}}
{{- $isRequired := eq (include "common.sizing.global.podAntiAffinityRequired" $context | trim) "true" -}}
{{- /* Get profile from either path */ -}}
{{- $profile := include "common.sizing.global.profile" $context | trim -}}
{{- if and $profile (hasSuffix "-ha" $profile) }}
podAntiAffinity:
{{- if $isRequired }}
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchLabels:
{{- range $key, $value := $labels }}
        {{ $key }}: {{ $value }}
{{- end }}
    topologyKey: {{ $topologyKey }}
{{- else }}
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchLabels:
{{- range $key, $value := $labels }}
          {{ $key }}: {{ $value }}
{{- end }}
      topologyKey: {{ $topologyKey }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Get kafka affinity
Usage: {{ include "common.sizing.kafka.affinity" . }}
*/}}
{{- define "common.sizing.kafka.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kafka") "context" .) }}
{{- end }}

{{/*
Get clickhouse affinity
Usage: {{ include "common.sizing.clickhouse.affinity" . }}
*/}}
{{- define "common.sizing.clickhouse.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "clickhouse") "context" .) }}
{{- end }}

{{/*
Get zookeeper affinity
Usage: {{ include "common.sizing.zookeeper.affinity" . }}
*/}}
{{- define "common.sizing.zookeeper.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "zookeeper") "context" .) }}
{{- end }}

{{/*
Get elasticsearch antiAffinity type
Returns "hard" or "soft" based on global config. Default is "hard" for HA profiles.
Checks for profile in both global.suseObservability.sizing.profile and global.sizing.profile.
Usage: {{ include "common.sizing.elasticsearch.antiAffinity" . }}
*/}}
{{- define "common.sizing.elasticsearch.antiAffinity" -}}
{{- $profile := include "common.sizing.global.profile" . | trim -}}
{{- if and $profile (hasSuffix "-ha" $profile) -}}
{{- $isRequired := eq (include "common.sizing.global.podAntiAffinityRequired" . | trim) "true" -}}
{{- if $isRequired }}hard{{- else }}soft{{- end -}}
{{- end }}
{{- end -}}

{{/*
Get elasticsearch antiAffinityTopologyKey
Returns custom topologyKey from global config or default "kubernetes.io/hostname".
Checks for profile in both global.suseObservability.sizing.profile and global.sizing.profile.
Usage: {{ include "common.sizing.elasticsearch.antiAffinityTopologyKey" . }}
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityTopologyKey" -}}
{{- $profile := include "common.sizing.global.profile" . | trim -}}
{{- if and $profile (hasSuffix "-ha" $profile) -}}
{{- include "common.sizing.global.topologyKey" . -}}
{{- end }}
{{- end -}}

{{/*
Helper to generate HBase-specific podAntiAffinity configuration.
For HBase components, outputs BOTH:
- Soft anti-affinity to spread all HBase pods across nodes
- Hard anti-affinity to prevent same-component pods on the same node

Usage: {{ include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hbase-master") "context" .) }}
*/}}
{{- define "common.sizing.hbasePodAntiAffinity" -}}
{{- $labels := .labels -}}
{{- $context := .context -}}
{{- $globalTopologyKey := include "common.sizing.global.topologyKey" $context | trim -}}
{{- $topologyKey := .topologyKey | default $globalTopologyKey | default "kubernetes.io/hostname" -}}
{{- $profile := include "common.sizing.global.profile" $context | trim -}}
{{- if and $profile (hasSuffix "-ha" $profile) }}
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
  - podAffinityTerm:
      labelSelector:
        matchLabels:
          app.kubernetes.io/instance: {{ $context.Release.Name }}
          app.kubernetes.io/name: hbase
      topologyKey: {{ $topologyKey }}
    weight: 1
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchLabels:
{{- range $key, $value := $labels }}
        {{ $key }}: {{ $value }}
{{- end }}
    topologyKey: {{ $topologyKey }}
{{- end }}
{{- end -}}

{{/*
Get hbase master affinity
Usage: {{ include "common.sizing.hbase.master.affinity" . }}
*/}}
{{- define "common.sizing.hbase.master.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hbase-master") "context" .) }}
{{- end }}

{{/*
Get hbase regionserver affinity
Usage: {{ include "common.sizing.hbase.regionserver.affinity" . }}
*/}}
{{- define "common.sizing.hbase.regionserver.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hbase-rs") "context" .) }}
{{- end }}

{{/*
Get hbase hdfs datanode affinity
Usage: {{ include "common.sizing.hbase.hdfs.datanode.affinity" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.datanode.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hdfs-dn") "context" .) }}
{{- end }}

{{/*
Get hbase hdfs namenode affinity
Usage: {{ include "common.sizing.hbase.hdfs.namenode.affinity" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.namenode.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hdfs-snn") "context" .) }}
{{- end }}

{{/*
Get hbase hdfs secondarynamenode affinity
Usage: {{ include "common.sizing.hbase.hdfs.secondarynamenode.affinity" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.secondarynamenode.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hdfs-nn") "context" .) }}
{{- end }}

{{/*
Get hbase tephra affinity
Usage: {{ include "common.sizing.hbase.tephra.affinity" . }}
*/}}
{{- define "common.sizing.hbase.tephra.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "tephra") "context" .) }}
{{- end }}

{{/*
Get hbase console affinity
Usage: {{ include "common.sizing.hbase.console.affinity" . }}
*/}}
{{- define "common.sizing.hbase.console.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "console") "context" .) }}
{{- end }}

{{/*
Get hbase stackgraph affinity
Usage: {{ include "common.sizing.hbase.stackgraph.affinity" . }}
*/}}
{{- define "common.sizing.hbase.stackgraph.affinity" -}}
{{- include "common.sizing.hbasePodAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "stackgraph") "context" .) }}
{{- end }}

{{/*
Get victoria-metrics-0 affinity
Usage: {{ include "common.sizing.victoria-metrics-0.affinity" . }}
*/}}
{{- define "common.sizing.victoria-metrics-0.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/name" "victoria-metrics-1") "context" .) }}
{{- end }}

{{/*
Get victoria-metrics-1 affinity
Usage: {{ include "common.sizing.victoria-metrics-1.affinity" . }}
*/}}
{{- define "common.sizing.victoria-metrics-1.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/name" "victoria-metrics-0") "context" .) }}
{{- end }}

{{/*
Get opentelemetry-collector affinity
Usage: {{ include "common.sizing.opentelemetry-collector.affinity" . }}
*/}}
{{- define "common.sizing.opentelemetry-collector.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "statefulset-collector") "context" .) }}
{{- end }}

{{/*
Get minio affinity
Usage: {{ include "common.sizing.minio.affinity" . }}
*/}}
{{- define "common.sizing.minio.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "minio") "context" .) }}
{{- end }}

{{/*
Get kafkaup-operator affinity
Usage: {{ include "common.sizing.kafkaup-operator.affinity" . }}
*/}}
{{- define "common.sizing.kafkaup-operator.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kafkaup-operator") "context" .) }}
{{- end }}

{{/*
Get prometheus-elasticsearch-exporter affinity
Usage: {{ include "common.sizing.prometheus-elasticsearch-exporter.affinity" . }}
*/}}
{{- define "common.sizing.prometheus-elasticsearch-exporter.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "prometheus-elasticsearch-exporter") "context" .) }}
{{- end }}

{{/*
Get anomaly-detection affinity
Usage: {{ include "common.sizing.anomaly-detection.affinity" . }}
*/}}
{{- define "common.sizing.anomaly-detection.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "anomaly-detection") "context" .) }}
{{- end }}

{{/*
Get kubernetes-rbac affinity
Usage: {{ include "common.sizing.kubernetes-rbac.affinity" . }}
*/}}
{{- define "common.sizing.kubernetes-rbac.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kubernetes-rbac") "context" .) }}
{{- end }}
