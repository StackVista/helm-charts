{{- define "common.metrics.annotations" -}}
{{- if .root.Values.all.metrics.enabled }}
annotations:
  ad.stackstate.com/server.check_names: '["openmetrics"]'
  ad.stackstate.com/server.init_configs: '[{}]'
  ad.stackstate.com/server.instances: '[ { "prometheus_url": "http://%%host%%:{{ .port }}/metrics", "namespace": "","metrics": ["*"] } ]'
{{- end }}
{{- end -}}
