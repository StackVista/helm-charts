{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.receiver-base-log" }}
<logger name="com.stackstate.api" level="INFO"/>
<logger name="com.stackstate.receiver" level="INFO"/>
<logger name="akka.actor.ActorSystemImpl" level="ERROR"/>

<logger name="com.stackstate.api.ReceiverRoutes" level="${STACKSTATE_RECEIVER_HTTP_LOG_LEVEL:-WARN}"/>
{{- end -}}
