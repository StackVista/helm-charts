{{/*
Receiver settings

<logger name="org.apache.pekko.actor.ActorSystemImpl" level="ERROR"/>
This is here to supress X-Forwarded-for spurious messages
*/}}
{{- define "stackstate.configmap.receiver-base-log" }}
<logger name="com.stackstate.api" level="INFO"/>
<logger name="com.stackstate.receiver" level="INFO"/>
<logger name="org.apache.pekko.actor.ActorSystemImpl" level="ERROR"/>
<logger name="com.stackstate.api.ReceiverRoutes" level="${STACKSTATE_RECEIVER_HTTP_LOG_LEVEL:-WARN}"/>
{{- end -}}
