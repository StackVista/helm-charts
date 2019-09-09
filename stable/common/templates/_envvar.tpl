{{- define "common.envvar.value" -}}
  {{- $name := index . 0 -}}
  {{- $value := index . 1 -}}

  name: {{ $name }}
  value: {{ default "" $value | quote }}
{{- end -}}

{{- define "common.envvar.configmap" -}}
  {{- $name := index . 0 -}}
  {{- $configMapName := index . 1 -}}
  {{- $configMapKey := index . 2 -}}

  name: {{ $name }}
  valueFrom:
    configMapKeyRef:
      name: {{ $configMapName }}
      key: {{ $configMapKey }}
{{- end -}}

{{- define "common.envvar.secret" -}}
  {{- $name := index . 0 -}}
  {{- $secretName := index . 1 -}}
  {{- $secretKey := index . 2 -}}

  name: {{ $name }}
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ $secretKey }}
{{- end -}}

{{- define "common.extraenv.container" }}
{{- if .Values.extraEnv.open }}
  {{- range $key, $value := .Values.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.extraEnv.secret }}
  {{- range $key, $value := .Values.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname" $ }}
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end }}
