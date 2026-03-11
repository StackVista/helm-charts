{{/*
=============================================================================
Sizing-based Infrastructure Affinity Overrides
These helpers merge sizing-chart-based affinity configurations into subchart values.
=============================================================================
*/}}

{{/*
Get Kafka affinity from sizing profile if applicable.
Returns affinity configuration or empty if not in global mode or non-HA profile.

Usage in kafka subchart context:
{{ include "suse-observability.sizing.kafka.affinity" . }}
*/}}
{{- define "suse-observability.sizing.kafka.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.kafka.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get ClickHouse affinity from sizing profile if applicable.

Usage in clickhouse subchart context:
{{ include "suse-observability.sizing.clickhouse.affinity" . }}
*/}}
{{- define "suse-observability.sizing.clickhouse.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.clickhouse.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get Zookeeper affinity from sizing profile if applicable.

Usage in zookeeper subchart context:
{{ include "suse-observability.sizing.zookeeper.affinity" . }}
*/}}
{{- define "suse-observability.sizing.zookeeper.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.zookeeper.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch antiAffinity strategy from sizing profile if applicable.
Returns "hard" for HA profiles, empty otherwise.

Usage in elasticsearch subchart context:
{{ include "suse-observability.sizing.elasticsearch.antiAffinity" . }}
*/}}
{{- define "suse-observability.sizing.elasticsearch.antiAffinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.elasticsearch.antiAffinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch antiAffinityTopologyKey from sizing profile if applicable.
Returns "kubernetes.io/hostname" for HA profiles, empty otherwise.

Usage in elasticsearch subchart context:
{{ include "suse-observability.sizing.elasticsearch.antiAffinityTopologyKey" . }}
*/}}
{{- define "suse-observability.sizing.elasticsearch.antiAffinityTopologyKey" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get Victoria Metrics 0 affinity from sizing profile if applicable.

Usage in victoria-metrics-0 subchart context:
{{ include "suse-observability.sizing.victoria-metrics-0.affinity" . }}
*/}}
{{- define "suse-observability.sizing.victoria-metrics-0.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.victoria-metrics-0.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get Victoria Metrics 1 affinity from sizing profile if applicable.

Usage in victoria-metrics-1 subchart context:
{{ include "suse-observability.sizing.victoria-metrics-1.affinity" . }}
*/}}
{{- define "suse-observability.sizing.victoria-metrics-1.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.victoria-metrics-1.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase master affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.master.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.master.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.master.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase regionserver affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.regionserver.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.regionserver.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.regionserver.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase tephra affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.tephra.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.tephra.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.tephra.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase HDFS datanode affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.hdfs.datanode.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.hdfs.datanode.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.hdfs.datanode.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase HDFS namenode affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.hdfs.namenode.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.hdfs.namenode.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.hdfs.namenode.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase HDFS secondarynamenode affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.hdfs.secondarynamenode.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.hdfs.secondarynamenode.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.hdfs.secondarynamenode.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase console affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.console.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.console.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.console.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase stackgraph affinity from sizing profile if applicable.

Usage in hbase subchart context:
{{ include "suse-observability.sizing.hbase.stackgraph.affinity" . }}
*/}}
{{- define "suse-observability.sizing.hbase.stackgraph.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- include "common.sizing.hbase.stackgraph.affinityConfig" . -}}
{{- end -}}
{{- end -}}

{{/*
Get OpenTelemetry Collector affinity from sizing profile if applicable.

Usage in opentelemetry-collector subchart context:
{{ include "suse-observability.sizing.opentelemetry-collector.affinity" . }}
*/}}
{{- define "suse-observability.sizing.opentelemetry-collector.affinity" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $affinity := include "common.sizing.opentelemetry-collector.affinityConfig" . | trim -}}
{{- if $affinity -}}
{{- $affinity -}}
{{- else -}}
{{- "{}" -}}
{{- end -}}
{{- else -}}
{{- "{}" -}}
{{- end -}}
{{- end -}}

{{/*
=============================================================================
Sizing-based Resource and Storage Overrides for Subcharts
These helpers apply sizing-chart-based resources and storage to subcharts.
=============================================================================
*/}}

{{/*
Get ClickHouse replicaCount from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.clickhouse.replicaCount" -}}
{{- $sizing := include "common.sizing.clickhouse.replicaCount" . | trim -}}
{{- default $sizing .Values.clickhouse.replicaCount -}}
{{- end -}}

{{/*
Get Elasticsearch storage size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.elasticsearch.volumeClaimTemplate.resources.requests.storage" -}}
{{- $sizing := include "common.sizing.elasticsearch.storage" . | trim -}}
{{- default $sizing .Values.elasticsearch.volumeClaimTemplate.resources.requests.storage -}}
{{- end -}}

{{/*
Get Elasticsearch replicas from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.elasticsearch.replicas" -}}
{{- $sizing := include "common.sizing.elasticsearch.replicas" . | trim -}}
{{- default $sizing .Values.elasticsearch.replicas -}}
{{- end -}}

{{/*
Get Elasticsearch esJavaOpts from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.elasticsearch.esJavaOpts" -}}
{{- $sizing := include "common.sizing.elasticsearch.esJavaOpts" . | trim -}}
{{- default $sizing .Values.elasticsearch.esJavaOpts -}}
{{- end -}}

{{/*
Get Kafka persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.kafka.persistence.size" -}}
{{- $sizing := include "common.sizing.kafka.persistence.size" . | trim -}}
{{- default $sizing .Values.kafka.persistence.size -}}
{{- end -}}

{{/*
Get Zookeeper persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.zookeeper.persistence.size" -}}
{{- $sizing := include "common.sizing.zookeeper.persistence.size" . | trim -}}
{{- default $sizing .Values.zookeeper.persistence.size -}}
{{- end -}}

{{/*
Get HBase stackgraph persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.hbase.stackgraph.persistence.size" -}}
{{- $sizing := include "common.sizing.hbase.stackgraph.persistence.size" . | trim -}}
{{- default $sizing .Values.hbase.stackgraph.persistence.size -}}
{{- end -}}

{{/*
Get HBase HDFS datanode persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.hbase.hdfs.datanode.persistence.size" -}}
{{- $sizing := include "common.sizing.hbase.hdfs.datanode.persistence.size" . | trim -}}
{{- default $sizing .Values.hbase.hdfs.datanode.persistence.size -}}
{{- end -}}

{{/*
Get Victoria Metrics 0 storage size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-0.server.persistentVolume.size" -}}
{{- $sizing := include "common.sizing.victoria-metrics.storage" . | trim -}}
{{- default $sizing (index .Values "victoria-metrics-0" "server" "persistentVolume" "size") -}}
{{- end -}}

{{/*
Get Victoria Metrics 1 storage size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-1.server.persistentVolume.size" -}}
{{- $sizing := include "common.sizing.victoria-metrics.storage" . | trim -}}
{{- default $sizing (index .Values "victoria-metrics-1" "server" "persistentVolume" "size") -}}
{{- end -}}

{{/*
Get Victoria Metrics 0 retentionPeriod from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-0.server.retentionPeriod" -}}
{{- $sizing := include "common.sizing.victoria-metrics.retention" . | trim -}}
{{- default $sizing (index .Values "victoria-metrics-0" "server" "retentionPeriod") -}}
{{- end -}}

{{/*
Get Victoria Metrics 1 retentionPeriod from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-1.server.retentionPeriod" -}}
{{- $sizing := include "common.sizing.victoria-metrics.retention" . | trim -}}
{{- default $sizing (index .Values "victoria-metrics-1" "server" "retentionPeriod") -}}
{{- end -}}
