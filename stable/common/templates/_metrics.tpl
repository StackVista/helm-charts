{{- define "common.metrics.annotations" -}}
{{- if .metrics.enabled }}
ad.stackstate.com/{{ .container_name }}.check_names: '["openmetrics"]'
ad.stackstate.com/{{ .container_name }}.init_configs: '[{}]'
ad.stackstate.com/{{ .container_name }}.instances: '[ { "prometheus_url": "http://%%host%%:{{ .port }}/metrics", "namespace": "stackstate", "metrics": ["*"] } ]'
{{- if .cluster_name }}
ad.stackstate.com/tags: '{ "kube_cluster_name": "{{ .cluster_name }}" }'
{{- end -}}
{{- end }}
{{- end -}}
