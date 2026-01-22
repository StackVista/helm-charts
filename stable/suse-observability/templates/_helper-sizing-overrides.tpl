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
{{- if include "suse-observability.global.enabled" . -}}
{{- $replicas := include "common.sizing.clickhouse.replicaCount" . | trim -}}
{{- if $replicas -}}
{{- $replicas -}}
{{- else -}}
{{- .Values.clickhouse.replicaCount -}}
{{- end -}}
{{- else -}}
{{- .Values.clickhouse.replicaCount -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch storage size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.elasticsearch.volumeClaimTemplate.resources.requests.storage" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.elasticsearch.storage" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- .Values.elasticsearch.volumeClaimTemplate.resources.requests.storage -}}
{{- end -}}
{{- else -}}
{{- .Values.elasticsearch.volumeClaimTemplate.resources.requests.storage -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch replicas from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.elasticsearch.replicas" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $replicas := include "common.sizing.elasticsearch.replicas" . | trim -}}
{{- if $replicas -}}
{{- $replicas -}}
{{- else -}}
{{- .Values.elasticsearch.replicas -}}
{{- end -}}
{{- else -}}
{{- .Values.elasticsearch.replicas -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch esJavaOpts from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.elasticsearch.esJavaOpts" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $esJavaOpts := include "common.sizing.elasticsearch.esJavaOpts" . | trim -}}
{{- if $esJavaOpts -}}
{{- $esJavaOpts -}}
{{- else -}}
-Xmx3g -Xms3g -Des.allow_insecure_settings=true
{{- end -}}
{{- else -}}
-Xmx3g -Xms3g -Des.allow_insecure_settings=true
{{- end -}}
{{- end -}}

{{/*
Get Kafka persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.kafka.persistence.size" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.kafka.persistence.size" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- .Values.kafka.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.kafka.persistence.size -}}
{{- end -}}
{{- end -}}

{{/*
Get Zookeeper persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.zookeeper.persistence.size" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.zookeeper.persistence.size" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- .Values.zookeeper.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.zookeeper.persistence.size -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase stackgraph persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.hbase.stackgraph.persistence.size" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.hbase.stackgraph.persistence.size" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- .Values.hbase.stackgraph.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.hbase.stackgraph.persistence.size -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase HDFS datanode persistence size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.hbase.hdfs.datanode.persistence.size" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.hbase.hdfs.datanode.persistence.size" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- .Values.hbase.hdfs.datanode.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.hbase.hdfs.datanode.persistence.size -}}
{{- end -}}
{{- end -}}

{{/*
Get Victoria Metrics 0 storage size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-0.server.persistentVolume.size" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.victoria-metrics.storage" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- index .Values "victoria-metrics-0" "server" "persistentVolume" "size" -}}
{{- end -}}
{{- else -}}
{{- index .Values "victoria-metrics-0" "server" "persistentVolume" "size" -}}
{{- end -}}
{{- end -}}

{{/*
Get Victoria Metrics 1 storage size from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-1.server.persistentVolume.size" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $storageSize := include "common.sizing.victoria-metrics.storage" . | trim -}}
{{- if $storageSize -}}
{{- $storageSize -}}
{{- else -}}
{{- index .Values "victoria-metrics-1" "server" "persistentVolume" "size" -}}
{{- end -}}
{{- else -}}
{{- index .Values "victoria-metrics-1" "server" "persistentVolume" "size" -}}
{{- end -}}
{{- end -}}

{{/*
Get Victoria Metrics 0 retentionPeriod from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-0.server.retentionPeriod" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $retention := include "common.sizing.victoria-metrics.retention" . | trim -}}
{{- if $retention -}}
{{- $retention -}}
{{- else -}}
{{- index .Values "victoria-metrics-0" "server" "retentionPeriod" -}}
{{- end -}}
{{- else -}}
{{- index .Values "victoria-metrics-0" "server" "retentionPeriod" -}}
{{- end -}}
{{- end -}}

{{/*
Get Victoria Metrics 1 retentionPeriod from sizing profile if applicable.
*/}}
{{- define "suse-observability.sizing.victoria-metrics-1.server.retentionPeriod" -}}
{{- if include "suse-observability.global.enabled" . -}}
{{- $retention := include "common.sizing.victoria-metrics.retention" . | trim -}}
{{- if $retention -}}
{{- $retention -}}
{{- else -}}
{{- index .Values "victoria-metrics-1" "server" "retentionPeriod" -}}
{{- end -}}
{{- else -}}
{{- index .Values "victoria-metrics-1" "server" "retentionPeriod" -}}
{{- end -}}
{{- end -}}
