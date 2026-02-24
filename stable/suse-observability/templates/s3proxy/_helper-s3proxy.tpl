{{/*
Determine if S3Proxy should be enabled.
S3Proxy is ALWAYS enabled because the settings-local PVC needs it.
Even when global.backup.enabled=false, we need S3Proxy to serve the
settings-local-backup bucket from the local settings PVC.
*/}}
{{- define "stackstate.s3proxy.enabled" -}}
true
{{- end -}}

{{/*
Full name for S3Proxy resources.
*/}}
{{- define "stackstate.s3proxy.fullname" -}}
suse-observability-s3proxy
{{- end -}}

{{/*
S3Proxy secret name.
*/}}
{{- define "stackstate.s3proxy.secretName" -}}
{{- if .Values.backup.storage.credentials.existingSecret -}}
{{- .Values.backup.storage.credentials.existingSecret -}}
{{- else -}}
{{- include "stackstate.s3proxy.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
S3Proxy endpoint (single port 9000).
With bucket-locator middleware, all buckets are served from the same endpoint.
The middleware routes requests to the appropriate backend based on bucket name.
*/}}
{{- define "stackstate.s3proxy.endpoint" -}}
{{ include "stackstate.s3proxy.fullname" . }}:{{ include "stackstate.s3proxy.port" . }}
{{- end -}}

{{/*
S3Proxy service port. Hardcoded to 9000.
*/}}
{{- define "stackstate.s3proxy.port" -}}
9000
{{- end -}}

{{/*
Determine if the main backup backend uses PVC storage.
True when PVC backend is enabled or no other backend is explicitly enabled.
This is separate from the settings PVC which is always present.
Also checks legacy minio gateway values for backward compatibility.
*/}}
{{- define "stackstate.s3proxy.mainBackendUsesPVC" -}}
{{- $pvcEnabled := .Values.backup.storage.backend.pvc.enabled -}}
{{- $s3Enabled := .Values.backup.storage.backend.s3.enabled -}}
{{- $azureEnabled := .Values.backup.storage.backend.azure.enabled -}}
{{- $legacyS3Enabled := and .Values.minio.s3gateway .Values.minio.s3gateway.enabled -}}
{{- $legacyAzureEnabled := and .Values.minio.azuregateway .Values.minio.azuregateway.enabled -}}
{{- if or $pvcEnabled (not (or $s3Enabled $azureEnabled $legacyS3Enabled $legacyAzureEnabled)) -}}
true
{{- end -}}
{{- end -}}

{{/*
Get S3Proxy access key.
Falls back to minio.accessKey for backward compatibility, or generates random key.
*/}}
{{- define "stackstate.s3proxy.accessKey" -}}
{{- if .Values.backup.storage.credentials.accessKey -}}
{{- .Values.backup.storage.credentials.accessKey -}}
{{- else if and .Values.minio.accessKey (ne .Values.minio.accessKey "") (ne .Values.minio.accessKey "setme") -}}
{{- .Values.minio.accessKey -}}
{{- else -}}
{{- randAlphaNum 20 -}}
{{- end -}}
{{- end -}}

{{/*
Get S3Proxy secret key.
Falls back to minio.secretKey for backward compatibility, or generates random key.
*/}}
{{- define "stackstate.s3proxy.secretKey" -}}
{{- if .Values.backup.storage.credentials.secretKey -}}
{{- .Values.backup.storage.credentials.secretKey -}}
{{- else if and .Values.minio.secretKey (ne .Values.minio.secretKey "") (ne .Values.minio.secretKey "setme") -}}
{{- .Values.minio.secretKey -}}
{{- else -}}
{{- randAlphaNum 40 -}}
{{- end -}}
{{- end -}}

{{/*
Return the image registry for S3Proxy.
Uses the common.image.registry helper with the s3proxy.image configuration.
*/}}
{{- define "stackstate.s3proxy.image.registry" -}}
{{- include "common.image.registry" (dict "image" .Values.s3proxy.image "context" $) -}}
{{- end -}}

