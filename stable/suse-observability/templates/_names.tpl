{{/*
Names for migrated resources. The canonical prefix is defined in the common chart.
*/}}
{{- define "stackstate.httpRoute.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}
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

{{/*
Keep the existing VictoriaMetrics PVC prefix used by backup and restore.
*/}}
{{- define "stackstate.victoriametrics.pvcPrefix" -}}
server-volume-{{ include "suse-observability.resourcePrefix" . }}-
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

{{/*
S3Proxy secret name.
*/}}
{{- define "stackstate.s3proxy.secretName" -}}
{{- if .Values.global.s3proxy.credentials.fromExternalSecret -}}
{{- .Values.global.s3proxy.credentials.fromExternalSecret -}}
{{- else -}}
{{- include "stackstate.s3proxy.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
S3 backend secret name.
Returns the external secret name if set, otherwise falls back to the s3proxy secret.
*/}}
{{- define "stackstate.s3proxy.s3BackendSecretName" -}}
{{- if .Values.backup.storage.backend.s3.fromExternalSecret -}}
{{- .Values.backup.storage.backend.s3.fromExternalSecret -}}
{{- else -}}
{{- include "stackstate.s3proxy.secretName" . -}}
{{- end -}}
{{- end -}}

{{/*
Azure backend secret name.
Returns the external secret name if set, otherwise falls back to the s3proxy secret.
*/}}
{{- define "stackstate.s3proxy.azureBackendSecretName" -}}
{{- if .Values.backup.storage.backend.azure.fromExternalSecret -}}
{{- .Values.backup.storage.backend.azure.fromExternalSecret -}}
{{- else -}}
{{- include "stackstate.s3proxy.secretName" . -}}
{{- end -}}
{{- end -}}

{{/*
Get the main backup PVC name.
For backward compatibility, we keep the old name "suse-observability-minio" since it was used in previous versions and may already exist in user clusters.
*/}}
{{- define "stackstate.backup.mainPvcName" -}}
{{ include "suse-observability.resourcePrefix" . }}-minio
{{- end -}}

{{/*
Service account name for S3Proxy.
Precedence: s3proxy.serviceAccount.name > minio.serviceAccount.name (deprecated) > S3Proxy fullname.
Set an explicit name to reuse an existing service account name, e.g. for IAM role bindings.
*/}}
{{- define "stackstate.s3proxy.serviceAccountName" -}}
{{- if .Values.s3proxy.serviceAccount.name -}}
{{- .Values.s3proxy.serviceAccount.name -}}
{{- else if and .Values.minio.serviceAccount .Values.minio.serviceAccount.name -}}
{{- .Values.minio.serviceAccount.name -}}
{{- else -}}
{{- include "stackstate.s3proxy.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Logic to determine ElasticSearch host.
*/}}
{{- define "stackstate.es.host" -}}
{{- include "stackstate.elasticsearch.fullname" . -}}-master-headless
{{- end -}}
