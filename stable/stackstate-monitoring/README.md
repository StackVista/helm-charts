# stackstate-monitoring

Helm chart for Monitoring Dashboards

Current chart version is `1.1.13`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Required Values

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| grafana.dashboards.enabled | bool | `true` | Whether or not to enable the Grafana dashboards |
| grafana.dashboards.labels | object | `{"grafana_dashboard":1}` | Automatically load dashboards with these labels |
| rules.additionalLabels | object | `{}` | List of additional labels to add to all rules |
| rules.enabled | bool | `true` | Whether or not to enable the rules |
| rules.namespaceRegex | string | `"stackstate-.*"` | Regex to match namespaces to be monitored |
| rules.podNoCpuCheckRegex | string | `".*spotlight-worker.*"` | Regex to match namespaces to be monitored |
