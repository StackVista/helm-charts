{{- define "stackstate.server.memory.resource" -}}
{{- $podMemoryLimitMB := (include "stackstate.storage.to.megabytes" .Mem) -}}
{{- $baseMemoryConsumptionMB := (include "stackstate.storage.to.megabytes" .BaseMem) -}}
{{- $grossMemoryLimit := (sub $podMemoryLimitMB  $baseMemoryConsumptionMB) | int -}}
{{- $javaHeapMemory := (div (mul $grossMemoryLimit (.JavaHeapFraction | int)) 100) | int -}}
{{- max $javaHeapMemory 0 -}}
{{- end -}}

{{- define "stackstate.server.cache.memory.limit" -}}
{{- $podMemoryLimitMB := ( include "stackstate.storage.to.megabytes" .Mem ) -}}
{{- $podMemoryLimitBytes := (mul (mul $podMemoryLimitMB 1000) 1000) -}}
{{- $thisMemoryFactor := (sub 100 .JavaHeapFraction) -}}
{{- $cacheSize := (div (mul $podMemoryLimitBytes $thisMemoryFactor) 100) | int -}}
{{- max (sub $cacheSize .BaseMem) 0 }}
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
    We merge the service specific env vars into the common ones to avoid duplicate entries in the env
*/}}
{{- define "stackstate.service.envvars" -}}
{{- $openEnvVars := merge .ServiceConfig.extraEnv.open .Values.stackstate.components.all.extraEnv.open }}
{{- $secretEnvVars := merge .ServiceConfig.extraEnv.secret .Values.stackstate.components.all.extraEnv.secret }}
{{- if eq (lower .Values.stackstate.components.all.deploymentStrategy.type) "rollingupdate" }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_singleWriter_releaseRevision" .Release.Revision }}
{{- else }}
  {{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_singleWriter_releaseRevision" "1" }}
{{- end -}}
{{- $xmxConfig := dict "Mem" .ServiceConfig.resources.limits.memory "BaseMem" .ServiceConfig.sizing.baseMemoryConsumption "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xmx := (include "stackstate.server.memory.resource" $xmxConfig) | int }}
{{- $xmsConfig := dict "Mem" .ServiceConfig.resources.requests.memory "BaseMem" .ServiceConfig.sizing.baseMemoryConsumption "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xms := include "stackstate.server.memory.resource" $xmsConfig | int }}
{{- $xmxParam := ( (gt $xmx 0) | ternary (printf "-Xmx%dm" $xmx) "") }}
{{- $xmsParam := ( (gt $xms 0) | ternary (printf "-Xms%dm" $xms) "") }}
{{- if .Values.stackstate.java.trustStorePassword }}
- name: JAVA_TRUSTSTORE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" . }}-common
      key: javaTrustStorePassword
{{- end }}
{{- if not $openEnvVars.JAVA_OPTS }}
- name: "JAVA_OPTS"
  value: {{ $xmxParam }} {{ $xmsParam }}
{{- end }}
{{- if $openEnvVars }}
  {{- range $key, $value := $openEnvVars }}
  {{- if eq $key "JAVA_OPTS" }}
- name: {{ $key }}
  value: "{{ $xmxParam }} {{ $xmsParam }} {{ $value }}"
  {{- else }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
  {{- end }}
{{- end }}
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
stringData:
  application_stackstate.conf: |
{{- if .service.config }}
{{- .service.config | nindent 4 -}}
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
- -Dlogback.configurationFile=/opt/docker/etc_log/logback.groovy
{{- end -}}
