{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "hbase.zookeeper.endpoint" -}}
{{- if .Values.zookeeper.enabled -}}
{{ .Release.Name }}-zookeeper-headless
{{- else -}}
{{- .Values.zookeeper.externalServers }}
{{- end -}}
{{- end -}}
