agent
=====
Helm chart for the StackState agent.

Current chart version is `0.3.0`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://kubernetes-charts.storage.googleapis.com/ | kube-state-metrics | 2.X.X |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:

* `stackstate.apiKey`
* `stackstate.process.agent.url`
* `stackstate.url`

Install them on the command line on Helm with the following command:

```shell
helm install \
--set stackstate.apiKey=<your-api-key> \
--set stackstate.process.agent.url=<your-process-agent-url> \
--set stackstate.url=<your-stackstate-url> \
stackstate/agent
```

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| external.kubeStateMetrics.enabled | bool | `true` | Whether or not to install the `kube-state-metrics` Deployment along with the StackState agent. Set to `false` if you have `kube-state-metrics` already installed on the cluster. |
| extraEnv | object | `{}` | Extra environment variables to be injected into the `DaemonSet` object. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| image.repository | string | `"docker.io/stackstate/stackstate-agent-2"` | Base container image registry. |
| image.tag | string | `"2.0.5"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Secrets / credentials needed for container image registry. |
| livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| livenessProbe.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the liveness probe. |
| livenessProbe.periodSeconds | int | `10` | `periodSeconds` for the liveness probe. |
| livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| livenessProbe.timeoutSeconds | int | `2` | `timeoutSeconds` for the liveness probe. |
| minReadySeconds | int | `0` | Number of seconds for which a newly created Pod should be ready without any of its containers crashing, for it to be considered available. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| readinessProbe.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the readiness probe. |
| readinessProbe.periodSeconds | int | `10` | `periodSeconds` for the readiness probe. |
| readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| readinessProbe.timeoutSeconds | int | `2` | `timeoutSeconds` for the readiness probe. |
| resources.limits.cpu | string | `"200m"` | CPU resource limits. |
| resources.limits.memory | string | `"256Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"200m"` | CPU resource requests. |
| resources.requests.memory | string | `"256Mi"` | Memory resource requests. |
| stackstate.apiKey | string | `nil` | **PROVIDE YOUR API KEY HERE** API key to be used by the StackState agent. |
| stackstate.networkTracing.enabled | bool | `true` | Whether or not to enable network tracing. |
| stackstate.process.agent.enabled | bool | `true` | Whether or not to enable the process agent. |
| stackstate.process.agent.url | string | `nil` | **PROVIDE STACKSTATE PROCESS AGENT URL HERE** URL of the StackState installation to receive data from the agent. |
| stackstate.skipSslValidation | bool | `false` | Set to true if self signed certificates are used. |
| stackstate.url | string | `nil` | **PROVIDE STACKSTATE URL HERE** URL of the StackState installation to receive data from the agent. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
| updateStrategy | object | `{}` | The update strategy for the DaemonSet object. |
