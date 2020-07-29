cluster-agent
=============
Helm chart for the StackState cluster agent.

Current chart version is `0.4.2`

Source code can be found [here](https://github.com/StackVista/stackstate-agent)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://kubernetes-charts.storage.googleapis.com/ | kube-state-metrics | 2.X.X |

If you already have the `kube-state-metrics` application installed in your Kubernetes cluster, set `dependencies.kubeStateMetrics.enabled` to `false` to disable installation via this Helm chart.

## Required Values

In order to successfully install this chart, you **must** provide the following variables:

* `stackstate.apiKey`
* `stackstate.cluster.name`
* `stackstate.url`

The parameter `stackstate.cluster.name` is entered when installing the Cluster Agent StackPack.

Install them on the command line on Helm with the following command:

```shell
helm install \
--set-string 'stackstate.apiKey'='<your-api-key>' \
--set-string 'stackstate.cluster.name'='<your-cluster-name>' \
--set-string 'stackstate.url'='<your-stackstate-url>' \
stackstate/cluster-agent
```

## Recommended Values

It is also recommended that you set a value for `stackstate.cluster.authToken`. If it is not provided, a value will be generated for you, but the value will change each time an upgrade is performed.

The command for **also** installing with a set token would be:

```shell
helm install \
--set-string 'stackstate.apiKey'='<your-api-key>' \
--set-string 'stackstate.cluster.name'='<your-cluster-name>' \
--set-string 'stackstate.cluster.authToken'='<your-cluster-token>' \
--set-string 'stackstate.url'='<your-stackstate-url>' \
stackstate/cluster-agent
```

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agent.affinity | object | `{}` | Affinity settings for pod assignment. |
| agent.apm.enabled | bool | `true` | Enable / disable the agent APM module. |
| agent.checksTagCardinality | string | `"orchestrator"` | low, orchestrator or high. Orchestrator level adds pod_name, high adds display_container_name |
| agent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| agent.image.repository | string | `"docker.io/stackstate/stackstate-agent-2"` | Base container image registry. |
| agent.image.tag | string | `"2.7.0"` | Default container image tag. |
| agent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| agent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| agent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| agent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| agent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| agent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| agent.logLevel | string | `"DEBUG"` | Logging level for agent processes. |
| agent.networkTracing.enabled | bool | `true` | Enable / disable the agent network tracing module. |
| agent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| agent.processAgent.enabled | bool | `true` | Enable / disable the agent process agent module. |
| agent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| agent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| agent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| agent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| agent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| agent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| agent.resources | object | `{}` | Resources for agent pods. |
| agent.skipSslValidation | bool | `false` | Set to true if self signed certificates are used. |
| agent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| agent.updateStrategy | object | `{"type":"RollingUpdate"}` | The update strategy for the DaemonSet object. |
| clusterAgent.affinity | object | `{}` | Affinity settings for pod assignment. |
| clusterAgent.collection.kubernetesEvents | bool | `true` | Enable / disable the cluster agent events collection. |
| clusterAgent.collection.kubernetesMetrics | bool | `true` | Enable / disable the cluster agent metrics collection. |
| clusterAgent.collection.kubernetesTimeout | int | `10` | Default timeout (in seconds) when obtaining informaton from the Kubernetes API. |
| clusterAgent.collection.kubernetesTopology | bool | `true` | Enable / disable the cluster agent topology collection. |
| clusterAgent.enabled | bool | `true` | Enable / disable the cluster agent. |
| clusterAgent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| clusterAgent.image.repository | string | `"docker.io/stackstate/stackstate-cluster-agent"` | Base container image registry. |
| clusterAgent.image.tag | string | `"2.7.0"` | Default container image tag. |
| clusterAgent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| clusterAgent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| clusterAgent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| clusterAgent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| clusterAgent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| clusterAgent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| clusterAgent.logLevel | string | `"DEBUG"` | Logging level for agent processes. |
| clusterAgent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| clusterAgent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| clusterAgent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| clusterAgent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| clusterAgent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| clusterAgent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| clusterAgent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| clusterAgent.replicaCount | int | `1` | Number of replicas of the cluster agent to deploy. |
| clusterAgent.resources.limits.cpu | string | `"200m"` | CPU resource limits. |
| clusterAgent.resources.limits.memory | string | `"256Mi"` | Memory resource limits. |
| clusterAgent.resources.requests.cpu | string | `"50m"` | CPU resource requests. |
| clusterAgent.resources.requests.memory | string | `"64Mi"` | Memory resource requests. |
| clusterAgent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| dependencies.kubeStateMetrics.enabled | bool | `true` | Whether or not to install the `kube-state-metrics` Deployment along with the StackState cluster agent. Set to `false` if you have `kube-state-metrics` already installed on the cluster. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| global.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| global.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| global.imagePullSecrets | list | `[]` | Secrets / credentials needed for container image registry. |
| nameOverride | string | `""` | Override the name of the chart. |
| stackstate.apiKey | string | `nil` | **PROVIDE YOUR API KEY HERE** API key to be used by the StackState agent. |
| stackstate.cluster.authToken | string | `""` | Provide a token to enable secure communication between the agent and the cluster agent. |
| stackstate.cluster.name | string | `nil` | **PROVIDE KUBERNETES CLUSTER NAME HERE** Name of the Kubernetes cluster where the agent will be installed. |
| stackstate.url | string | `nil` | **PROVIDE STACKSTATE URL HERE** URL of the StackState installation to receive data from the agent. |
