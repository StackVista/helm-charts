{{/*
Endpoint-related helper templates for HBase chart.
These templates provide connection strings and URIs for dependent services.
*/}}

{{/*
Logic to determine Zookeeper endpoint.

Usage: {{ include "hbase.zookeeper.endpoint" . }}
Returns: Zookeeper connection string
*/}}
{{- define "hbase.zookeeper.endpoint" -}}
{{- .Values.zookeeper.externalServers }}
{{- end -}}
