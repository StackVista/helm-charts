# stackstate-monitoring

Helm chart for Monitoring Dashboards

Current chart version is `1.0.42`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Required Values

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| grafana.dashboards.enabled | bool | `true` | Whether or not to enable the Grafana dashboards |
| rules.additionalLabels | object | `{}` | List of additional labels to add to all rules |
| rules.enabled | bool | `true` | Whether or not to enable the rules |
| rules.namespaceRegex | string | `"stackstate-.*"` | Regex to match namespaces to be monitored |
| rules.podNoCpuCheckRegex | string | `".*spotlight-worker.*"` | Regex to match namespaces to be monitored |
