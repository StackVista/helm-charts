{{- define "stackstate.server.memory.resource" -}}
{{- $baseMemoryConsumption := "380Mi" -}}
{{- $javaHeapMemoryFraction := 75 -}}
{{- $podMemoryLimitMB := (include "stackstate.storage.to.megabytes" .) -}}
{{- $baseMemoryConsumptionMB := (include "stackstate.storage.to.megabytes" $baseMemoryConsumption) -}}
{{- $javaHeapMemory := (sub $podMemoryLimitMB  $baseMemoryConsumptionMB) | int -}}
{{- $jjj := (div (mul $javaHeapMemory $javaHeapMemoryFraction) 100) | int -}}
{{- max $jjj 0 -}}
{{- end -}}

{{/*
    Extra environment variables for pods inherited through `stackstate.components.*.extraEnv`
*/}}
{{- define "stackstate.server.based.envvars" -}}
{{- $xmx := (include "stackstate.server.memory.resource" .ServiceConfig.resources.limits.memory) | int }}
{{- $xms := include "stackstate.server.memory.resource" .ServiceConfig.resources.requests.memory | int }}
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
