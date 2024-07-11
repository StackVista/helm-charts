{{/*
Logic to determine remote write endpoint.
*/}}
{{- define "stackstate.metrics.remotewrite.url" -}}
http://{{- include "stackstate.metrics.remoteWriteEndpoint" . -}}{{.Values.stackstate.components.all.metricStore.remoteWritePath}}
{{- end -}}

{{/*
Logic to determine promql query endpoint.
*/}}
{{- define "stackstate.metrics.query.url" -}}
http://{{- include "stackstate.metrics.queryApiEndpoint" . -}}{{.Values.stackstate.components.all.metricStore.queryApiPath}}
{{- end -}}

{{/*
Logic to determine metric store host and port
*/}}
{{- define "stackstate.metrics.remoteWriteEndpoint" -}}
{{- .Values.stackstate.components.all.metricStore.remoteWriteEndpoint | required "stackstate.components.all.metricStore.remoteWriteEndpoint is a required value." -}}
{{- end -}}

{{/*
Logic to determine metric store host and port
*/}}
{{- define "stackstate.metrics.queryApiEndpoint" -}}
{{- .Values.stackstate.components.all.metricStore.queryApiEndpoint | required "stackstate.components.all.metricStore.queryApiEndpoint is a required value." -}}
{{- end -}}

{{/*
Logic to determine metric store consumer group
*/}}
{{- define "stackstate.metrics.kafka2PromGroupId" -}}
{{- .Values.stackstate.components.all.metricStore.kafka2PromGroupId | required "stackstate.components.all.metricStore.kafka2PromGroupId is a required value." -}}
{{- end -}}

{{/*
Logic to determine ElasticSearch endpoint.
*/}}
{{- define "stackstate.es.endpoint" -}}
{{- if .Values.elasticsearch.enabled -}}
{{- .Values.elasticsearch.clusterName -}}-{{ .Values.elasticsearch.nodeGroup }}-headless:9200
{{- else -}}
{{- .Values.stackstate.components.all.elasticsearchEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "stackstate.kafka.endpoint" -}}
{{- if .Values.kafka.enabled -}}
{{- .Values.kafka.fullnameOverride -}}-headless:9092
{{- else -}}
{{- .Values.stackstate.components.all.kafkaEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine MinIO endpoint.
*/}}
{{- define "stackstate.minio.endpoint" -}}
{{- .Values.minio.fullnameOverride -}}:9000
{{- end -}}

{{/*
Logic to determine Router endpoint.
*/}}
{{- define "stackstate.router.endpoint" -}}
http://{{ template "common.fullname.short" . }}-router:8080
{{- end -}}

{{/*
Logic to determine MinIO keys.
*/}}
{{- define "stackstate.minio.keys" -}}
{{- .Values.minio.fullnameOverride -}}
{{- end -}}

{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "stackstate.zookeeper.endpoint" -}}
{{- if .Values.zookeeper.enabled -}}
{{- .Values.zookeeper.fullnameOverride -}}-headless:2181
{{- else -}}
{{- .Values.stackstate.components.all.zookeeperEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Clickhouse endpoint.
*/}}
{{- define "stackstate.clickhouse.endpoint" -}}
{{- .Values.clickhouse.fullnameOverride }}-headless:8123
{{- end -}}

{{/*
Logic to determine otel collector endpoint.
*/}}
{{- define "stackstate.otel.http.host" -}}
{{- if .Values.opentelemetry.enabled -}}
{{- index .Values "opentelemetry-collector" "fullnameOverride" }}
{{- else -}}
{{- .Values.stackstate.components.all.otelCollectorEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Comma-separated list of the endpoints that need to be up and running before the initializer can be started.
*/}}
{{ define "stackstate.initializer.prerequisites" -}}
{{- if .Values.clickhouse.enabled -}}
{{- include "stackstate.clickhouse.endpoint" . -}},
{{- end -}}
{{- include "stackstate.kafka.endpoint" . -}},
{{- if eq .Values.hbase.deployment.mode "Distributed" -}}
{{- include "stackstate.zookeeper.endpoint" . -}},
{{- .Release.Name }}-hbase-hdfs-nn-headful:9000
{{- else -}}
{{- .Release.Name }}-hbase-stackgraph:2182,
{{- end -}}
{{- end -}}

{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "stackstate.vmagent.endpoint" -}}
{{- if .Values.stackstate.components.vmagent.fullNameOverride -}}
{{- .Values.stackstate.components.vmagent.fullNameOverride -}}
{{- else -}}
http://{{ template "common.fullname.short" . }}-vmagent
{{- end -}}
{{- end -}}
