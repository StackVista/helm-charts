{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.sync-log" }}
import ch.qos.logback.classic.sift.MDCBasedDiscriminator
import ch.qos.logback.classic.sift.SiftingAppender

appender("Syncs", SiftingAppender) {
  discriminator(MDCBasedDiscriminator) {
    key = "sync"
    defaultValue = "none"
  }

  sift {
    appender("Console", ConsoleAppender) {
      encoder(PatternLayoutEncoder) {
        pattern = "%date %-5level ${sync} - %msg%n"
      }
    }
  }
}

logger("com.stackstate.sync.SyncLogging", INFO, ["Syncs"], false)
logger("com.stackstate.sync.ExtTopoLogging", INFO, ["Syncs"], false)
{{- end -}}
