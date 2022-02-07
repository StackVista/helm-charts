# helm-reaper

The StackState Helm Reaper

Current chart version is `1.0.0`

**Homepage:** <https://gitlab.com/StackVista/devops/helm-charts>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.stackstate.io/ | common | 0.4.19 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the Helm reaper cron job |
| image.registry | string | `"quay.io"` | Registry containing the image for the Helm reaper cron job |
| image.repository | string | `"stackstate/sts-ci-images"` | Repository containing the image for the Helm reaper cron job |
| image.tag | string | `"stackstate-devops-0925b2ff"` | Tag of the image for the Helm reaper cron job |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| reapNamespace | string | `"default"` | The namespace to reap Helm releases from. |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"64Mi"` |  |
| resources.requests.cpu | string | `"25m"` |  |
| resources.requests.memory | string | `"32Mi"` |  |
| schedule | string | `"0 */1 * * *"` | The cron schedule for the Helm reaper cron job. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
