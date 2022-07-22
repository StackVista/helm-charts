# gitlab-steward

![Version: 0.5.1](https://img.shields.io/badge/Version-0.5.1-informational?style=flat-square) ![AppVersion: 0.5.0](https://img.shields.io/badge/AppVersion-0.5.0-informational?style=flat-square)
Steward -- GitLab environment cleaner
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Jeroen van Erp | <jvanerp@stackstate.com> |  |
| Remco Beckers | <rbeckers@stackstate.com> |  |
| Vincent Partington | <vpartington@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| backoffLimit | int | `3` | For failed jobs, how many times to retry. |
| extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| failedJobsHistoryLimit | int | `5` | The number of failed CronJob executions that are saved. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.pullSecretUsername | string | `nil` | Specify username and password to create an image pull secret that is used to pull the imagepullSecretUsername: |
| image.pullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. Is ignored when specifying a pullSecretUsername/password |
| image.pullsecretPassword | string | `nil` |  |
| image.repository | string | `"quay.io/stackstate/gitlab-steward"` | Base container image registry. |
| image.tag | string | `"v0.5.0"` | Default container image tag. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"128Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"128Mi"` | Memory resource requests. |
| restartPolicy | string | `"OnFailure"` | For failed jobs, how to handle restarts. |
| schedule | string | `"*/10 * * * *"` | Default schedule for this CronJob. |
| serviceAccount.annotations | object | `{}` | Extra annotations for the `ServiceAccount` object. |
| serviceAccount.create | bool | `true` | Create the `ServiceAccount` object. |
| steward.dryRun | string | `"False"` | Show which environments *would be* stopped, but don't actually stop them. |
| steward.gitlab.apiToken | string | `nil` | **REQUIRED** The GitLab API token. |
| steward.logLevel | string | `""` | The logging level of the application; one of 'debug', 'info', or 'warning' |
| steward.maxDuration | int | `2` | Amount of time (in days) before an environment is stopped. |
| steward.stackstateProject | string | `"stackstate"` |  |
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
