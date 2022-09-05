# petros-d-kubelet-stats-exporter

![Version: 0.9.4](https://img.shields.io/badge/Version-0.9.4-informational?style=flat-square) ![AppVersion: 0.9.2](https://img.shields.io/badge/AppVersion-0.9.2-informational?style=flat-square)

The kubelet stats exporter for ephemeral storage metrics

Current chart version is `0.9.4`

**Homepage:** <https://gitlab.com/stackvista/devops/docker-petros-d-kubelet-stats-exporter>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | ops@stackstate.com |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| dashboards.enabled | bool | `true` | Enables Dashboard resource for prometheus-operator |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the Kommoner operator |
| image.registry | string | `"quay.io"` | Registry containing the image for the Kommoner operator |
| image.repository | string | `"stackstate/petros-d-kubelet-stats-exporter"` | Repository containing the image for the Kommoner operator |
| image.tag | string | `"0.9.2"` | Tag of the image for the Kommoner operator |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"250Mi"` |  |
| resources.requests.cpu | string | `"10m"` |  |
| resources.requests.memory | string | `"50Mi"` |  |
| serviceMonitor.enabled | bool | `true` | Enables ServiceMonitor resource for prometheus-operator |
| serviceMonitor.interval | string | `"300s"` |  |
| serviceMonitor.scrapeTimeout | string | `"120s"` | Scrape timeout for the ServiceMonitor |
