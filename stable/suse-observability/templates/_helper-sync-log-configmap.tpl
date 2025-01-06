{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.sync-log" }}
<appender name="Syncs" class="ch.qos.logback.classic.sift.SiftingAppender">
    <discriminator class="ch.qos.logback.classic.sift.MDCBasedDiscriminator">
        <key>sync</key>
        <defaultValue>none</defaultValue>
    </discriminator>
    <sift>
    <appender name="Console" class="ch.qos.logback.core.ConsoleAppender">
        <encoder class="ch.qos.logback.classic.encoder.PatternLayoutEncoder">
            <pattern>%date %-5level ${sync} - %msg%n</pattern>
        </encoder>
    </appender>
    </sift>
</appender>

<logger name="com.stackstate.sync.SyncLogging" level="INFO">
    <appender-ref ref="Syncs" />
</logger>
<logger name="com.stackstate.sync.ExtTopoLogging" level="INFO">
    <appender-ref ref="Syncs" />
</logger>
{{- end -}}
