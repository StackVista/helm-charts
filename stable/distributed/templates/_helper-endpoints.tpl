{{/*
Logic to determine ElasticSearch endpoint.
*/}}
{{- define "distributed.es.endpoint" -}}
{{- if .Values.elasticsearch.enabled -}}
{{- .Values.elasticsearch.clusterName -}}-{{ .Values.elasticsearch.nodeGroup }}
{{- else -}}
{{- .Values.stackstate.components.all.elasticsearchEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "distributed.kafka.endpoint" -}}
{{- if .Values.kafka.enabled -}}
{{- .Values.kafka.fullnameOverride -}}
{{- else -}}
{{- .Values.stackstate.components.all.kafkaEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine Kafka endpoints as a comma-separated list.
*/}}
{{- define "distributed.kafka.separateEndpoints" -}}
{{- $clusterDomain := .Values.clusterDomain -}}
{{- $replicaCount := int .Values.kafka.replicaCount -}}
{{- $releaseNamespace := .Release.Namespace -}}
{{- $kafkaEndpoint := include "distributed.kafka.endpoint" . -}}
{{- $kafkaHeadlessServiceName := printf "%s-%s" $kafkaEndpoint "headless" | trunc 63 -}}
{{ range $i, $e := until $replicaCount }}{{ if $i }},{{ end }}{{ $kafkaEndpoint }}-{{ $e }}.{{ $kafkaHeadlessServiceName }}.{{ $releaseNamespace }}.svc.{{ $clusterDomain }}:9092{{ end }}
{{- end -}}

{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "distributed.zookeeper.endpoint" -}}
{{- if .Values.zookeeper.enabled -}}
{{- .Values.zookeeper.fullnameOverride -}}
{{- else -}}
{{- .Values.stackstate.components.all.zookeeperEndpoint -}}
{{- end -}}
{{- end -}}
