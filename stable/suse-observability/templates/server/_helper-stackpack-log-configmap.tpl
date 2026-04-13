{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.stackpack-log" }}
<appender name="StackPacks" class="ch.qos.logback.classic.sift.SiftingAppender">
    <discriminator class="ch.qos.logback.classic.sift.MDCBasedDiscriminator">
        <key>stackpack</key>
        <defaultValue>none</defaultValue>
    </discriminator>
    <sift>
    <appender name="Console" class="ch.qos.logback.core.ConsoleAppender">
        <encoder class="ch.qos.logback.classic.encoder.PatternLayoutEncoder">
            <pattern>%date %-5level ${stackpack} - %msg%n</pattern>
        </encoder>
    </appender>
    </sift>
</appender>

<logger name="com.stackstate.stackpackmanager.StackPackLogging" level="INFO" additivity="false">
    <appender-ref ref="StackPacks" />
</logger>
{{- end -}}
