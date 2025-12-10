{{/*
AFFINITY TEMPLATES
=============================================================================
*/}}

{{/*
Helper to generate podAntiAffinity configuration
Usage: {{ include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kafka") "topologyKey" "kubernetes.io/hostname" "context" .) }}
*/}}
{{- define "common.sizing.podAntiAffinity" -}}
{{- $labels := .labels -}}
{{- $topologyKey := .topologyKey | default "kubernetes.io/hostname" -}}
{{- $context := .context -}}
{{- if $context.Values -}}
{{- if $context.Values.global -}}
{{- if $context.Values.global.sizing -}}
{{- if $context.Values.global.sizing.profile -}}
{{- $profile := $context.Values.global.sizing.profile -}}
{{- if hasSuffix "-ha" $profile }}
podAntiAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchLabels:
{{- range $key, $value := $labels }}
        {{ $key }}: {{ $value }}
{{- end }}
    topologyKey: {{ $topologyKey }}
{{- end }}
{{- end }}
{{- end }}
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
Usage: {{ include "common.sizing.elasticsearch.antiAffinity" . }}
*/}}
{{- define "common.sizing.elasticsearch.antiAffinity" -}}
{{- if .Values.global -}}
{{- if .Values.global.sizing -}}
{{- if .Values.global.sizing.profile -}}
{{- $profile := .Values.global.sizing.profile -}}
{{- if hasSuffix "-ha" $profile }}hard
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Get elasticsearch antiAffinityTopologyKey
Usage: {{ include "common.sizing.elasticsearch.antiAffinityTopologyKey" . }}
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityTopologyKey" -}}
{{- if .Values.global -}}
{{- if .Values.global.sizing -}}
{{- if .Values.global.sizing.profile -}}
{{- $profile := .Values.global.sizing.profile -}}
{{- if hasSuffix "-ha" $profile }}kubernetes.io/hostname
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Get hbase master affinity
Usage: {{ include "common.sizing.hbase.master.affinity" . }}
*/}}
{{- define "common.sizing.hbase.master.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hbase-master") "context" .) }}
{{- end }}

{{/*
Get hbase regionserver affinity
Usage: {{ include "common.sizing.hbase.regionserver.affinity" . }}
*/}}
{{- define "common.sizing.hbase.regionserver.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hbase-rs") "context" .) }}
{{- end }}

{{/*
Get hbase hdfs datanode affinity
Usage: {{ include "common.sizing.hbase.hdfs.datanode.affinity" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.datanode.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hdfs-dn") "context" .) }}
{{- end }}

{{/*
Get hbase hdfs namenode affinity
Usage: {{ include "common.sizing.hbase.hdfs.namenode.affinity" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.namenode.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hdfs-snn") "context" .) }}
{{- end }}

{{/*
Get hbase hdfs secondarynamenode affinity
Usage: {{ include "common.sizing.hbase.hdfs.secondarynamenode.affinity" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.secondarynamenode.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "hdfs-nn") "context" .) }}
{{- end }}

{{/*
Get hbase tephra affinity
Usage: {{ include "common.sizing.hbase.tephra.affinity" . }}
*/}}
{{- define "common.sizing.hbase.tephra.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "tephra") "context" .) }}
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
