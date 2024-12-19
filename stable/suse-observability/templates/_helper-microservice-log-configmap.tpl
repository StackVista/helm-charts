{{/*
Shared settings in configmap for logging on stackstate microservices
*/}}
{{- define "stackstate.configmap.microservices-log" }}
<configuration scan="true" scanPeriod="5 seconds">
<!-- This can be used for debugging -->
    <!-- statusListener class="ch.qos.logback.core.status.OnConsoleStatusListener"/-->

    <appender name="Console" class="ch.qos.logback.core.ConsoleAppender">
        <filter class="ch.qos.logback.classic.filter.ThresholdFilter">
            <level>${log.console.level:-INFO}</level>
        </filter>
        <encoder class="ch.qos.logback.classic.encoder.PatternLayoutEncoder">
            <pattern>%date %-5level %logger{60} - %msg%n</pattern>
        </encoder>
    </appender>

    <root level="{{- .RootLevel -}}">
        <appender-ref ref="Console" />
    </root>

    <!-- Logging from values.yaml -->
    {{ .AdditionalLogging }}

    <!-- Custom logging configuration goes here -->
</configuration>
{{- end -}}
