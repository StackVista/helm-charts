{{/*
Logic to determine ElasticSearch endpoint.
*/}}
{{- define "distributed.es.endpoint" -}}
{{- if .Values.elasticsearch.enabled -}}
{{- .Values.elasticsearch.clusterName -}}-headless:9200
{{- else -}}
{{- .Values.stackstate.components.all.elasticsearchEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "distributed.kafka.endpoint" -}}
{{- if .Values.kafka.enabled -}}
{{- .Values.kafka.fullnameOverride -}}-headless:9092
{{- else -}}
{{- .Values.stackstate.components.all.kafkaEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "distributed.zookeeper.endpoint" -}}
{{- if .Values.zookeeper.enabled -}}
{{- .Values.zookeeper.fullnameOverride -}}-headless:2181
{{- else -}}
{{- .Values.stackstate.components.all.zookeeperEndpoint -}}
{{- end -}}
{{- end -}}
