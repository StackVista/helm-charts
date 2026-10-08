{{/*
Resource-name helpers shared by declarations and consumers.
The canonical prefix is defined in the common chart. Helpers retaining legacy
names are documented below so extraction can precede an actual naming migration.
*/}}

{{/*
The stateless router Deployment uses the canonical product name. Its Service,
static and dynamic configuration, accounts, Secrets and hook resources retain
their identities. Mode scripts select Deployments by stable release labels so
pre-upgrade hooks reach the old router and post-upgrade hooks reach the new one.
*/}}
{{- define "stackstate.router.deployment.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-router
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
The shared Secret holds common environment variables and inline trust material.
Keep its legacy identity consistent with the producer and blob resolvers;
external Secret names and keys are selected independently.
*/}}
{{- define "stackstate.common.secret.fullname" -}}
{{ template "common.fullname.short" . }}-common
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

{{/*
Legacy collector base name, retained for endpoint/Envoy naming compatibility.
The parent endpoint ConfigMap has its own shared helper in common.
*/}}
{{- define "stackstate.otelCollector.defaultFullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-otel-collector
{{- end -}}

{{- define "stackstate.otelCollector.fullname" -}}
{{- index .Values "opentelemetry-collector" "fullnameOverride" | default (include "stackstate.otelCollector.defaultFullname" .) -}}
{{- end -}}

{{/*
Resolve the collector Service in its subchart context. Keep the legacy address
when the collector is disabled or its generic fullname differs from the old
parent expression (e.g. cleared/overlong overrides). Correcting those existing
endpoint mismatches is separate from extracting names without changing renders.
Envoy cluster identifiers continue to use stackstate.otelCollector.fullname.
*/}}
{{- define "stackstate.otelCollector.service.fullname" -}}
{{- $legacy := include "stackstate.otelCollector.fullname" . -}}
{{- $collector := index .Subcharts "opentelemetry-collector" -}}
{{- if and $collector (eq $legacy (include "opentelemetry-collector.fullname" $collector)) -}}
{{- include "opentelemetry-collector.service.fullname" $collector -}}
{{- else -}}
{{- $legacy -}}
{{- end -}}
{{- end -}}

{{- define "stackstate.s3proxy.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{- define "stackstate.s3proxy.configmap.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy-config
{{- end -}}

{{- define "stackstate.s3proxy.extraEnvSecret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy-extra-env
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
{{- include "stackstate.s3proxy.secret.fullname" . -}}
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
{{- include "stackstate.s3proxy.main.persistentvolumeclaim.fullname" . -}}
{{- end -}}

{{/* Legacy interface; keep the settings claim's release-dependent identity. */}}
{{- define "stackstate.backup.settingsPvcName" -}}
{{- include "stackstate.s3proxy.settings.persistentvolumeclaim.fullname" . -}}
{{- end -}}

{{/*
Service account name for S3Proxy.
Precedence: s3proxy.serviceAccount.name > minio.serviceAccount.name (deprecated) > default ServiceAccount fullname.
Set an explicit name to reuse an existing service account name, e.g. for IAM role bindings.
*/}}
{{- define "stackstate.s3proxy.serviceAccountName" -}}
{{- if .Values.s3proxy.serviceAccount.name -}}
{{- .Values.s3proxy.serviceAccount.name -}}
{{- else if and .Values.minio.serviceAccount .Values.minio.serviceAccount.name -}}
{{- .Values.minio.serviceAccount.name -}}
{{- else -}}
{{- include "stackstate.s3proxy.serviceaccount.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Resolve the Elasticsearch headless Service in its subchart context when it
matches the previous platform endpoint. Preserve existing endpoint mismatches
for customer naming overrides and non-master groups.
*/}}
{{- define "stackstate.es.host" -}}
{{- $legacy := printf "%s-master-headless" (include "stackstate.elasticsearch.fullname" .) -}}
{{- $elasticsearch := index .Subcharts "elasticsearch" -}}
{{- if and $elasticsearch (eq $elasticsearch.Values.nodeGroup "master") (eq $legacy (printf "%s-headless" (include "elasticsearch.masterService" $elasticsearch))) -}}
{{- include "elasticsearch.headless.service.fullname" $elasticsearch -}}
{{- else -}}
{{- $legacy -}}
{{- end -}}
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

{{/*
Persistent claim identities retain their legacy expressions, independently of
Deployment names. Do not switch these helpers to a canonical prefix without a
storage migration plan: changing a claim name does not move its existing data.
Stackpack claims are also consumed by the backup configuration.
*/}}
{{- define "stackstate.api.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-api-txlog
{{- end -}}

{{- define "stackstate.authorizationSync.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-authorization-sync-txlog
{{- end -}}

{{- define "stackstate.checks.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-checks-txlog
{{- end -}}

{{- define "stackstate.checks.tmp.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-checks-tmp
{{- end -}}

{{- define "stackstate.healthSync.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync-txlog
{{- end -}}

{{- define "stackstate.healthSync.tmp.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-health-sync-tmp
{{- end -}}

{{- define "stackstate.notification.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-notification-txlog
{{- end -}}

{{- define "stackstate.state.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-state-txlog
{{- end -}}

{{- define "stackstate.state.tmp.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-state-tmp
{{- end -}}

{{- define "stackstate.sync.txlog.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-sync-txlog
{{- end -}}

{{- define "stackstate.sync.tmp.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-sync-tmp
{{- end -}}

{{- define "stackstate.stackpacks.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-stackpacks
{{- end -}}

{{- define "stackstate.stackpacks.local.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-stackpacks-local
{{- end -}}

{{/* Keep the stackpack scripts ConfigMap independent of the stackpack PVCs. */}}
{{- define "stackstate.stackpacks.scripts.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-stackpacks-scripts
{{- end -}}

{{/*
Backup resource names retain their existing fixed or release-derived identities.
Keep each resource independent, including ConfigMaps and Secrets that currently
share a name. Backup PVC renames require a data migration plan. Hook timestamps
and Argo CD generated-name prefixes are part of the existing upgrade contract.
*/}}
{{- define "stackstate.backup.log.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-backup-log
{{- end -}}

{{- define "stackstate.backup.restore.scripts.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-backup-restore-scripts
{{- end -}}

{{- define "stackstate.backup.configuration.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-sts-backup-conf
{{- end -}}

{{- define "stackstate.backup.stackpacks.service.fullname" -}}
{{ template "common.fullname.short" . }}-backup-stackpacks
{{- end -}}

{{- define "stackstate.backup.stackgraph.tmp.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-backup-stackgraph-tmp-data
{{- end -}}

{{- define "stackstate.backup.stackgraph.v2.tmp.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-backup-stackgraph-v2-tmp-data
{{- end -}}

{{- define "stackstate.backup.configuration.persistentvolumeclaim.fullname" -}}
{{ template "common.fullname.short" . }}-settings-backup-data
{{- end -}}

{{- define "stackstate.backup.stackgraph.v2.cronjob.fullname" -}}
{{ template "common.fullname.short" . }}-backup-sg-v2
{{- end -}}

{{- define "stackstate.backup.configuration.cronjob.fullname" -}}
{{ template "common.fullname.short" . }}-backup-conf
{{- end -}}

{{- define "stackstate.backup.init.job.fullname" -}}
{{ template "common.fullname.short" . }}-backup-init-{{ now | date "02t150405" }}
{{- end -}}

{{- define "stackstate.backup.configuration.init.job.fullname" -}}
{{ template "common.fullname.short" . }}-init-pvc-{{ now | date "02t150405" }}
{{- end -}}

{{- define "stackstate.backup.clickhouse.cleanup.job.fullname" -}}
{{ template "common.fullname.short" . }}-ch-clean{{ now | date "02t150405" }}
{{- end -}}

{{- define "stackstate.backup.init.cronjob.fullname" -}}
{{ template "common.fullname.short" . }}-backup-init
{{- end -}}

{{- define "stackstate.backup.init.job.generateName" -}}
backup-init-
{{- end -}}

{{- define "stackstate.backup.configuration.init.job.generateName" -}}
init-pvc-
{{- end -}}

{{- define "stackstate.backup.clickhouse.cleanup.job.generateName" -}}
ch-clean
{{- end -}}

{{- define "stackstate.backup.config.configmap.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-backup-config
{{- end -}}

{{- define "stackstate.backup.config.secret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-backup-config
{{- end -}}

{{- define "stackstate.backup.stackgraph.cronjob.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-backup-sg
{{- end -}}

{{/* Manual backup/restore resources keep fixed names; their YAML data keys
and application-level snapshot/bucket identifiers are independent. */}}
{{- define "stackstate.backup.elasticsearch.list.job.fullname" -}}
elasticsearch-list-snapshots
{{- end -}}

{{- define "stackstate.backup.elasticsearch.restore.job.fullname" -}}
elasticsearch-restore-snapshot
{{- end -}}

{{- define "stackstate.backup.stackgraph.list.job.fullname" -}}
stackgraph-list-backups
{{- end -}}

{{- define "stackstate.backup.stackgraph.restore.job.fullname" -}}
stackgraph-restore-backup
{{- end -}}

{{- define "stackstate.backup.configuration.list.job.fullname" -}}
configuration-list-backups
{{- end -}}

{{- define "stackstate.backup.configuration.restore.job.fullname" -}}
configuration-restore-backup
{{- end -}}

{{- define "stackstate.backup.configuration.download.job.fullname" -}}
configuration-download-backup
{{- end -}}

{{- define "stackstate.backup.configuration.upload.job.fullname" -}}
configuration-upload-backup
{{- end -}}

{{- define "stackstate.backup.victoriaMetrics.list.job.fullname" -}}
victoria-metrics-list-backups
{{- end -}}

{{- define "stackstate.backup.victoriaMetrics.restore.job.fullname" -}}
victoria-metrics-restore-backup
{{- end -}}

{{- define "stackstate.backup.stackgraph.restore.persistentvolumeclaim.fullname" -}}
stackgraph-restore-backup
{{- end -}}

{{/*
PDBs always use the canonical prefix, independently of release and fullname
settings. Keep selectors and disruption budgets independent of names;
receiver/correlate retain one PDB each in split and unsplit modes.
Helm replacement temporarily blocks evictions while both budget names coexist.
*/}}
{{- define "stackstate.aiAssistant.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ai-assistant
{{- end -}}

{{- define "stackstate.api.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-api
{{- end -}}

{{- define "stackstate.authorizationSync.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-authorization-sync
{{- end -}}

{{- define "stackstate.checks.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-checks
{{- end -}}

{{- define "stackstate.correlate.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-correlate
{{- end -}}

{{- define "stackstate.e2es.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-e2es
{{- end -}}

{{- define "stackstate.victoriametrics.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-victoriametrics
{{- end -}}

{{- define "stackstate.healthSync.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-health-sync
{{- end -}}

{{- define "stackstate.mcp.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-mcp
{{- end -}}

{{- define "stackstate.notification.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-notification
{{- end -}}

{{- define "stackstate.receiver.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-receiver
{{- end -}}

{{- define "stackstate.router.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-router
{{- end -}}

{{- define "stackstate.server.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-server
{{- end -}}

{{- define "stackstate.state.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-state
{{- end -}}

{{- define "stackstate.sync.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-sync
{{- end -}}

{{- define "stackstate.ui.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ui
{{- end -}}

{{- define "stackstate.vmagent.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-vmagent
{{- end -}}

{{- define "stackstate.workloadObserver.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-workload-observer
{{- end -}}

{{/*
ServiceMonitor resource identities are independent of the Services they select.
Use the canonical prefix regardless of release and fullname settings.
Receiver/correlate variants each have their own helper; the shared renderers
receive resolved names while component types continue to control labels only.
*/}}
{{- define "stackstate.api.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-api
{{- end -}}

{{- define "stackstate.authorizationSync.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-authorization-sync
{{- end -}}

{{- define "stackstate.checks.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-checks
{{- end -}}

{{- define "stackstate.correlate.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-correlate
{{- end -}}

{{- define "stackstate.correlate.connection.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-correlate-connection
{{- end -}}

{{- define "stackstate.correlate.httpTracing.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-correlate-http-tracing
{{- end -}}

{{- define "stackstate.correlate.aggregator.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-correlate-aggregator
{{- end -}}

{{- define "stackstate.e2es.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-e2es
{{- end -}}

{{- define "stackstate.healthSync.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-health-sync
{{- end -}}

{{- define "stackstate.initializer.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-initializer
{{- end -}}

{{- define "stackstate.notification.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-notification
{{- end -}}

{{- define "stackstate.receiver.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-receiver
{{- end -}}

{{- define "stackstate.receiver.base.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-receiver-base
{{- end -}}

{{- define "stackstate.receiver.logs.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-receiver-logs
{{- end -}}

{{- define "stackstate.receiver.processAgent.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-receiver-process-agent
{{- end -}}

{{- define "stackstate.router.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-router
{{- end -}}

{{- define "stackstate.s3proxy.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{- define "stackstate.server.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-server
{{- end -}}

{{- define "stackstate.slicing.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-slicing
{{- end -}}

{{- define "stackstate.state.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-state
{{- end -}}

{{- define "stackstate.sync.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-sync
{{- end -}}

{{- define "stackstate.ui.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ui
{{- end -}}

{{/*
UI, MCP and AI Assistant resources retain their existing canonical names.
Keep each resource identity independent of Services, accounts and Secrets, even
when the outputs match. Legacy generic helpers remain available for Envoy
identifiers and external templates. AI Assistant StatefulSet renaming requires a
separate storage migration: its generated PVC identities depend on that name.
*/}}
{{- define "stackstate.ui.deployment.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ui
{{- end -}}

{{- define "stackstate.ui.service.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ui
{{- end -}}

{{- define "stackstate.ui.secret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ui
{{- end -}}

{{- define "stackstate.mcp.deployment.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-mcp
{{- end -}}

{{- define "stackstate.mcp.service.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-mcp
{{- end -}}

{{- define "stackstate.mcp.secret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-mcp
{{- end -}}

{{- define "stackstate.mcp.serviceaccount.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-mcp
{{- end -}}

{{- define "stackstate.aiAssistant.statefulset.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ai-assistant
{{- end -}}

{{- define "stackstate.aiAssistant.service.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ai-assistant
{{- end -}}

{{- define "stackstate.aiAssistant.secret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ai-assistant
{{- end -}}

{{- define "stackstate.aiAssistant.serviceaccount.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-ai-assistant
{{- end -}}

{{/*
Operational controllers retain their existing canonical identities. Keep each
resource independent and retain the generic helpers for external templates.
The workload observer and vmagent StatefulSet names determine generated PVC
identities; changing their outputs requires a separate storage migration.
*/}}
{{- define "stackstate.replicationChecker.deployment.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-replication-checker
{{- end -}}

{{- define "stackstate.replicationChecker.serviceaccount.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-replication-checker
{{- end -}}

{{- define "stackstate.replicationChecker.role.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-replication-checker
{{- end -}}

{{- define "stackstate.replicationChecker.rolebinding.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-replication-checker
{{- end -}}

{{- define "stackstate.workloadObserver.statefulset.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-workload-observer
{{- end -}}

{{- define "stackstate.workloadObserver.serviceaccount.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-workload-observer
{{- end -}}

{{- define "stackstate.workloadObserver.role.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-workload-observer
{{- end -}}

{{- define "stackstate.workloadObserver.rolebinding.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-workload-observer
{{- end -}}

{{- define "stackstate.vmagent.statefulset.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-vmagent
{{- end -}}

{{- define "stackstate.vmagent.configmap.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-vmagent
{{- end -}}

{{- define "stackstate.vmagent.service.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-vmagent
{{- end -}}

{{/*
S3Proxy resources retain their current names. Secret and ServiceAccount selectors
above preserve external credentials and explicit/deprecated account overrides;
these helpers supply only their internally managed/default identities.
*/}}
{{- define "stackstate.s3proxy.deployment.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{- define "stackstate.s3proxy.service.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{- define "stackstate.s3proxy.secret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{- define "stackstate.s3proxy.serviceaccount.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-s3proxy
{{- end -}}

{{/*
Existing S3Proxy storage identities. Preserve the legacy settings prefix and
MinIO claim name: changing either output requires a data migration strategy.
*/}}
{{- define "stackstate.s3proxy.settings.persistentvolumeclaim.fullname" -}}
{{- include "common.fullname.short" . -}}-backup-settings-data
{{- end -}}

{{- define "stackstate.s3proxy.main.persistentvolumeclaim.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-minio
{{- end -}}

{{/*
Remaining global identities retain their legacy names and scope. Hook names
must remain distinct from the regular pull Secret. Keep the existing timestamp
and Argo CD generateName behavior for topic-creation Jobs.
*/}}
{{- define "suse-observability.pullSecret.hook.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-pull-secret-hook
{{- end -}}

{{- define "stackstate.ingress.fullname" -}}
{{ include "common.fullname.short" . }}
{{- end -}}

{{- define "stackstate.securitycontextconstraints.fullname" -}}
{{ template "common.fullname.short" . }}-{{ .Release.Namespace }}
{{- end -}}

{{- define "stackstate.kafkaTopicCreate.job.fullname" -}}
{{ template "common.fullname.short" . }}-topic-create-{{ now | date "02t150405" }}
{{- end -}}

{{- define "stackstate.kafkaTopicCreate.job.generateName" -}}
topic-create-
{{- end -}}

{{- define "stackstate.victoriametrics.service.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-victoriametrics
{{- end -}}
