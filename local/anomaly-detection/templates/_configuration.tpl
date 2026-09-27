{{/*
StackState URL interface. Parents may override it to configure their bundled
workloads. Always call this helper instead of reading .Values.stackstate.instance
in consuming templates; the template guard test checks direct access.
*/}}
{{- define "anomaly-detection.stackstate.instance" -}}
{{ tpl .Values.stackstate.instance . }}
{{- end }}
