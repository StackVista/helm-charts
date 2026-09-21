{{/*
Canonical prefix for migrated resources. Storage identities still use their existing helpers.
*/}}
{{- define "suse-observability.resourcePrefix" -}}
suse-observability
{{- end -}}

{{- define "stackstate.httpRoute.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}
{{- end -}}

{{- define "suse-observability.pullSecret.name" -}}
{{ include "suse-observability.resourcePrefix" . }}-pull-secret
{{- end -}}

{{/*
Hooks need a separate Secret because they can run before installation or after deletion.
*/}}
{{- define "suse-observability.pullSecret.hookName" -}}
{{ include "suse-observability.pullSecret.name" . }}-hook
{{- end -}}

{{- define "stackstate.victoriametrics.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-victoriametrics
{{- end -}}

{{- define "stackstate.victoriametrics.instance.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-victoria-metrics-{{ .instanceIndex }}
{{- end -}}

{{- define "stackstate.ui.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ui
{{- end -}}

{{- define "stackstate.replicationChecker.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-replication-checker
{{- end -}}

{{- define "stackstate.workloadObserver.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-workload-observer
{{- end -}}

{{- define "stackstate.backup.stackgraph.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-backup-sg
{{- end -}}

{{- define "stackstate.backup.config.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-backup-config
{{- end -}}

{{/*
The parent ConfigMap keeps this name even when the collector's fullname is overridden.
*/}}
{{- define "stackstate.otelCollector.defaultFullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-otel-collector
{{- end -}}

{{- define "stackstate.otelCollector.fullname" -}}
{{- index .Values "opentelemetry-collector" "fullnameOverride" | default (include "stackstate.otelCollector.defaultFullname" .) -}}
{{- end -}}

{{- define "stackstate.s3proxy.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{- define "stackstate.s3proxy.configmap.fullname" -}}
{{ include "stackstate.s3proxy.fullname" . }}-config
{{- end -}}

{{- define "stackstate.s3proxy.extraEnvSecret.fullname" -}}
{{ include "stackstate.s3proxy.fullname" . }}-extra-env
{{- end -}}

{{- define "stackstate.vmagent.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-vmagent
{{- end -}}

{{- define "stackstate.kafka.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-kafka
{{- end -}}

{{- define "stackstate.zookeeper.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-zookeeper
{{- end -}}

{{- define "stackstate.elasticsearch.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-elasticsearch
{{- end -}}

{{- define "stackstate.clickhouse.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-clickhouse
{{- end -}}

{{- define "stackstate.backup.clickhouse.backup.service" -}}
{{ include "stackstate.clickhouse.fullname" . }}-backup
{{- end -}}

{{/*
MCP fullname helper
*/}}
{{- define "stackstate.mcp.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-mcp
{{- end -}}

{{/*
AI Assistant fullname helper
*/}}
{{- define "stackstate.ai-assistant.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ai-assistant
{{- end -}}
