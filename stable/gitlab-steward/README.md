gitlab-steward
==============
Steward -- GitLab environment cleaner

Current chart version is `0.2.1`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)



## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| backoffLimit | int | `3` | For failed jobs, how many times to retry. |
| extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| failedJobsHistoryLimit | int | `5` | The number of failed CronJob executions that are saved. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/stackstate/python-steward"` | Base container image registry. |
| image.tag | string | `"master"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"128Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"128Mi"` | Memory resource requests. |
| restartPolicy | string | `"OnFailure"` | For failed jobs, how to handle restarts. |
| schedule | string | `"*/10 * * * *"` | Default schedule for this CronJob. |
| steward.dryRun | string | `"False"` | Show which environments *would be* stopped, but don't actually stop them. |
| steward.gitlab.apiToken | string | `nil` | **REQUIRED** The GitLab API token. |
| steward.logLevel | string | `""` | The logging level of the application; one of 'debug', 'info', or 'warning' |
| steward.maxDuration | int | `2` | Amount of time (in days) before an environment is stopped. |
| steward.stackstateProject | string | `"stackstate"` |  |
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
