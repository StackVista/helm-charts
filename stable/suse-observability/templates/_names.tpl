{{/*
Resource-name helpers shared by declarations and consumers.
The canonical prefix is defined in the common chart. Helpers retaining legacy
names are documented below so extraction can precede an actual naming migration.
*/}}

{{/*
Router workload and static configuration names retain their legacy expressions.
Keep these independent from the router Service, dynamic mode ConfigMaps and
hook resources, whose naming and upgrade contracts are separate.
*/}}
{{- define "stackstate.router.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-router
{{- end -}}

{{- define "stackstate.router.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-router
{{- end -}}

{{- define "stackstate.router.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-router
{{- end -}}

{{- define "stackstate.router.secret.fullname" -}}
{{ template "common.fullname.short" . }}-router
{{- end -}}

{{/*
The router Service uses the global hostname prefix so bundled subcharts resolve
the same address. Preserve its global overrides and truncation semantics; local
workload overrides do not apply. Keep the legacy stackstate.router.name helper
independent for customer templates that reference existing names.
*/}}
{{- define "stackstate.router.service.fullname" -}}
{{ template "stackstate.hostname.prefix" . }}-router
{{- end -}}

{{/*
Router mode configuration and hook identities retain their legacy names.
Automatic-mode scripts write the automatic ConfigMap; active and maintenance
ConfigMaps are chart-managed alternatives. Keep each identity independent.
*/}}
{{- define "stackstate.router.mode.active.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-router-active
{{- end -}}

{{- define "stackstate.router.mode.maintenance.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-router-maintenance
{{- end -}}

{{- define "stackstate.router.mode.automatic.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-router-automatic
{{- end -}}

{{/* Resolve the ConfigMap mounted by the router and reject unsupported modes. */}}
{{- define "stackstate.router.mode.configmap.fullname" -}}
{{- if eq .Values.stackstate.components.router.mode.status "active" -}}
{{ include "stackstate.router.mode.active.configmap.fullname" . }}
{{- else if eq .Values.stackstate.components.router.mode.status "maintenance" -}}
{{ include "stackstate.router.mode.maintenance.configmap.fullname" . }}
{{- else if eq .Values.stackstate.components.router.mode.status "automatic" -}}
{{ include "stackstate.router.mode.automatic.configmap.fullname" . }}
{{- else -}}
{{- fail "stackstate.components.router.mode.status must be one of: active, maintenance, automatic" -}}
{{- end -}}
{{- end -}}

{{- define "stackstate.router.mode.scripts.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-router-mode-scripts
{{- end -}}

{{- define "stackstate.router.mode.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-router-mode-scripts
{{- end -}}

{{- define "stackstate.router.mode.role.fullname" -}}
{{ template "common.fullname.short" . }}-router-mode
{{- end -}}

{{- define "stackstate.router.mode.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-router-mode
{{- end -}}

{{/* Preserve Helm's per-render timestamps and Argo CD's generated-name prefixes. */}}
{{- define "stackstate.router.mode.active.job.fullname" -}}
{{ template "common.fullname.short" . }}-set-active-{{ now | date "02t150405" }}
{{- end -}}

{{- define "stackstate.router.mode.maintenance.job.fullname" -}}
{{ template "common.fullname.short" . }}-set-maintenance-{{ now | date "02t150405" }}
{{- end -}}

{{- define "stackstate.router.mode.active.job.generateName" -}}
set-active-
{{- end -}}

{{- define "stackstate.router.mode.maintenance.job.generateName" -}}
set-maintenance-
{{- end -}}

{{/*
Internally managed license and email Secrets retain their legacy resource names.
External-secret selection stays in the existing stackstate.secret.name helpers;
customer-provided names must not receive a chart prefix.
*/}}
{{- define "stackstate.license.secret.fullname" -}}
{{ template "common.fullname.short" . }}-license
{{- end -}}

{{- define "stackstate.email.secret.fullname" -}}
{{ template "common.fullname.short" . }}-email
{{- end -}}

{{/*
The authentication Secret declaration and existing-password lookup share this
legacy identity. Renaming it requires a credential migration strategy; external
Secret selection remains in stackstate.secret.name.auth.
*/}}
{{- define "stackstate.auth.secret.fullname" -}}
{{ template "common.fullname.short" . }}-auth
{{- end -}}

{{/*
The receiver API-key Secret declaration and legacy lookup retain this identity.
External selection and optional Secret creation remain separate from naming.
*/}}
{{- define "stackstate.apiKey.secret.fullname" -}}
{{ template "common.fullname.short" . }}-api-key
{{- end -}}

{{/*
API configuration resource names. Preserve the existing naming expressions;
moving declarations and references to these helpers must not rename resources.
*/}}
{{- define "stackstate.api.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-api
{{- end -}}

{{- define "stackstate.api.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-api-log
{{- end -}}

{{- define "stackstate.api.secret.fullname" -}}
{{ template "common.fullname.short" . }}-api
{{- end -}}

{{/*
Checks, notification and synchronization configuration resources retain their
legacy names while declarations and consumers move to dedicated helpers.
*/}}
{{- define "stackstate.checks.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-checks
{{- end -}}

{{- define "stackstate.checks.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-checks-log
{{- end -}}

{{- define "stackstate.checks.secret.fullname" -}}
{{ template "common.fullname.short" . }}-checks
{{- end -}}

{{- define "stackstate.notification.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-notification
{{- end -}}

{{- define "stackstate.notification.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-notification-log
{{- end -}}

{{- define "stackstate.notification.secret.fullname" -}}
{{ template "common.fullname.short" . }}-notification
{{- end -}}

{{- define "stackstate.healthSync.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync
{{- end -}}

{{- define "stackstate.healthSync.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync-log
{{- end -}}

{{- define "stackstate.healthSync.secret.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync
{{- end -}}

{{- define "stackstate.authorizationSync.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync
{{- end -}}

{{- define "stackstate.authorizationSync.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync-log
{{- end -}}

{{- define "stackstate.authorizationSync.secret.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync
{{- end -}}

{{/*
State, sync and slicing configuration resources retain their existing names.
*/}}
{{- define "stackstate.state.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-state
{{- end -}}

{{- define "stackstate.state.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-state-log
{{- end -}}

{{- define "stackstate.state.secret.fullname" -}}
{{ template "common.fullname.short" . }}-state
{{- end -}}

{{- define "stackstate.sync.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-sync
{{- end -}}

{{- define "stackstate.sync.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-sync-log
{{- end -}}

{{- define "stackstate.sync.secret.fullname" -}}
{{ template "common.fullname.short" . }}-sync
{{- end -}}

{{- define "stackstate.slicing.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-slicing
{{- end -}}

{{- define "stackstate.slicing.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-slicing-log
{{- end -}}

{{- define "stackstate.slicing.secret.fullname" -}}
{{ template "common.fullname.short" . }}-slicing
{{- end -}}

{{/*
Server, receiver, correlate, initializer and e2es retain their existing configuration names.
Split receiver and correlate Deployments share their component configuration resources.
*/}}
{{- define "stackstate.server.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-server
{{- end -}}

{{- define "stackstate.server.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-server-log
{{- end -}}

{{- define "stackstate.server.secret.fullname" -}}
{{ template "common.fullname.short" . }}-server
{{- end -}}

{{- define "stackstate.receiver.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-receiver
{{- end -}}

{{- define "stackstate.receiver.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-receiver-log
{{- end -}}

{{- define "stackstate.receiver.secret.fullname" -}}
{{ template "common.fullname.short" . }}-receiver
{{- end -}}

{{- define "stackstate.correlate.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-correlate
{{- end -}}

{{- define "stackstate.correlate.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-correlate-log
{{- end -}}

{{- define "stackstate.correlate.secret.fullname" -}}
{{ template "common.fullname.short" . }}-correlate
{{- end -}}

{{- define "stackstate.initializer.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-initializer
{{- end -}}

{{- define "stackstate.initializer.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-initializer-log
{{- end -}}

{{- define "stackstate.initializer.secret.fullname" -}}
{{ template "common.fullname.short" . }}-initializer
{{- end -}}

{{- define "stackstate.e2es.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-e2es
{{- end -}}

{{- define "stackstate.e2es.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-e2es-log
{{- end -}}

{{- define "stackstate.e2es.secret.fullname" -}}
{{ template "common.fullname.short" . }}-e2es
{{- end -}}

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

{{/* Deployment identities are extracted independently of configuration resources.
Preserve the existing names while moving declarations to dedicated helpers. */}}
{{- define "stackstate.api.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-api
{{- end -}}

{{- define "stackstate.checks.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-checks
{{- end -}}

{{- define "stackstate.notification.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-notification
{{- end -}}

{{- define "stackstate.healthSync.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync
{{- end -}}

{{- define "stackstate.authorizationSync.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync
{{- end -}}

{{/* Deployment names for state, sync, slicing, server, initializer and e2es
retain their existing release and override behavior. */}}
{{- define "stackstate.state.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-state
{{- end -}}

{{- define "stackstate.sync.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-sync
{{- end -}}

{{- define "stackstate.slicing.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-slicing
{{- end -}}

{{- define "stackstate.server.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-server
{{- end -}}

{{- define "stackstate.initializer.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-initializer
{{- end -}}

{{- define "stackstate.e2es.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-e2es
{{- end -}}

{{/* Split worker suffixes are shared by Deployments, Services and ServiceMonitors.
Keep their existing names and type-dependent suffixes unchanged. */}}
{{- define "stackstate.receiver.name.postfix" -}}
  {{- if .ReceiverType }}-{{ .ReceiverType }}{{ else }}{{ end }}
{{- end -}}

{{- define "stackstate.receiver.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-receiver{{ template "stackstate.receiver.name.postfix" . }}
{{- end -}}

{{- define "stackstate.correlate.name.postfix" -}}
  {{- if .CorrelateType }}-{{ .CorrelateType }}{{ else }}{{ end }}
{{- end -}}

{{- define "stackstate.correlate.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-correlate{{ template "stackstate.correlate.name.postfix" . }}
{{- end -}}

{{/* RBAC resources have independent names, even when their legacy names match.
Preserve the namespace handling and truncation of the existing common helpers. */}}
{{- define "stackstate.getPods.role.fullname" -}}
{{ template "common.fullname.short" . }}-get-pods
{{- end -}}

{{- define "stackstate.getPods.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-get-pods
{{- end -}}

{{- define "stackstate.authorization.clusterrole.fullname" -}}
{{ template "common.fullname.cluster.unique" . }}-authorization
{{- end -}}

{{- define "stackstate.authorization.clusterrolebinding.fullname" -}}
{{ template "common.fullname.cluster.unique" . }}-authorization
{{- end -}}

{{- define "stackstate.authentication.clusterrolebinding.fullname" -}}
{{ template "common.fullname.cluster.unique" . }}-authentication
{{- end -}}

{{- define "stackstate.rbacAgent.role.fullname" -}}
{{ template "common.fullname.short" . }}-rbac-agent
{{- end -}}

{{- define "stackstate.rbacAgent.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-rbac-agent
{{- end -}}

{{/* Instance Role identities may be referenced by customer-managed bindings.
Preserve their legacy names. External group identities remain independent. */}}
{{- define "stackstate.k8s.authorization.instance.admin.role.fullname" -}}
{{ template "common.fullname.short" . }}-instance-admin
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.admin.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-instance-admin
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.observer.role.fullname" -}}
{{ template "common.fullname.short" . }}-instance-observer
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.observer.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-instance-observer
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.troubleshooter.role.fullname" -}}
{{ template "common.fullname.short" . }}-instance-troubleshooter
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.troubleshooter.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-instance-troubleshooter
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.basicAccess.role.fullname" -}}
{{ template "common.fullname.short" . }}-instance-basic-access
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.basicAccess.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-instance-basic-access
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.recommendedAccess.role.fullname" -}}
{{ template "common.fullname.short" . }}-instance-recommended-access
{{- end -}}

{{- define "stackstate.k8s.authorization.instance.recommendedAccess.rolebinding.fullname" -}}
{{ template "common.fullname.short" . }}-instance-recommended-access
{{- end -}}

{{/* ServiceAccount identities remain independent of Deployment and Service names.
Preserve the existing release and override behavior. */}}
{{- define "stackstate.api.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-api
{{- end -}}

{{- define "stackstate.server.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-server
{{- end -}}

{{- define "stackstate.checks.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-checks
{{- end -}}

{{- define "stackstate.notification.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-notification
{{- end -}}

{{- define "stackstate.healthSync.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync
{{- end -}}

{{- define "stackstate.authorizationSync.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync
{{- end -}}

{{- define "stackstate.initializer.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-initializer
{{- end -}}

{{- define "stackstate.slicing.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-slicing
{{- end -}}

{{- define "stackstate.state.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-state
{{- end -}}

{{- define "stackstate.sync.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-sync
{{- end -}}

{{/* Service identities retain their existing release and override behavior. */}}
{{- define "stackstate.checks.service.fullname" -}}
{{ template "common.fullname.short" . }}-checks
{{- end -}}

{{- define "stackstate.notification.service.fullname" -}}
{{ template "common.fullname.short" . }}-notification
{{- end -}}

{{- define "stackstate.healthSync.service.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync
{{- end -}}

{{- define "stackstate.state.service.fullname" -}}
{{ template "common.fullname.short" . }}-state
{{- end -}}

{{- define "stackstate.sync.service.fullname" -}}
{{ template "common.fullname.short" . }}-sync
{{- end -}}

{{- define "stackstate.slicing.service.fullname" -}}
{{ template "common.fullname.short" . }}-slicing
{{- end -}}

{{- define "stackstate.e2es.service.fullname" -}}
{{ template "common.fullname.short" . }}-e2es
{{- end -}}

{{- define "stackstate.api.service.fullname" -}}
{{ template "common.fullname.short" . }}-api-headless
{{- end -}}

{{- define "stackstate.server.service.fullname" -}}
{{ template "common.fullname.short" . }}-server-headless
{{- end -}}

{{- define "stackstate.initializer.service.fullname" -}}
{{ template "common.fullname.short" . }}-initializer
{{- end -}}

{{- define "stackstate.authorizationSync.service.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync
{{- end -}}

{{- define "stackstate.receiver.service.fullname" -}}
{{ template "common.fullname.short" . }}-receiver
{{- end -}}

{{- define "stackstate.receiver.base.service.fullname" -}}
{{ template "common.fullname.short" . }}-receiver-base
{{- end -}}

{{- define "stackstate.receiver.logs.service.fullname" -}}
{{ template "common.fullname.short" . }}-receiver-logs
{{- end -}}

{{- define "stackstate.receiver.processAgent.service.fullname" -}}
{{ template "common.fullname.short" . }}-receiver-process-agent
{{- end -}}

{{- define "stackstate.correlate.service.fullname" -}}
{{ template "common.fullname.short" . }}-correlate{{ template "stackstate.correlate.name.postfix" . }}
{{- end -}}

{{/* Collector requests use the base receiver in split mode and the single
receiver otherwise. */}}
{{- define "stackstate.receiver.target.service.fullname" -}}
{{- if eq (include "stackstate.receiver.split.enabled" .) "true" -}}
{{- include "stackstate.receiver.base.service.fullname" . -}}
{{- else -}}
{{- include "stackstate.receiver.service.fullname" . -}}
{{- end -}}
{{- end -}}

{{/* API clients use the API Service in split mode and the server Service otherwise.
This selects a Service identity; Envoy cluster identifiers remain independent. */}}
{{- define "stackstate.api.target.service.fullname" -}}
{{- if include "suse-observability.features.enabled" (dict "key" "server.split" "context" .) -}}
{{- include "stackstate.api.service.fullname" . -}}
{{- else -}}
{{- include "stackstate.server.service.fullname" . -}}
{{- end -}}
{{- end -}}

{{/* Authorization clients use authorization-sync in split mode and the server
Service otherwise. */}}
{{- define "stackstate.authorizationSync.target.service.fullname" -}}
{{- if include "suse-observability.features.enabled" (dict "key" "server.split" "context" .) -}}
{{- include "stackstate.authorizationSync.service.fullname" . -}}
{{- else -}}
{{- include "stackstate.server.service.fullname" . -}}
{{- end -}}
{{- end -}}
