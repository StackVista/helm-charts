{{- define "stackstate.server.memory.resource" -}}
{{- $podMemoryLimitMB := (include "stackstate.storage.to.megabytes" .Mem) -}}
{{- $baseMemoryConsumptionMB := (include "stackstate.storage.to.megabytes" .BaseMem) -}}
{{- $grossMemoryLimit := (sub $podMemoryLimitMB  $baseMemoryConsumptionMB) | int -}}
{{- $javaHeapMemory := (div (mul $grossMemoryLimit (.JavaHeapFraction | int)) 100) | int -}}
{{- max $javaHeapMemory 0 -}}
{{- end -}}

{{/*
    Extra environment variables for pods inherited through `stackstate.components.*.extraEnv`
    We merge the service specific env vars into the common ones to avoid duplicate entries in the env
*/}}
{{- define "stackstate.service.envvars" -}}
{{- $openEnvVars := merge .ServiceConfig.extraEnv.open .Values.stackstate.components.all.extraEnv.open }}
{{- $secretEnvVars := merge .ServiceConfig.extraEnv.secret .Values.stackstate.components.all.extraEnv.secret }}
{{- $_ := set $openEnvVars "CONFIG_FORCE_stackstate_misc_releaseRevision" .Release.Revision }}
{{- $xmxConfig := dict "Mem" .ServiceConfig.resources.limits.memory "BaseMem" .ServiceConfig.sizing.baseMemoryConsumption "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xmx := (include "stackstate.server.memory.resource" $xmxConfig) | int }}
{{- $xmsConfig := dict "Mem" .ServiceConfig.resources.requests.memory "BaseMem" .ServiceConfig.sizing.baseMemoryConsumption "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xms := include "stackstate.server.memory.resource" $xmsConfig | int }}
{{- $xmxParam := ( (gt $xmx 0) | ternary (printf "-Xmx%dm" $xmx) "") }}
{{- $xmsParam := ( (gt $xms 0) | ternary (printf "-Xms%dm" $xms) "") }}
{{- if not $openEnvVars.JAVA_OPTS }}
- name: "JAVA_OPTS"
  value: {{ $xmxParam }} {{ $xmsParam }}
{{- end }}
{{- if $openEnvVars }}
  {{- range $key := (keys $openEnvVars | sortAlpha) }}
  {{- $value := get $openEnvVars $key }}
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
