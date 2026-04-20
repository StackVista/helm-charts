{{/*
Logic to determine remote write endpoint for single node deployment of Victoria Metrics.
*/}}
{{- define "stackstate.metrics.victoriametrics.singleNode.remotewrite.url" -}}
http://{{- include "stackstate.metrics.victoriametrics.singleNode.remoteWriteEndpoint" . -}}{{ include "stackstate.metricStore.remoteWritePath" . }}
{{- end -}}

{{/*
Logic to determine promql query endpoint. It
*/}}
{{- define "stackstate.metrics.query.url" -}}
http://suse-observability-victoriametrics:8428
{{- end -}}

{{/*
Logic to determine metric store host and port for single node deployment of Victoria Metrics.
*/}}
{{- define "stackstate.metrics.victoriametrics.singleNode.remoteWriteEndpoint" -}}
suse-observability-victoria-metrics-{{ .instanceIndex }}:8428
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
{{- include "stackstate.elasticsearch.fullname" . -}}-master-headless:9200
{{- end -}}

{{/*
Logic to determine ElasticSearch host.
*/}}
{{- define "stackstate.es.host" -}}
{{- include "stackstate.elasticsearch.fullname" . -}}-master-headless
{{- end -}}

{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "stackstate.kafka.endpoint" -}}
{{- include "stackstate.kafka.fullname" . -}}-headless:9092
{{- end -}}

{{/*
Logic to determine Router endpoint.
*/}}
{{- define "stackstate.router.endpoint" -}}
http://{{ template "stackstate.router.name" . }}:8080
{{- end -}}

{{/*
Logic to determine MinIO keys.
*/}}
{{- define "stackstate.minio.keys" -}}
{{- include "stackstate.s3proxy.secretName" . -}}
{{- end -}}

{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "stackstate.zookeeper.endpoint" -}}
{{- include "stackstate.zookeeper.fullname" . -}}-headless:2181
{{- end -}}

{{/*
Clickhouse endpoint.
*/}}
{{- define "stackstate.clickhouse.endpoint" -}}
{{- include "stackstate.clickhouse.fullname" . -}}-headless:8123
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
Comma-separated list of the endpoints that need to be up to consider hdfs running
*/}}
{{ define "stackgraph.hbase.waitfor" -}}
{{- $deploymentMode := include "suse-observability.hbase.deploymentMode" . -}}
{{- if eq $deploymentMode "Distributed" -}}
{{- include "stackstate.zookeeper.endpoint" . }},{{ .Release.Name }}-hbase-hdfs-nn-headful:9000
{{- else -}}
{{- .Release.Name }}-hbase-stackgraph:2182
{{- end -}}
{{- end -}}

{{/*
Comma-separated list of the endpoints that need to be up and running before the initializer can be started.
*/}}
{{ define "stackstate.initializer.prerequisites" -}}
{{- include "stackstate.clickhouse.endpoint" . -}},
{{- include "stackstate.kafka.endpoint" . -}},
{{- include "stackgraph.hbase.waitfor" . -}}
{{- end -}}

{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "stackstate.vmagent.endpoint" -}}
{{- include "stackstate.vmagent.fullname" . -}}
{{- end -}}

{{/*
Logic to determine Zookeeper endpoint for stackgraph.
*/}}
{{- define "stackgraph.zookeeper.endpoint" -}}
{{- $deploymentMode := include "suse-observability.hbase.deploymentMode" . -}}
{{- if eq $deploymentMode "Distributed" }}
{{- include "stackstate.zookeeper.endpoint" . }}
{{- else }}
{{- .Release.Name }}-hbase-stackgraph:2182
{{- end }}
{{- end -}}
