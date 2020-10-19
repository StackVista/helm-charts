{{- define "stackstate.server.memory.resource" -}}
{{- $podMemoryLimitMB := (include "stackstate.storage.to.megabytes" .Mem) -}}
{{- $baseMemoryConsumptionMB := (include "stackstate.storage.to.megabytes" .BaseMem) -}}
{{- $grossMemoryLimit := (sub $podMemoryLimitMB  $baseMemoryConsumptionMB) | int -}}
{{- $javaHeapMemory := (div (mul $grossMemoryLimit (.JavaHeapFraction | int)) 100) | int -}}
{{- max $javaHeapMemory 0 -}}
{{- end -}}

{{/*
    Extra environment variables for pods inherited through `stackstate.components.*.extraEnv`
*/}}
{{- define "stackstate.server.based.envvars" -}}
{{- $xmxConfig := dict "Mem" .ServiceConfig.resources.limits.memory "BaseMem" .ServiceConfig.sizing.baseMemoryConsumption "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xmx := (include "stackstate.server.memory.resource" $xmxConfig) | int }}
{{- $xmsConfig := dict "Mem" .ServiceConfig.resources.requests.memory "BaseMem" .ServiceConfig.sizing.baseMemoryConsumption "JavaHeapFraction" .ServiceConfig.sizing.javaHeapMemoryFraction }}
{{- $xms := include "stackstate.server.memory.resource" $xmsConfig | int }}
{{- $xmxParam := ( (gt $xmx 0) | ternary (printf "-Xmx %dm" $xmx) "") }}
{{- $xmsParam := ( (gt $xms 0) | ternary (printf "-Xms %dm" $xms) "") }}
{{- if not .ServiceConfig.extraEnv.open.JAVA_OPTS }}
{{- if not .Values.stackstate.components.all.extraEnv.open.JAVA_OPTS }}
- name: "JAVA_OPTS"
  value: {{ $xmxParam }} {{ $xmsParam }}
{{- else }}
- name: "JAVA_OPTS"
  value: {{ $xmxParam }} {{ $xmsParam }} {{ .Values.stackstate.components.all.extraEnv.open.JAVA_OPTS }}
{{- end }}
{{- end }}
{{- if .ServiceConfig.extraEnv.open }}
  {{- range $key, $value := .ServiceConfig.extraEnv.open  }}
  {{- if eq $key "JAVA_OPTS" }}
- name: {{ $key }}
  value: "{{ $xmxParam }} {{ $xmsParam }} {{ $value }}"
  {{- else }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
  {{- end }}
{{- end }}
{{- if .ServiceConfig.extraEnv.secret }}
  {{- range $key, $value := .ServiceConfig.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-{{ $.ServiceName }}
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}
