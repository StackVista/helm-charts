{{- define "stackstate.replicationChecker.enabled" -}}
{{- $enabled := .Values.stackstate.components.replicationChecker.enabled -}}
{{- if kindIs "invalid" $enabled -}}
{{- hasSuffix "-ha" (include "common.sizing.global.profile" . | trim) -}}
{{- else -}}
{{- $enabled -}}
{{- end -}}
{{- end -}}
