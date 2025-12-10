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
