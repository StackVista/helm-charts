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
{{- if .Values.minio.enabled -}}
{{- .Values.minio.fullnameOverride -}}:9000
{{- else -}}
{{- .Values.stackstate.components.all.minioEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine MinIO keys.
*/}}
{{- define "stackstate.minio.keys" -}}
{{- if .Values.minio.enabled -}}
{{- .Values.minio.fullnameOverride -}}
{{- else -}}
{{- .Values.stackstate.components.all.minioKeys -}}
{{- end -}}
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
