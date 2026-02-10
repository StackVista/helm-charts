{{- define "stackstate.server.memory.resource" -}}
{{- $podMemoryLimitMB := (include "stackstate.storage.to.megabytes" .Mem) -}}
{{- $grossMemoryLimit := (subf $podMemoryLimitMB .BaseMemoryConsumptionMB) | int -}}
{{- $javaHeapMemory := (divf (mulf $grossMemoryLimit (.JavaHeapFraction | int)) 100) | int -}}
{{- max $javaHeapMemory 0 -}}
{{- end -}}

{{- define "stackstate.server.cache.memory.limit" -}}
{{- $podMemoryLimitMB := ( include "stackstate.storage.to.megabytes" .Mem ) -}}
{{- $podMemoryLimitBytes := (mulf (mulf $podMemoryLimitMB 1000) 1000) -}}
{{- $thisMemoryFactor := (sub 100 .JavaHeapFraction) -}}
{{- $cacheSize := (divf (mulf $podMemoryLimitBytes $thisMemoryFactor) 100) | int -}}
{{- max (sub $cacheSize .BaseMem) 0 }}
{{- end -}}

{{- define "stackstate.server.image.tag" -}}
{{- $hbaseVersionImageSuffix := ternary "-2.5" "" (eq .Values.hbase.version "2.5") -}}
{{- $image := default .Values.stackstate.components.all.image.tag .ImageTag  -}}
{{- $image -}}{{- $hbaseVersionImageSuffix -}}
{{- end -}}

{{/*
    Include the parallelism env var which is derived from the CPU requests on the deployment
    Receives the CPU requests string as argument and uses it to calc the effective parallelism on the sync service.
    The parallelism is the requested CPU divided by 2 (parallel component and relation workers) or 1
*/}}
{{- define "stackstate.server.cpu.parallelism" }}
{{- $podCpuCoreRequest := (include "stackstate.cpu_resource.to.cpu_core" .) }}
- name: CONFIG_FORCE_stackstate_sync_parallelWorkers
  value: {{ max $podCpuCoreRequest 1 | quote }}
{{- end }}

