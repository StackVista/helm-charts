{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.correlate-base-log" }}
<appender name="Workers" class="ch.qos.logback.classic.sift.SiftingAppender">
    <discriminator class="ch.qos.logback.classic.sift.MDCBasedDiscriminator">
        <key>worker</key>
        <defaultValue>none</defaultValue>
    </discriminator>
    <sift>
    <appender name="Console" class="ch.qos.logback.core.ConsoleAppender">
        <encoder class="ch.qos.logback.classic.encoder.PatternLayoutEncoder">
            <pattern>"date %-5level ${worker} - %msg%n</pattern>
        </encoder>
    </appender>
    </sift>
</appender>

<logger name="org.apache.kafka" level="INFO"/>
<logger name="com.stackstate" level="ERROR"/>
<logger name="com.stackstate.correlate" level="${STACKSTATE_CORRELATE_LOG_LEVEL:-INFO}">
    <appender-ref ref="Workers" />
</logger>
{{- end -}}
