{{/*
Logic to determine Zookeeper endpoint.
*/}}
{{- define "hbase.zookeeper.endpoint" -}}
{{- if .Values.zookeeper.enabled -}}
{{ include "call-nested" (list . "zookeeper" "zookeeper.fullname") }}
{{- else -}}
{{- .Values.zookeeper.externalServers }}
{{- end -}}
{{- end -}}
