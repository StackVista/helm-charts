# helm-reaper

The StackState Helm Reaper

Current chart version is `2.0.3`

**Homepage:** <https://gitlab.com/StackVista/devops/helm-charts>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common/ | common | * |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| deleteOlderThan | int | `3` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the Helm reaper cron job |
| image.registry | string | `"quay.io"` | Registry containing the image for the Helm reaper cron job |
| image.repository | string | `"stackstate/sts-ci-images"` | Repository containing the image for the Helm reaper cron job |
| image.tag | string | `"stackstate-devops-0925b2ff"` | Tag of the image for the Helm reaper cron job |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| reapNamespaceExcludeLabel | string | `"helm-reaper.stackstate.io/exclude=true"` | The namespaces having this label will be ignored |
| reapNamespaceLabels | string | `"saas.stackstate.io/performance-test=true,saas.stackstate.io/branch-deployment=true"` | The comma-separated list of the labels to filter namespaces to reap |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"64Mi"` |  |
| resources.requests.cpu | string | `"25m"` |  |
| resources.requests.memory | string | `"32Mi"` |  |
| schedule | string | `"0 */1 * * *"` | The cron schedule for the Helm reaper cron job. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
