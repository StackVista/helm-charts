{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "hbase.zookeeper.endpoint" -}}
{{- .Values.zookeeper.externalServers }}
{{- end -}}
