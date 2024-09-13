{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.stackpack-log" }}
import ch.qos.logback.classic.sift.MDCBasedDiscriminator
import ch.qos.logback.classic.sift.SiftingAppender

appender("StackPacks", SiftingAppender) {
  discriminator(MDCBasedDiscriminator) {
    key = "stackpack"
    defaultValue = "none"
  }

  sift {
    appender("Console", ConsoleAppender) {
      encoder(PatternLayoutEncoder) {
        pattern = "%date %-5level ${stackpack} - %msg%n"
      }
    }
  }
}

logger("com.stackstate.stackpackmanager.StackPackLogging", INFO, ["StackPacks"], false)
{{- end -}}
