{{- define "common.metrics.annotations" -}}
{{- if and .metrics.enabled .metrics.agentAnnotationsEnabled }}
ad.stackstate.com/{{ .container_name }}.check_names: '["openmetrics"]'
ad.stackstate.com/{{ .container_name }}.init_configs: '[{}]'
ad.stackstate.com/{{ .container_name }}.instances: '[ { "prometheus_url": "http://%%host%%:{{ .port }}/metrics", "namespace": "stackstate", "metrics": {{ (.filter | default .metrics.defaultAgentMetricsFilter) | default "[\"*\"]" }} } ]'
{{- end }}
{{- end -}}
