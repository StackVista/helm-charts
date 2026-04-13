{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.e2es-base-log" }}
    <logger name="org.apache.kafka" level="INFO"/>
    <logger name="org.elasticsearch" level="INFO"/>
{{- end -}}
