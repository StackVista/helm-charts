{{- define "common.metrics.annotations" -}}
{{- if .metrics.enabled }}
ad.stackstate.com/server.check_names: '["openmetrics"]'
ad.stackstate.com/server.init_configs: '[{}]'
ad.stackstate.com/server.instances: '[ { "prometheus_url": "http://%%host%%:{{ .port }}/metrics", "namespace": "","metrics": ["*"] } ]'
{{- end }}
{{- end -}}