{{/*
Merge nodeSelector from stackstate.components.all and s3proxy.
s3proxy.nodeSelector takes precedence over all.nodeSelector.
*/}}
{{- define "stackstate.s3proxy.nodeSelector" -}}
{{- $allNodeSelector := .Values.stackstate.components.all.nodeSelector | default dict -}}
{{- $s3proxyNodeSelector := .Values.s3proxy.nodeSelector | default dict -}}
{{- $merged := merge $s3proxyNodeSelector $allNodeSelector -}}
{{- if $merged -}}
nodeSelector:
  {{- toYaml $merged | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
Merge affinity from stackstate.components.all and s3proxy.
s3proxy.affinity takes precedence over all.affinity.
*/}}
{{- define "stackstate.s3proxy.affinity" -}}
{{- $affinity := include "suse-observability.global.affinity" (dict "componentAffinity" .Values.s3proxy.affinity "allAffinity" .Values.stackstate.components.all.affinity "context" .) -}}
{{- if $affinity -}}
affinity:
  {{- $affinity | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
Concatenate tolerations from stackstate.components.all and s3proxy.
Both lists are combined (s3proxy tolerations are appended after all tolerations).
*/}}
{{- define "stackstate.s3proxy.tolerations" -}}
{{- $allTolerations := .Values.stackstate.components.all.tolerations | default list -}}
{{- $s3proxyTolerations := .Values.s3proxy.tolerations | default list -}}
{{- $merged := concat $allTolerations $s3proxyTolerations -}}
{{- if $merged -}}
tolerations:
  {{- toYaml $merged | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
Get the settings bucket name for local settings backup.
*/}}
{{- define "stackstate.s3proxy.localSettingsBucketName" -}}
local-settings-backup
{{- end -}}

{{/*
Check if S3 backend is enabled (either via new or legacy values).
*/}}
{{- define "stackstate.s3proxy.s3BackendEnabled" -}}
{{- if .Values.backup.storage.backend.s3.enabled -}}
true
{{- else if and .Values.minio.s3gateway .Values.minio.s3gateway.enabled -}}
true
{{- end -}}
{{- end -}}

{{/*
Check if Azure backend is enabled (either via new or legacy values).
*/}}
{{- define "stackstate.s3proxy.azureBackendEnabled" -}}
{{- if .Values.backup.storage.backend.azure.enabled -}}
true
{{- else if and .Values.minio.azuregateway .Values.minio.azuregateway.enabled -}}
true
{{- end -}}
{{- end -}}

{{/*
Get S3 backend endpoint (from new or legacy values).
*/}}
{{- define "stackstate.s3proxy.s3BackendEndpoint" -}}
{{- if .Values.backup.storage.backend.s3.endpoint -}}
{{- .Values.backup.storage.backend.s3.endpoint -}}
{{- else if and .Values.minio.s3gateway .Values.minio.s3gateway.serviceEndpoint -}}
{{- .Values.minio.s3gateway.serviceEndpoint -}}
{{- end -}}
{{- end -}}

{{/*
Get S3 backend access key (from new or legacy values).
*/}}
{{- define "stackstate.s3proxy.s3BackendAccessKey" -}}
{{- if .Values.backup.storage.backend.s3.accessKey -}}
{{- .Values.backup.storage.backend.s3.accessKey -}}
{{- else if and .Values.minio.s3gateway .Values.minio.s3gateway.accessKey -}}
{{- .Values.minio.s3gateway.accessKey -}}
{{- end -}}
{{- end -}}

{{/*
Get S3 backend secret key (from new or legacy values).
*/}}
{{- define "stackstate.s3proxy.s3BackendSecretKey" -}}
{{- if .Values.backup.storage.backend.s3.secretKey -}}
{{- .Values.backup.storage.backend.s3.secretKey -}}
{{- else if and .Values.minio.s3gateway .Values.minio.s3gateway.secretKey -}}
{{- .Values.minio.s3gateway.secretKey -}}
{{- end -}}
{{- end -}}

{{/*
Get Azure backend account name (from new or legacy values).
For legacy azuregateway, the account name was passed via minio.accessKey.
*/}}
{{- define "stackstate.s3proxy.azureBackendAccountName" -}}
{{- if .Values.backup.storage.backend.azure.accountName -}}
{{- .Values.backup.storage.backend.azure.accountName -}}
{{- else if and .Values.minio.azuregateway .Values.minio.azuregateway.enabled .Values.minio.accessKey (ne .Values.minio.accessKey "") (ne .Values.minio.accessKey "setme") -}}
{{- .Values.minio.accessKey -}}
{{- end -}}
{{- end -}}

{{/*
Get Azure backend account key (from new or legacy values).
For legacy azuregateway, the account key was passed via minio.secretKey.
*/}}
{{- define "stackstate.s3proxy.azureBackendAccountKey" -}}
{{- if .Values.backup.storage.backend.azure.accountKey -}}
{{- .Values.backup.storage.backend.azure.accountKey -}}
{{- else if and .Values.minio.azuregateway .Values.minio.azuregateway.enabled .Values.minio.secretKey (ne .Values.minio.secretKey "") (ne .Values.minio.secretKey "setme") -}}
{{- .Values.minio.secretKey -}}
{{- end -}}
{{- end -}}

