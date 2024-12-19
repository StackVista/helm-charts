{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.correlate-base-log" }}
    <logger name="org.apache.kafka" level="INFO"/>
    <logger name="com.stackstate" level="ERROR"/>
    <logger name="com.stackstate.correlate" level="${STACKSTATE_CORRELATE_LOG_LEVEL:-INFO}"/>
{{- end -}}
