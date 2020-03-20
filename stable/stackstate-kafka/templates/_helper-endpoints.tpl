{{/*
Logic to determine Kafka endpoint.
*/}}
{{- define "stackstate-kafka.kafka.endpoint" -}}
{{- .Values.kafka.fullnameOverride -}}-headless:9092
{{- end -}}

{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "stackstate-kafka.zookeeper.endpoint" -}}
{{- if .Values.kafka.zookeeper.enabled -}}
{{- .Values.zookeeper.fullnameOverride -}}-headless:2181
{{- else -}}
{{- .Values.kafka.externalZookeeper.servers -}}
{{- end -}}
{{- end -}}