{{/*
Check for deprecated minio values and return deprecation warnings as a list.
This helper is used to generate ConfigMap annotations that surface warnings to users.
*/}}
{{- define "stackstate.s3proxy.deprecationWarnings" -}}
{{- $warnings := list -}}
{{- if and .Values.minio.s3gateway .Values.minio.s3gateway.enabled -}}
{{- $warnings = append $warnings "DEPRECATED: minio.s3gateway.* values are deprecated. Use backup.storage.backend.s3.* instead." -}}
{{- end -}}
{{- if and .Values.minio.azuregateway .Values.minio.azuregateway.enabled -}}
{{- $warnings = append $warnings "DEPRECATED: minio.azuregateway.* values are deprecated. Use backup.storage.backend.azure.* instead." -}}
{{- end -}}
{{- if and .Values.minio.accessKey (ne .Values.minio.accessKey "") (ne .Values.minio.accessKey "setme") (not (and .Values.minio.azuregateway .Values.minio.azuregateway.enabled)) -}}
{{- $warnings = append $warnings "DEPRECATED: minio.accessKey is deprecated. Use backup.storage.credentials.accessKey instead." -}}
{{- end -}}
{{- if and .Values.minio.secretKey (ne .Values.minio.secretKey "") (ne .Values.minio.secretKey "setme") (not (and .Values.minio.azuregateway .Values.minio.azuregateway.enabled)) -}}
{{- $warnings = append $warnings "DEPRECATED: minio.secretKey is deprecated. Use backup.storage.credentials.secretKey instead." -}}
{{- end -}}
{{- if and .Values.minio.fullnameOverride (ne .Values.minio.fullnameOverride "") -}}
{{- $warnings = append $warnings "DEPRECATED: minio.fullnameOverride is deprecated. S3Proxy resources now use 'suse-observability-s3proxy' naming by default." -}}
{{- end -}}
{{- toJson $warnings -}}
{{- end -}}

{{/*
Check if any legacy minio values are being used.
Returns "true" if deprecated values are detected.
*/}}
{{- define "stackstate.s3proxy.hasDeprecatedValues" -}}
{{- $hasDeprecated := false -}}
{{- if and .Values.minio.s3gateway .Values.minio.s3gateway.enabled -}}
{{- $hasDeprecated = true -}}
{{- else if and .Values.minio.azuregateway .Values.minio.azuregateway.enabled -}}
{{- $hasDeprecated = true -}}
{{- else if and .Values.minio.accessKey (ne .Values.minio.accessKey "") (ne .Values.minio.accessKey "setme") -}}
{{- $hasDeprecated = true -}}
{{- else if and .Values.minio.secretKey (ne .Values.minio.secretKey "") (ne .Values.minio.secretKey "setme") -}}
{{- $hasDeprecated = true -}}
{{- end -}}
{{- if $hasDeprecated -}}
true
{{- end -}}
{{- end -}}

{{/*
Get the settings PVC name.
Uses a consistent naming scheme: backup-settings-data
*/}}
{{- define "stackstate.backup.settingsPvcName" -}}
{{- include "common.fullname.short" . -}}-backup-settings-data
{{- end -}}

{{/*
Get the main backup PVC name.
For backward compatibility, we keep the old name "suse-observability-minio" since it was used in previous versions and may already exist in user clusters.
*/}}
{{- define "stackstate.backup.mainPvcName" -}}
suse-observability-minio
{{- end -}}

{{/*
Calculate JVM heap parameters (Xms, Xmx) for S3Proxy based on resource configuration.
Uses the common stackstate.jvm.heapParams helper.
Returns the JAVA_OPTS string with calculated -Xms and -Xmx values.
*/}}
{{- define "stackstate.s3proxy.javaOpts" -}}
{{- include "stackstate.jvm.heapParams" (dict "MemoryLimit" .Values.s3proxy.resources.limits.memory "MemoryRequest" .Values.s3proxy.resources.requests.memory "BaseMemoryConsumption" .Values.s3proxy.sizing.baseMemoryConsumption "JavaHeapMemoryFraction" .Values.s3proxy.sizing.javaHeapMemoryFraction) -}}
{{- end -}}
