{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "stackstate-k8s-agent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "stackstate-k8s-agent.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Release name of the chart. Can be used in subcharts because it does not use anything from .Values
*/}}
{{- define "stackstate-k8s-agent.releasename" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "stackstate-k8s-agent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "stackstate-k8s-agent.labels" -}}
app.kubernetes.io/name: {{ include "stackstate-k8s-agent.name" . }}
helm.sh/chart: {{ include "stackstate-k8s-agent.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Cluster agent checksum annotations
*/}}
{{- define "stackstate-k8s-agent.checksum-configs" }}
checksum/secret: {{ include (print $.Template.BasePath "/secret.yaml") . | sha256sum }}
{{- end }}

{{/*
StackState URL function
*/}}
{{- define "stackstate-k8s-agent.stackstate.url" -}}
{{- if not (hasPrefix "http" (tpl .Values.stackstate.url .)) -}}
{{- fail "SUSE Observability Ingest URL should start with the http or https protocol." -}}
{{- end -}}
{{ tpl .Values.stackstate.url . | quote }}
{{- end }}

{{- define "stackstate-k8s-agent.clusterNameEnvValue" -}}
{{- if .Values.global.clusterName.fromSecret }}
valueFrom:
  secretKeyRef:
    name: {{ tpl .Values.global.clusterName.fromSecret . | quote }}
    key: STS_CLUSTER_NAME
{{- else }}
value: {{ .Values.stackstate.cluster.name | quote }}
{{- end }}
{{- end }}

{{- define "stackstate-k8s-agent.urlEnvValue" -}}
{{- if .Values.global.url.fromSecret }}
valueFrom:
  secretKeyRef:
    name: {{ tpl .Values.global.url.fromSecret . | quote }}
    key: STS_URL
{{- else }}
value: {{ include "stackstate-k8s-agent.stackstate.url" . }}
{{- end }}
{{- end }}

{{- define "stackstate-k8s-agent.hostname" -}}
{{- if .Values.global.clusterName.fromSecret -}}
$(KUBERNETES_HOSTNAME)-$(STS_CLUSTER_NAME)
{{- else -}}
$(KUBERNETES_HOSTNAME)-{{ .Values.stackstate.cluster.name }}
{{- end -}}
{{- end }}

{{- define "stackstate-k8s-agent.platformUrlEnv" -}}
{{- if .Values.global.url.fromSecret }}
- name: STS_URL
  {{- include "stackstate-k8s-agent.urlEnvValue" . | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Derive platform OTLP endpoint from StackState URL or use explicit override.
Default: appends /otel to stackstate.url.
Override: otel.platformHttpOtlpEndpoint takes precedence over otel.platformGrpcOtlpEndpoint
when both are set. See stackstate-k8s-agent.platform.otlp.useGrpc.
*/}}
{{- define "stackstate-k8s-agent.platform.otlp.endpoint" -}}
{{- if .Values.otel.platformHttpOtlpEndpoint -}}
  {{- $endpoint := .Values.otel.platformHttpOtlpEndpoint -}}
  {{- if not (or (hasPrefix "http://" $endpoint) (hasPrefix "https://" $endpoint)) -}}
    {{- fail "otel.platformHttpOtlpEndpoint must start with http:// or https://" -}}
  {{- end -}}
  {{- $endpoint -}}
{{- else if .Values.otel.platformGrpcOtlpEndpoint -}}
  {{- $endpoint := .Values.otel.platformGrpcOtlpEndpoint -}}
  {{- if or (hasPrefix "http://" $endpoint) (hasPrefix "https://" $endpoint) -}}
    {{- fail "otel.platformGrpcOtlpEndpoint must not include an http(s):// scheme (format: host:port)" -}}
  {{- end -}}
  {{- if not (contains ":" $endpoint) -}}
    {{- fail "otel.platformGrpcOtlpEndpoint must include a port (format: host:port, e.g. otlp-my-instance.example.com:443)" -}}
  {{- end -}}
  {{- $endpoint -}}
{{- else if .Values.global.url.fromSecret -}}
  $(STS_URL)/otel
{{- else -}}
  {{- printf "%s/otel" (tpl .Values.stackstate.url . | trimSuffix "/") -}}
{{- end -}}
{{- end }}

{{/*
Returns "true" when gRPC should be used: otel.platformGrpcOtlpEndpoint is set and
otel.platformHttpOtlpEndpoint is not. Empty string otherwise (HTTP is the default).
*/}}
{{- define "stackstate-k8s-agent.platform.otlp.useGrpc" -}}
{{- if and (not .Values.otel.platformHttpOtlpEndpoint) .Values.otel.platformGrpcOtlpEndpoint -}}
true
{{- end -}}
{{- end }}

{{/*
Returns "true" when the HTTP OTLP endpoint uses plain http:// (no TLS), so the
exporter must be configured with insecure: true.
*/}}
{{- define "stackstate-k8s-agent.platform.otlp.insecure" -}}
{{- if hasPrefix "http://" (.Values.otel.platformHttpOtlpEndpoint | default "") -}}
true
{{- end -}}
{{- end }}

{{- define "stackstate-k8s-agent.configmap.override.checksum" -}}
{{- if .Values.clusterAgent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/cluster-agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}

{{- define "stackstate-k8s-agent.nodeAgent.configmap.override.checksum" -}}
{{- if .Values.nodeAgent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/node-agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}

{{- define "stackstate-k8s-agent.logsAgent.configmap.override.checksum" -}}
checksum/override-configmap: {{ include (print $.Template.BasePath "/logs-agent-configmap.yaml") . | sha256sum }}
{{- end }}

{{- define "stackstate-k8s-agent.otelLogsAgent.configmap.override.checksum" -}}
checksum/override-configmap: {{ include (print $.Template.BasePath "/otel/logsagent/configmap.yaml") . | sha256sum }}
{{- end }}

{{- define "stackstate-k8s-agent.checksAgent.configmap.override.checksum" -}}
{{- if .Values.checksAgent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/checks-agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}


{{/*
Return the image registry
*/}}
{{- define "stackstate-k8s-agent.imageRegistry" -}}
  {{- if .Values.global }}
    {{- .Values.global.imageRegistry | default .Values.all.image.registry -}}
  {{- else -}}
    {{- .Values.all.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Render a fully qualified external image reference. Unlike the StackState-owned
agent images, these upstream images must not be prefixed with global.imageRegistry.
*/}}
{{- define "stackstate-k8s-agent.externalImage" -}}
{{- if .image.tag -}}
{{- printf "%s:%s" .image.repository .image.tag -}}
{{- else -}}
{{- .image.repository -}}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains a template.
Usage:
{{ include "stackstate-k8s-agent.tplvalue.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "stackstate-k8s-agent.tplvalue.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.pull-secret.name" -}}
{{ include "stackstate-k8s-agent.fullname" . }}-pull-secret
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names evaluating values as templates
{{ include "stackstate-k8s-agent.image.pullSecrets" ( dict "images" (list .Values.path.to.the.image1, .Values.path.to.the.image2) "context" $) }}
*/}}
{{- define "stackstate-k8s-agent.image.pullSecrets" -}}
  {{- $pullSecrets := list }}
  {{- $context := .context }}
  {{- if $context.Values.global }}
    {{- range $context.Values.global.imagePullSecrets -}}
      {{/* Is plain array of strings, compatible with all bitnami charts */}}
      {{- $pullSecrets = append $pullSecrets (include "stackstate-k8s-agent.tplvalue.render" (dict "value" . "context" $context)) -}}
    {{- end -}}
  {{- end -}}
  {{- range $context.Values.imagePullSecrets -}}
    {{- $pullSecrets = append $pullSecrets (include "stackstate-k8s-agent.tplvalue.render" (dict "value" .name "context" $context)) -}}
  {{- end -}}
  {{- range .images -}}
    {{- if .pullSecretName -}}
      {{- $pullSecrets = append $pullSecrets (include "stackstate-k8s-agent.tplvalue.render" (dict "value" .pullSecretName "context" $context)) -}}
    {{- end -}}
  {{- end -}}
  {{- $pullSecrets = append $pullSecrets (include "stackstate-k8s-agent.pull-secret.name" $context)  -}}
  {{- if (not (empty $pullSecrets)) -}}
imagePullSecrets:
    {{- range $pullSecrets | uniq }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Check whether the kubernetes-state-metrics configuration is overridden. If so, return 'true' else return nothing (which is false).
{{ include "stackstate-k8s-agent.kube-state-metrics.overridden" $ }}
*/}}
{{- define "stackstate-k8s-agent.kube-state-metrics.overridden" -}}
{{- if .Values.clusterAgent.config.override }}
  {{- range $i, $val := .Values.clusterAgent.config.override }}
    {{- if and (eq $val.name "conf.yaml") (eq $val.path "/etc/stackstate-agent/conf.d/kubernetes_state.d") }}
true
    {{- end }}
  {{- end }}
{{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.nodeAgent.kube-state-metrics.overridden" -}}
{{- if .Values.nodeAgent.config.override }}
  {{- range $i, $val := .Values.nodeAgent.config.override }}
    {{- if and (eq $val.name "auto_conf.yaml") (eq $val.path "/etc/stackstate-agent/conf.d/kubernetes_state.d") }}
true
    {{- end }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Return the appropriate os label
*/}}
{{- define "label.os" -}}
{{- if semverCompare "^1.14-0" .Capabilities.KubeVersion.GitVersion -}}
kubernetes.io/os
{{- else -}}
beta.kubernetes.io/os
{{- end -}}
{{- end -}}

{{/*
Returns a YAML with extra annotations
*/}}
{{- define "stackstate-k8s-agent.global.extraAnnotations" -}}
{{- with .Values.global.extraAnnotations }}
{{- toYaml . }}
{{- end }}
{{- end -}}

{{/*
Returns a YAML with extra labels
*/}}
{{- define "stackstate-k8s-agent.global.extraLabels" -}}
{{- with .Values.global.extraLabels }}
{{- toYaml . }}
{{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.apiKeyEnv" -}}
- name: STS_API_KEY
  valueFrom:
    secretKeyRef:
{{- if .Values.stackstate.manageOwnSecrets }}
      name: {{ .Values.stackstate.customSecretName | quote }}
      key: {{ .Values.stackstate.customApiKeySecretKey | quote }}
{{- else }}
      name: {{ tpl .Values.global.apiKey.fromSecret . | quote }}
      key: STS_API_KEY
{{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.clusterAgentAuthTokenEnv" -}}
- name: STS_CLUSTER_AGENT_AUTH_TOKEN
  valueFrom:
    secretKeyRef:
{{- if .Values.stackstate.manageOwnSecrets }}
      name: {{ .Values.stackstate.customSecretName | quote }}
      key: {{ .Values.stackstate.customClusterAuthTokenSecretKey | quote }}
{{- else }}
      name: {{ tpl .Values.global.clusterAgentAuthToken.fromSecret . | quote }}
      key: STS_CLUSTER_AGENT_AUTH_TOKEN
{{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.externalOrInternal" -}}
{{- if .external }}
{{- tpl .external . }}
{{- else }}
{{- template "stackstate-k8s-agent.releasename" . }}-{{ .internalName }}
{{- end }}
{{- end }}

{{- define "stackstate-k8s-agent.secret.internal.name" -}}
{{ include "stackstate-k8s-agent.releasename" . }}-secrets
{{- end -}}

{{- define "stackstate-k8s-agent.url.configmap.internal.name" -}}
{{ include "stackstate-k8s-agent.releasename" . }}-url
{{- end -}}

{{- define "stackstate-k8s-agent.clusterName.configmap.internal.name" -}}
{{ include "stackstate-k8s-agent.releasename" . }}-cluster-name
{{- end -}}

{{- define "stackstate-k8s-agent.api-key.secret.name" -}}
{{ include "stackstate-k8s-agent.externalOrInternal" (merge (dict "external" .Values.global.receiverApiKeySecret "internalName" "api-key") .) | quote }}
{{- end }}

{{/*
Custom certificates ConfigMap name
*/}}
{{- define "stackstate-k8s-agent.customCertificates.configmap.name" -}}
{{ include "stackstate-k8s-agent.releasename" . }}-custom-certificates
{{- end -}}

{{/*
Custom certificates volume definition
*/}}
{{- define "stackstate-k8s-agent.customCertificates.volume" -}}
{{- if .Values.global.customCertificates.enabled }}
- name: custom-certificates
  configMap:
    name: {{ if .Values.global.customCertificates.configMapName }}{{ .Values.global.customCertificates.configMapName }}{{ else }}{{ include "stackstate-k8s-agent.customCertificates.configmap.name" . }}{{ end }}
{{- end }}
{{- end -}}

{{/*
Custom certificates volume mount definition
*/}}
{{- define "stackstate-k8s-agent.customCertificates.volumeMount" -}}
{{- if .Values.global.customCertificates.enabled }}
- name: custom-certificates
  mountPath: /etc/pki/tls/certs
  readOnly: true
{{- end }}
{{- end -}}

{{/*
Custom certificates ConfigMap checksum annotation
*/}}
{{- define "stackstate-k8s-agent.customCertificates.checksum" -}}
{{- if and .Values.global.customCertificates.enabled .Values.global.customCertificates.pemData (not .Values.global.customCertificates.configMapName) }}
checksum/custom-certificates: {{ include (print $.Template.BasePath "/custom-certificates-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Node agent validation - with no cluster agent to aggregate cluster-wide metadata, each node
agent falls back to reading it from the API server itself: a Node get per host-metadata cycle
and a cluster-wide Endpoints list every kubernetes_metadata_tag_update_freq seconds, per node.
The agent is meant to be switched off as a whole rather than left in that state.
*/}}
{{- define "stackstate-k8s-agent.nodeAgent.validate" -}}
{{- if not .Values.clusterAgent.enabled }}
{{- fail "nodeAgent.enabled is true but clusterAgent.enabled is false. Without a cluster agent every node agent reads cluster-wide metadata from the API server itself, which scales with node count. Disable the node agent as well, or re-enable the cluster agent." }}
{{- end }}
{{- end -}}

{{/*
Checks agent validation - the clusterchecks config provider fetches its check configurations
from the cluster agent, so the checks agent has no source of work without one.
*/}}
{{- define "stackstate-k8s-agent.checksAgent.validate" -}}
{{- if not .Values.clusterAgent.enabled }}
{{- fail "checksAgent.enabled is true but clusterAgent.enabled is false. The checks agent receives its check configurations from the cluster agent via the clusterchecks config provider, so it would run with nothing to do. Set checksAgent.enabled to false as well, or re-enable the cluster agent." }}
{{- end }}
{{- end -}}

{{/*
Custom certificates validation - fail if both configMapName and pemData are provided
*/}}
{{- define "stackstate-k8s-agent.customCertificates.validate" -}}
{{- if and .Values.global.customCertificates.enabled .Values.global.customCertificates.configMapName .Values.global.customCertificates.pemData }}
{{- fail "Error: Both global.customCertificates.configMapName and global.customCertificates.pemData are provided. Please use only one approach - either specify an external ConfigMap name OR provide PEM data directly, not both." }}
{{- end }}
{{- end -}}

{{/*
Helpers for remote kube cache service (used by process-agent pod-correlation)
*/}}
{{- define "stackstate-k8s-agent.remoteKubeCache.serviceName" -}}
{{- printf "%s-remote-kube-cache" .Release.Name -}}
{{- end -}}

{{- define "stackstate-k8s-agent.remoteKubeCache.grpcPort" -}}
50055
{{- end -}}

{{- define "stackstate-k8s-agent.remoteKubeCache.address" -}}
{{- printf "%s:%s" (include "stackstate-k8s-agent.remoteKubeCache.serviceName" .) (include "stackstate-k8s-agent.remoteKubeCache.grpcPort" .) -}}
{{- end -}}

{{- define "stackstate-k8s-agent.processAgent.podCorrelation.remoteCache.enabled" -}}
{{- if and .Values.nodeAgent.enabled .Values.nodeAgent.containers.processAgent.enabled .Values.processAgent.podCorrelation.enabled .Values.processAgent.podCorrelation.remoteCache }}
true
{{- end }}
{{- end -}}

{{/*
Determine whether the cluster collector should be deployed.
True when OTel is enabled and the collector is explicitly enabled.
*/}}
{{- define "stackstate-k8s-agent.k8sResourceCollector.enabled" -}}
{{- if and (include "stackstate-k8s-agent.otel.enabled" .) .Values.otel.k8sResourceCollector.enabled }}
true
{{- end }}
{{- end -}}

{{/*
Send Kubernetes topology through the cluster collector's cluster-agent-compatible
exporter instead of the cluster agent's topology check.
*/}}
{{- define "stackstate-k8s-agent.kubernetesTopologyCompat.enabled" -}}
{{- if and .Values.global.features.experimentalOtelKubernetesTopology (include "stackstate-k8s-agent.k8sResourceCollector.enabled" .) }}
true
{{- end }}
{{- end -}}

{{/*
Cluster-wide object watches read by the cluster-agent topology collectors, gated by
the same resource switches as the cluster agent.
*/}}
{{- define "stackstate-k8s-agent.kubernetesTopologyCompat.objects" -}}
{{- $res := .Values.clusterAgent.collection.kubernetesResources }}
{{- $objects := dict "nodes" (dict "group" "") "pods" (dict "group" "") "services" (dict "group" "") }}
{{- if $res.namespaces }}{{ $_ := set $objects "namespaces" (dict "group" "") }}{{ end }}
{{- if $res.persistentvolumes }}
{{- $_ := set $objects "persistentvolumes" (dict "group" "") }}
{{- if $res.volumeattachments }}{{ $_ := set $objects "volumeattachments" (dict "group" "storage.k8s.io") }}{{ end }}
{{- end }}
{{- if or $res.persistentvolumes $res.persistentvolumeclaims }}{{ $_ := set $objects "persistentvolumeclaims" (dict "group" "") }}{{ end }}
{{- range $name := list "deployments" "replicasets" "daemonsets" "statefulsets" }}
{{- if index $res $name }}{{ $_ := set $objects $name (dict "group" "apps") }}{{ end }}
{{- end }}
{{- range $name := list "jobs" "cronjobs" }}
{{- if index $res $name }}{{ $_ := set $objects $name (dict "group" "batch") }}{{ end }}
{{- end }}
{{- if $res.ingresses }}{{ $_ := set $objects "ingresses" (dict "group" "networking.k8s.io") }}{{ end }}
{{- range $name := list "configmaps" "secrets" }}
{{- if index $res $name }}{{ $_ := set $objects $name (dict "group" "") }}{{ end }}
{{- end }}
{{- $objects | toYaml }}
{{- end -}}

{{/*
Seconds in a duration made of h, m and s units, such as 5m or 1h30m.
*/}}
{{- define "stackstate-k8s-agent.durationSeconds" -}}
{{- $duration := toString . }}
{{- if eq $duration "0" }}{{ $duration = "0s" }}{{ end }}
{{- $parts := regexFindAll "[0-9]+[hms]" $duration -1 }}
{{- if or (not $parts) (ne (join "" $parts) $duration) }}
{{- fail (printf "unsupported duration %q: use h, m and s units, e.g. 5m" $duration) }}
{{- end }}
{{- $total := 0 }}
{{- range $parts }}
{{- $count := trimSuffix "h" . | trimSuffix "m" | trimSuffix "s" | atoi }}
{{- if hasSuffix "h" . }}{{ $total = add $total (mul $count 3600) }}
{{- else if hasSuffix "m" . }}{{ $total = add $total (mul $count 60) }}
{{- else }}{{ $total = add $total $count }}{{ end }}
{{- end }}
{{- $total }}
{{- end -}}

{{/*
How long the exporter keeps sending without a new complete observer snapshot:
three snapshot intervals, and never less than one export interval. Zero is the
receiver's five-minute default.
*/}}
{{- define "stackstate-k8s-agent.kubernetesTopologyCompat.snapshotMaxAgeSeconds" -}}
{{- $snapshot := include "stackstate-k8s-agent.durationSeconds" .Values.otel.k8sResourceCollector.crDiscovery.snapshotInterval | atoi }}
{{- if eq $snapshot 0 }}{{ $snapshot = 300 }}{{ end }}
{{- max (mul 3 $snapshot) (.Values.clusterAgent.config.topology.collectionInterval | int) }}
{{- end -}}

{{/*
Resource switches for the cluster-agent-compatible exporter. The receiver
replaces Secret data by its hash and truncates ConfigMap data on ingestion.
*/}}
{{- define "stackstate-k8s-agent.kubernetesTopologyCompat.resources" -}}
{{- $res := .Values.clusterAgent.collection.kubernetesResources }}
{{- range $name := list "persistentvolumes" "persistentvolumeclaims" "namespaces" "daemonsets" "deployments" "replicasets" "statefulsets" "ingresses" "jobs" "cronjobs" "configmaps" "secrets" }}
{{ $name }}: {{ index $res $name | default false }}
{{- end }}
{{- end -}}

{{/*
Headless Service DNS for peer-to-peer cache sync between cluster collector replicas.
*/}}
{{- define "stackstate-k8s-agent.k8sResourceCollector.peerSync.dns" -}}
{{- printf "%s-k8s-resource-collector-headless.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- end -}}

{{/*
Merge enabled integration overlays into otel.k8sResourceCollector values.

Each integrations/<name>.yaml file is loaded via .Files.Get, parsed, and deep-merged
over the base otel.k8sResourceCollector values using mustMergeOverwrite. The merge is
additive: integration API groups are added to crDiscovery.apiGroups.include
alongside any operator-supplied entries.

Returns a dict equivalent to .Values.otel.k8sResourceCollector with integration groups merged in.
*/}}
{{- define "stackstate-k8s-agent.k8sResourceCollector.mergedValues" -}}
{{- $integrations := .Values.otel.integrations | default dict }}
{{- $files := .Files }}
{{- $overlays := list }}
{{- if index $integrations "suseRuntimeEnforcer" }}
  {{- $overlays = append $overlays "integrations/suse-runtime-enforcer.yaml" }}
{{- end }}
{{- if index $integrations "suseAdmissionController" }}
  {{- $overlays = append $overlays "integrations/suse-admission-controller.yaml" }}
{{- end }}
{{- if index $integrations "suseVirtualization" }}
  {{- $overlays = append $overlays "integrations/suse-virtualization.yaml" }}
{{- end }}
{{- if index $integrations "suseSbomScanner" }}
  {{- $overlays = append $overlays "integrations/suse-sbom-scanner.yaml" }}
{{- end }}
{{- $vals := deepCopy .Values.otel.k8sResourceCollector }}
{{- range $overlays }}
  {{- $overlay := $files.Get . | fromYaml }}
  {{- $overlayVals := dig "otel" "k8sResourceCollector" dict $overlay }}
  {{- $vals = mustMergeOverwrite $vals $overlayVals }}
{{- end }}
{{- if include "stackstate-k8s-agent.kubernetesTopologyCompat.enabled" . }}
  {{- $objects := $vals.objects | default dict }}
  {{- range $name, $obj := include "stackstate-k8s-agent.kubernetesTopologyCompat.objects" . | fromYaml }}
    {{- $_ := set $objects $name $obj }}
  {{- end }}
  {{- $_ := set $vals "objects" $objects }}
{{- end }}
{{- $vals | toYaml }}
{{- end -}}

{{/*
Determine whether OTel-based features are enabled on this agent.
*/}}
{{- define "stackstate-k8s-agent.otel.enabled" -}}
{{- if .Values.otel.enabled }}
true
{{- end }}
{{- end -}}

{{/*
Determine whether OTel Prometheus scraping should be deployed.
*/}}
{{- define "stackstate-k8s-agent.otelPrometheusScraping.enabled" -}}
{{- if and (include "stackstate-k8s-agent.otel.enabled" .) .Values.otel.prometheusScraping.enabled }}
true
{{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.otelMetricsScraper.name" -}}
{{- printf "%s-otel-metrics-scraper" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelMetricsScraper.configName" -}}
{{- printf "%s-config" (include "stackstate-k8s-agent.otelMetricsScraper.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.name" -}}
{{- printf "%s-otel-target-allocator" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.configName" -}}
{{- printf "%s-config" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelMetricsScraper.selectorLabels" -}}
app.kubernetes.io/component: otel-metrics-scraper
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "stackstate-k8s-agent.name" . }}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.selectorLabels" -}}
app.kubernetes.io/component: otel-target-allocator
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "stackstate-k8s-agent.name" . }}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.serviceUrl" -}}
{{- if .Values.otel.prometheusScraping.targetAllocator.mtlsEnabled -}}
{{- printf "https://%s:8443" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) -}}
{{- else -}}
{{- printf "http://%s:8080" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) -}}
{{- end -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.selfSignedIssuerName" -}}
{{- printf "%s-selfsigned" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.caIssuerName" -}}
{{- printf "%s-ca" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.caSecretName" -}}
{{- printf "%s-ca" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTargetAllocator.tlsSecretName" -}}
{{- printf "%s-tls" (include "stackstate-k8s-agent.otelTargetAllocator.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelMetricsScraper.tlsSecretName" -}}
{{- printf "%s-tls" (include "stackstate-k8s-agent.otelMetricsScraper.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Determine whether the OTel Telemetry Gateway should be deployed.
*/}}
{{- define "stackstate-k8s-agent.otelTelemetryGateway.enabled" -}}
{{- if and (include "stackstate-k8s-agent.otel.enabled" .) .Values.otel.telemetryGateway.enabled }}
true
{{- end }}
{{- end -}}

{{/*
Determine whether the SUSE Observability Agent marker CRD should be installed.
The marker should only exist when at least one product-facing OTel integration
path is active; otel.enabled=true with all subcomponents disabled is not enough.
*/}}
{{- define "stackstate-k8s-agent.otelMarkerCrd.enabled" -}}
{{- if or (include "stackstate-k8s-agent.otelTelemetryGateway.enabled" .) (include "stackstate-k8s-agent.otelPrometheusScraping.enabled" .) }}
true
{{- end }}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTelemetryGateway.name" -}}
{{- printf "%s-otel-telemetry-gateway" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTelemetryGateway.configName" -}}
{{- printf "%s-config" (include "stackstate-k8s-agent.otelTelemetryGateway.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "stackstate-k8s-agent.otelTelemetryGateway.selectorLabels" -}}
app.kubernetes.io/component: otel-telemetry-gateway
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "stackstate-k8s-agent.name" . }}
{{- end -}}

{{/*
Target Allocator expects Kubernetes LabelSelector objects. Values accept either
a full LabelSelector or a simple label map, which is wrapped as matchLabels.
*/}}
{{- define "stackstate-k8s-agent.otelPrometheusScraping.labelSelector" -}}
{{- $selector := .selector | default dict -}}
{{- if empty $selector -}}
{}
{{- else if or (hasKey $selector "matchLabels") (hasKey $selector "matchExpressions") -}}
{{- toYaml $selector -}}
{{- else -}}
matchLabels:
{{ toYaml $selector | indent 2 }}
{{- end -}}
{{- end -}}

{{/*
Debug exporter config helper.
Given a debug config dict, returns either the empty string (when disabled) or
the debug exporter configuration block with validated verbosity.
Validates that verbosity is one of: basic, normal, detailed.
Validates that all pipeline entries are subset of: traces, logs, metrics.
*/}}
{{- define "stackstate-k8s-agent.collector.debugExporter.config" -}}
{{- if .enabled -}}
{{- $validVerbosities := list "basic" "normal" "detailed" -}}
{{- if not (has .verbosity $validVerbosities) -}}
{{- fail (printf "debug.verbosity must be one of %v, got: %s" $validVerbosities .verbosity) -}}
{{- end -}}
{{- $validSignals := list "traces" "logs" "metrics" -}}
{{- $pipelines := .pipelines | default list -}}
{{- range $pipeline := $pipelines -}}
{{- if not (has $pipeline $validSignals) -}}
{{- fail (printf "debug.pipelines must be subset of %v, got: %s" $validSignals $pipeline) -}}
{{- end -}}
{{- end -}}
debug:
  verbosity: {{ .verbosity }}
{{- end -}}
{{- end }}

{{/*
Debug exporter suffix helper.
Given a debug config dict and a signal string (traces|logs|metrics),
returns either the empty string or ", debug" to append to a pipeline's
exporters list when the signal is included in debug.pipelines.
*/}}
{{- define "stackstate-k8s-agent.collector.debugExporter.suffix" -}}
{{- if .debug.enabled }}
{{- $signal := .signal }}
{{- $pipelines := .debug.pipelines | default list }}
{{- if has $signal $pipelines }}, debug{{ end }}
{{- end }}
{{- end }}

{{/*
The container-level securityContext required by the restricted Pod Security Standard.
Not applied to the node agent, process agent or logs agent: those need host access and
can never satisfy restricted. See the required-privileges documentation.
*/}}
{{- define "stackstate-k8s-agent.container.restrictedSecurityContext" -}}
allowPrivilegeEscalation: false
capabilities:
  drop:
    - ALL
privileged: false
runAsNonRoot: true
seccompProfile:
  type: RuntimeDefault
{{- end }}

{{- define "stackstate-k8s-agent.logsAgent.memoryMiB" -}}
{{- $quantity := dig "limits" "memory" "" .Values.otelLogsAgent.resources | toString -}}
{{- if not (regexMatch "^\\+?([0-9]+(\\.[0-9]*)?|\\.[0-9]+)([EPTGMK]i|[EPTGMkmun]|[eE][+-]?[0-9]+)?$" $quantity) -}}
{{- fail "otelLogsAgent.resources.limits.memory must be a positive Kubernetes memory quantity for OTel logs" -}}
{{- end -}}
{{- $unit := regexFind "([EPTGMK]i|[EPTGMkmun])$" $quantity -}}
{{- $units := dict "" 1.0 "n" 0.000000001 "u" 0.000001 "m" 0.001 "k" 1000.0 "M" 1000000.0 "G" 1000000000.0 "T" 1000000000000.0 "P" 1000000000000000.0 "E" 1000000000000000000.0 "Ki" 1024.0 "Mi" 1048576.0 "Gi" 1073741824.0 "Ti" 1099511627776.0 "Pi" 1125899906842624.0 "Ei" 1152921504606846976.0 -}}
{{- $mib := divf (mulf (trimSuffix $unit $quantity | float64) (index $units $unit)) 1048576.0 | floor | int -}}
{{- if lt $mib 3 -}}{{- fail "otelLogsAgent.resources.limits.memory must be at least 3Mi for OTel logs" -}}{{- end -}}
{{- $mib -}}
{{- end -}}

{{- define "stackstate-k8s-agent.logsAgent.validate" -}}
{{- if not .Values.global.clusterName.fromSecret -}}
{{- if or (not .Values.stackstate.cluster.name) (contains "\"" .Values.stackstate.cluster.name) (contains "\\" .Values.stackstate.cluster.name) (regexMatch "[[:cntrl:]]" .Values.stackstate.cluster.name) -}}
{{- fail "OTel logs cluster names must not require Promtail label escaping" -}}
{{- end -}}
{{- end -}}
{{- if not .Values.global.url.fromSecret -}}
{{- $url := tpl .Values.stackstate.url . | trimAll "/" -}}
{{- $parsed := urlParse $url -}}
{{- if or (not (has $parsed.scheme (list "http" "https"))) (not $parsed.host) $parsed.userinfo $parsed.query $parsed.fragment (not (hasSuffix "/stsAgent" $parsed.path)) -}}
{{- fail "OTel logs require an HTTP(S) stackstate.url ending in /stsAgent without credentials, query or fragment" -}}
{{- end -}}
{{- end -}}
{{- if and (ne .Values.otelLogsAgent.exportMode "promtail") (include "stackstate-k8s-agent.platform.otlp.useGrpc" .) -}}
{{- $endpoint := .Values.otel.platformGrpcOtlpEndpoint -}}
{{- if not (regexMatch "^(\\[[0-9A-Fa-f:.]+(%[A-Za-z0-9_.-]+)?\\]|[A-Za-z0-9._-]+):[0-9]+$" $endpoint) -}}
{{- fail "otel.platformGrpcOtlpEndpoint must have a nonempty host and numeric port (host:port)" -}}
{{- end -}}
{{- $port := regexFind "[0-9]+$" $endpoint | int -}}
{{- if or (lt $port 1) (gt $port 65535) -}}{{- fail "otel.platformGrpcOtlpEndpoint port must be between 1 and 65535" -}}{{- end -}}
{{- if hasPrefix "https://" (.Values.global.proxy.url | lower) -}}
{{- fail "OTel logs gRPC requires an http:// CONNECT proxy; use the HTTP OTLP endpoint for an https:// proxy" -}}
{{- end -}}
{{- end -}}
{{- end -}}