{{/*
    Extra environment variables for pods inherited through `stackstate.components.*.extraEnv`
    We merge the service specific env vars into the common ones to avoid duplicate entries in the env.

    Env var precedence (highest to lowest):
    1. Component-specific extraEnv.open (e.g., stackstate.components.api.extraEnv.open)
    2. Common extraEnv.open (stackstate.components.all.extraEnv.open)
    3. Sizing profile env vars
    4. Deployment-specific defaults (passed via .DeploymentEnv)
    5. Chart defaults (hardcoded values in this helper)

    This allows users to override any pre-defined env var via extraEnv.

    Usage:
    {{- $deploymentEnv := dict "KAFKA_BROKERS" (include "stackstate.kafka.endpoint" .) "ELASTICSEARCH_URI" (printf "http://%s" (include "stackstate.es.endpoint" .)) }}
    {{- include "stackstate.service.envvars" (merge (dict "DeploymentEnv" $deploymentEnv) . $serviceConfig) }}
*/}}
{{- define "stackstate.service.envvars" -}}
{{/*
Memory used by a JVM process can be calculated as follows:
JVM memory = Heap memory+ Metaspace + CodeCache + (ThreadStackSize * Number of Threads) + DirectByteBuffers + Jvm-native.
'BaseMemoryConsumption' encapsulates OS-level memory requirements and following parts of JVM memory: Metaspace, CodeCache,
ThreadStackSize * Number of Threads and Jvm-native.
'JavaHeapMemoryFraction' percent of the remainder of 'BaseMemoryConsumption' subtracted from pod's memory limit is given to 'Xmx'.
'JavaHeapMemoryFraction' percent of the remainder of 'BaseMemoryConsumption' subtracted from pod's memory request is given to 'Xms'.
Remainder of 'BaseMemoryConsumption' and 'Xmx' subtracted from pod's memory limit is given to 'DirectMemory'.
Sum of 'BaseMemoryConsumption', 'Xmx' and 'DirectMemory' totals to pod's memory limit.
*/}}
{{- $baseMemoryConsumptionMB := (include "stackstate.storage.to.megabytes" .ServiceConfig.sizing.baseMemoryConsumption)}}
{{- $xmxConfig := dict "Mem" .ServiceConfig.resources.limits.memory "BaseMemoryConsumptionMB" $baseMemoryConsumptionMB "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xmx := (include "stackstate.server.memory.resource" $xmxConfig) | int }}
{{- $xmsConfig := dict "Mem" .ServiceConfig.resources.requests.memory "BaseMemoryConsumptionMB" $baseMemoryConsumptionMB "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xms := include "stackstate.server.memory.resource" $xmsConfig | int }}
{{- $xmxParam := ( (gt $xmx 0) | ternary (printf "-Xmx%dm" $xmx) "") }}
{{- $xmsParam := ( (gt $xms 0) | ternary (printf "-Xms%dm" $xms) "") }}
{{- $directMem := (subf (subf (include "stackstate.storage.to.megabytes" .ServiceConfig.resources.limits.memory) $baseMemoryConsumptionMB) $xmx) | int }}
{{- $directMemParam := ( (gt $directMem 0) | ternary ( printf "-XX:MaxDirectMemorySize=%dm" $directMem) "") }}
{{- $otelInstrumentationServiceConfig := .ServiceConfig.otelInstrumentation | default dict }}
{{- $otelJavaAgentOpt := or .Values.stackstate.components.all.otelInstrumentation.enabled $otelInstrumentationServiceConfig.enabled | default false | ternary " -javaagent:/opt/docker/opentelemetry-javaagent.jar" "" }}
{{- $defaultJavaOpts := printf "%s %s %s%s" $directMemParam $xmxParam $xmsParam $otelJavaAgentOpt | trim }}

{{/* Start with chart defaults - these can be overridden by user extraEnv */}}
{{- $openEnvVars := dict }}
{{- $_ := set $openEnvVars "STACKSTATE_EDITION" .Values.stackstate.deployment.edition }}
{{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_featureSwitches_instanceDebugApi" (.Values.stackstate.instanceDebugApi.enabled | toString) }}
{{- $_ := set $openEnvVars "PLATFORM_VERSION" .Chart.Version }}
{{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_newMetrics" "true" }}
{{- if eq (lower .Values.stackstate.components.all.deploymentStrategy.type) "rollingupdate" }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_singleWriter_releaseRevision" (.Release.Revision | toString) }}
{{- else }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_singleWriter_releaseRevision" "1" }}
{{- end -}}
{{- if include "suse-observability.features.enabled" (dict "key" "traces" "context" .) }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces" "true" }}
{{- end -}}
{{- if .Values.global.features.experimentalStackpacks }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2" "true" }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_featureSwitches_enableTopologyStreamSync" "true" }}
{{- end -}}
{{- if .Values.stackstate.topology.retentionHours }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackgraph_retentionWindowMs" (mul (.Values.stackstate.topology.retentionHours | int) (mul 60 (mul 60 1000)) | toString) }}
{{- end }}
{{/* OTEL env vars that have simple string values */}}
{{- if or .Values.stackstate.components.all.otelInstrumentation.enabled $otelInstrumentationServiceConfig.enabled }}
  {{- $_ := set $openEnvVars "OTEL_EXPORTER_OTLP_ENDPOINT" (index $otelInstrumentationServiceConfig "otlpExporterEndpoint" | default .Values.stackstate.components.all.otelInstrumentation.otlpExporterEndpoint) }}
  {{- $_ := set $openEnvVars "OTEL_EXPORTER_OTLP_PROTOCOL" (index $otelInstrumentationServiceConfig "otlpExporterProtocol" | default .Values.stackstate.components.all.otelInstrumentation.otlpExporterProtocol) }}
  {{- $_ := set $openEnvVars "OTEL_SERVICE_NAME" "stackstate-$(STS_SERVICE_NAME)" }}
  {{- $_ := set $openEnvVars "OTEL_RESOURCE_ATTRIBUTES" (printf "service.namespace=%s,service.instance.id=$(POD_NAME)" (tpl (default .Values.stackstate.components.all.otelInstrumentation.serviceNamespace $otelInstrumentationServiceConfig.serviceNamespace) .)) }}
{{- end }}

{{/* Merge deployment-specific env vars (passed via .DeploymentEnv, lower priority than user extraEnv) */}}
{{/* IMPORTANT: Use merge (dict) as first arg to avoid mutating .DeploymentEnv */}}
{{- if .DeploymentEnv }}
  {{- $openEnvVars = merge (dict) .DeploymentEnv $openEnvVars }}
{{- end }}

{{/* Merge sizing profile env vars (lower priority than user extraEnv) */}}
{{/* IMPORTANT: Use merge (dict) as first arg to avoid mutating profile dict */}}
{{- $profileEnv := include "common.sizing.stackstate.all.extraEnv.open" . | trim -}}
{{- if $profileEnv }}
  {{- $profileEnvDict := fromYaml $profileEnv }}
  {{- $openEnvVars = merge (dict) $profileEnvDict $openEnvVars }}
{{- end }}

{{/* Merge common extraEnv.open (user values override profile and defaults) */}}
{{/* IMPORTANT: Use merge (dict) as first arg to avoid mutating .Values */}}
{{- $openEnvVars = merge (dict) .Values.stackstate.components.all.extraEnv.open $openEnvVars }}

{{/* Merge component-specific extraEnv.open (highest priority) */}}
{{/* IMPORTANT: Use merge (dict) as first arg to avoid mutating .ServiceConfig */}}
{{- $openEnvVars = merge (dict) .ServiceConfig.extraEnv.open $openEnvVars }}

{{/* Merge secret env vars */}}
{{- $secretEnvVars := merge (dict) .ServiceConfig.extraEnv.secret .Values.stackstate.components.all.extraEnv.secret }}

{{/* Handle JAVA_OPTS specially - append user value to computed JVM params if provided */}}
{{- $userJavaOpts := index $openEnvVars "JAVA_OPTS" }}
{{- if $userJavaOpts }}
  {{- $_ := set $openEnvVars "JAVA_OPTS" (printf "%s %s" $defaultJavaOpts $userJavaOpts | trim) }}
{{- else }}
  {{- $_ := set $openEnvVars "JAVA_OPTS" $defaultJavaOpts }}
{{- end }}

{{/* Output env vars that use valueFrom (cannot be merged via dict) */}}
{{- if .Values.stackstate.java.trustStorePassword }}
- name: JAVA_TRUSTSTORE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" . }}-common
      key: javaTrustStorePassword
{{- end }}
{{- if or .Values.stackstate.components.all.otelInstrumentation.enabled $otelInstrumentationServiceConfig.enabled }}
- name: "STS_SERVICE_NAME"
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: metadata.labels['app.kubernetes.io/component']
- name: "POD_NAME"
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: metadata.name
{{- end }}
{{/* Output all merged open env vars */}}
{{- range $key, $value := $openEnvVars }}
- name: {{ $key }}
  value: {{ $value | quote }}
{{- end }}
{{/* Output secret env vars */}}
{{- if $secretEnvVars }}
  {{- range $key :=  (keys $secretEnvVars | sortAlpha)  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-{{ $.ServiceName }}
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Helper to output PLATFORM_VERSION env var.
Note: For deployments using stackstate.service.envvars, this is already included via merge.
This helper is kept for backward compatibility with other templates (e.g., backup jobs).
*/}}
{{- define "stackstate.env.platform_version" }}
- name: PLATFORM_VERSION
  value: "{{- .Chart.Version -}}"
{{- end -}}

{{- define "stackstate.service.secret.data" -}}
{{- $secrets := .service.extraEnv.secret -}}
{{- $additionalSecrets := default dict .additionalSecrets -}}
{{- $context := .context -}}
{{- $allSecrets := merge $secrets $context.Values.stackstate.components.all.extraEnv.secret -}}
data:
{{- range $key, $value := $allSecrets }}
  {{ $key }}: {{ $value | b64enc | quote }}
{{- end }}
{{- range $key, $value := $additionalSecrets }}
  {{ $key }}: {{ $value | b64enc | quote }}
{{- end }}
{{- end -}}

{{- define "stackstate.service.configmap.data" -}}
data:
  application_stackstate.conf: |
{{- if .service.config }}
{{- .service.config | nindent 4 -}}
{{- end }}
{{- end -}}

{{- define "stackstate.service.configmap.clickhouseconfig" -}}
{{- if include "suse-observability.features.enabled" (dict "key" "traces" "context" .) }}
stackstate.traces.clickHouse = {{- .Values.stackstate.components.all.clickHouse | toPrettyJson }}
{{- end }}
{{- end -}}

{{/*
Secrets dict for custom certificates for stackstate services
*/}}
{{- define "stackstate.service.mountsecrets" -}}
{{- $mountSecrets := dict }}
{{- if hasKey .Values.stackstate.authentication.ldap "ssl" }}
  {{- if or .Values.stackstate.authentication.ldap.ssl.trustCertificates .Values.stackstate.authentication.ldap.ssl.trustCertificatesBase64Encoded }}
    {{- $_ := set $mountSecrets "ldapTrustCertificates" "ldap-certificates.pem" }}
  {{- end }}
  {{- if or .Values.stackstate.authentication.ldap.ssl.trustStore .Values.stackstate.authentication.ldap.ssl.trustStoreBase64Encoded }}
    {{- $_ := set $mountSecrets "ldapTrustStore" "ldap-cacerts" }}
  {{- end }}
{{- end }}
{{- if or .Values.stackstate.java.trustStore .Values.stackstate.java.trustStoreBase64Encoded }}
    {{- $_ := set $mountSecrets "javaTrustStore" "java-cacerts" }}
{{- end }}
{{ $mountSecrets | toYaml }}
{{- end -}}

{{/*
Mount secrets and log config in pod for stackstate services
*/}}
{{- define "stackstate.service.pod.volumes" -}}
{{- $mountSecrets := fromYaml (include "stackstate.service.mountsecrets" .root ) }}
- name: config-volume-log
  configMap:
    name: {{ template "common.fullname.short" .root }}-{{ .pod_name }}-log
- name: config-volume
  configMap:
    name: {{ template "common.fullname.short" .root }}-{{ .pod_name }}
{{- if $mountSecrets }}
- name: service-secrets-volume
  secret:
    secretName: {{ template "common.fullname.short" .root }}-common
    items:
{{- range $key, $val := $mountSecrets }}
    - key: {{ $key }}
      path: {{ $val }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Mount secrets and log config in container for for stackstate services
*/}}
{{- define "stackstate.service.container.volumes" -}}
{{- $mountSecrets := fromYaml (include "stackstate.service.mountsecrets" .) }}
- name: config-volume-log
  mountPath: /opt/docker/etc_log
- name: config-volume
  mountPath: /opt/docker/etc/application_stackstate.conf
  subPath: application_stackstate.conf
{{- if $mountSecrets }}
- name: service-secrets-volume
  mountPath: /opt/docker/secrets
{{- end }}
{{- end -}}

{{/*
Arguments applicable for all StackState services
NOTE: $JAVA_TRUSTSTORE_PASSWORD cannot be passed via JAVA_OPTS because the startup script
does not expand variables within the JAVA_OPTS anymore
*/}}
{{- define "stackstate.service.args" -}}
{{- if or .Values.stackstate.java.trustStore .Values.stackstate.java.trustStoreBase64Encoded }}
- -Djavax.net.ssl.trustStore=/opt/docker/secrets/java-cacerts
- -Djavax.net.ssl.trustStoreType=jks
{{- if .Values.stackstate.java.trustStorePassword }}
- -Djavax.net.ssl.trustStorePassword=$(JAVA_TRUSTSTORE_PASSWORD)
{{- end }}
{{- end }}
- -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml
{{- end -}}

{{/*
Volume for transaction logs
*/}}
{{- define "stackstate.service.transactionLog.volume" -}}
- name: application-log
  persistentVolumeClaim:
    claimName: {{ template "common.fullname.short" .root }}-{{ .pod_name }}-txlog
{{- end -}}

{{/*
VolumeMount for transaction logs
*/}}
{{- define "stackstate.service.transactionLog.volumeMount" -}}
- name: application-log
  mountPath: /opt/docker/logs
{{- end -}}

{{/*
Volume for stackpacks logs
*/}}
{{- define "stackstate.stackpacks.volume" -}}
- name: stackpacks
  persistentVolumeClaim:
    claimName: {{ template "common.fullname.short" . }}-stackpacks
{{- end -}}

{{/*
VolumeMount for stackpacks
*/}}
{{- define "stackstate.stackpacks.volumeMount" -}}
- name: stackpacks
  mountPath: /var/stackpacks
{{- end -}}


{{/*
Volume for stackpacks local
*/}}
{{- define "stackstate.stackpacks.local.volume" -}}
- name: stackpacks-local
  persistentVolumeClaim:
    claimName: {{ template "common.fullname.short" . }}-stackpacks-local
{{- end -}}

{{/*
VolumeMount for stackpacks local
*/}}
{{- define "stackstate.stackpacks.local.volumeMount" -}}
- name: stackpacks-local
  mountPath: /var/stackpacks_local
{{- end -}}

{{/*
yaml vars
*/}}
{{- define "stackstate.server.yaml.maxSizeLimit" -}}
{{- $maxSizeYamlMB := (include "stackstate.storage.to.megabytes" .Values.stackstate.components.api.yaml.maxSizeLimit) -}}
{{- if eq $maxSizeYamlMB .Values.stackstate.components.api.yaml.maxSizeLimit -}}
{{- .Values.stackstate.components.api.yaml.maxSizeLimit -}}
{{- else -}}
{{- mulf $maxSizeYamlMB 1000000  | int -}}
{{- end -}}
{{- end -}}

{{- define "stackstate.backup.yaml.maxSizeLimit" -}}
{{- $maxSizeYamlMB := (include "stackstate.storage.to.megabytes" .Values.backup.configuration.yaml.maxSizeLimit) -}}
{{- if eq $maxSizeYamlMB .Values.backup.configuration.yaml.maxSizeLimit -}}
{{- .Values.backup.configuration.yaml.maxSizeLimit -}}
{{- else -}}
{{- mulf $maxSizeYamlMB 1000000  | int -}}
{{- end -}}

{{- end -}}

{{- define "stackstate.config.email" -}}
{{- if .Values.stackstate.email.enabled }}
stackstate{
  email {
    properties = {{ .Values.stackstate.email.additionalProperties | toJson }}
    sender = {{ .Values.stackstate.email.sender | quote }}
    server {
      protocol = {{ .Values.stackstate.email.server.protocol | quote }}
      host = {{ .Values.stackstate.email.server.host | quote }}
      port = {{ .Values.stackstate.email.server.port | quote }}
      username = ${SMTP_USER_NAME}
      password = ${SMTP_PASSWORD}
    }
  }
}
{{- end -}}
{{- end -}}
