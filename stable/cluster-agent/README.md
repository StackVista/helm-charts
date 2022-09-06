# cluster-agent

Helm chart for the StackState cluster agent.

Current chart version is `2.2.0`

**Homepage:** <https://github.com/StackVista/stackstate-agent>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | kube-state-metrics | 2.2.5 |

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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agent.affinity | object | `{}` | Affinity settings for pod assignment. |
| agent.apm.enabled | bool | `true` | Enable / disable the agent APM module. |
| agent.checksTagCardinality | string | `"orchestrator"` | low, orchestrator or high. Orchestrator level adds pod_name, high adds display_container_name |
| agent.config | object | `{"override":[]}` |  |
| agent.config.override | list | `[]` | A list of objects containing three keys `name`, `path` and `data`, specifying filenames at specific paths which need to be (potentially) overridden using a mounted configmap |
| agent.containerRuntime.customSocketPath | string | `""` | If the container socket path does not match the default for CRI-O, Containerd or Docker, supply a custom socket path. |
| agent.containers.agent.env | object | `{}` | Additional environment variables for the agent container |
| agent.containers.agent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| agent.containers.agent.image.repository | string | `"stackstate/stackstate-agent-2"` | Base container image repository. |
| agent.containers.agent.image.tag | string | `"2.17.2"` | Default container image tag. |
| agent.containers.agent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| agent.containers.agent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| agent.containers.agent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| agent.containers.agent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| agent.containers.agent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| agent.containers.agent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| agent.containers.agent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off # If not set, fall back to the value of agent.logLevel. |
| agent.containers.agent.processAgent.enabled | bool | `false` | Enable / disable the agent process agent module. - deprecated |
| agent.containers.agent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| agent.containers.agent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| agent.containers.agent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| agent.containers.agent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| agent.containers.agent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| agent.containers.agent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| agent.containers.agent.resources.limits.cpu | string | `"400m"` | Memory resource limits. |
| agent.containers.agent.resources.limits.memory | string | `"256Mi"` |  |
| agent.containers.agent.resources.requests.cpu | string | `"100m"` | Memory resource requests. |
| agent.containers.agent.resources.requests.memory | string | `"128Mi"` |  |
| agent.containers.processAgent.enabled | bool | `true` | Enable / disable the process agent container. |
| agent.containers.processAgent.env | object | `{}` | Additional environment variables for the process-agent container |
| agent.containers.processAgent.image.pullPolicy | string | `"IfNotPresent"` | Process-agent container image pull policy. |
| agent.containers.processAgent.image.repository | string | `"stackstate/stackstate-process-agent"` | Process-agent container image repository. |
| agent.containers.processAgent.image.tag | string | `"4.0.7"` | Default process-agent container image tag. |
| agent.containers.processAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off # If not set, fall back to the value of agent.logLevel. |
| agent.containers.processAgent.resources.limits.cpu | string | `"400m"` | Memory resource limits. |
| agent.containers.processAgent.resources.limits.memory | string | `"768Mi"` |  |
| agent.containers.processAgent.resources.requests.cpu | string | `"100m"` | Memory resource requests. |
| agent.containers.processAgent.resources.requests.memory | string | `"128Mi"` |  |
| agent.logLevel | string | `"INFO"` | Logging level for agent processes. |
| agent.networkTracing.enabled | bool | `true` | Enable / disable the agent network tracing module. |
| agent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| agent.priorityClassName | string | `""` | Priority class for agent pods. |
| agent.protocolInspection.enabled | bool | `true` | Enable / disable the agent protocol inspection. |
| agent.scc.enabled | bool | `false` | Enable / disable the installation of the SecurityContextConfiguration needed for installation on OpenShift. |
| agent.service | object | `{"annotations":{},"loadBalancerSourceRanges":["10.0.0.0/8"],"type":"ClusterIP"}` | The Kubernetes service for the agent |
| agent.service.annotations | object | `{}` | Annotations for the service |
| agent.service.loadBalancerSourceRanges | list | `["10.0.0.0/8"]` | The IP4 CIDR allowed to reach LoadBalancer for the service. For LoadBalancer type of service only. |
| agent.service.type | string | `"ClusterIP"` | Type of Kubernetes service: ClusterIP, LoadBalancer, NodePort |
| agent.serviceaccount.annotations | object | `{}` | Annotations for the service account for the agent daemonset pods |
| agent.skipSslValidation | bool | `false` | Set to true if self signed certificates are used. |
| agent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| agent.updateStrategy | object | `{"rollingUpdate":{"maxUnavailable":100},"type":"RollingUpdate"}` | The update strategy for the DaemonSet object. |
| all.image.registry | string | `"quay.io"` | The image registry to use. |
| clusterAgent.affinity | object | `{}` | Affinity settings for pod assignment. |
| clusterAgent.collection.kubernetesEvents | bool | `true` | Enable / disable the cluster agent events collection. |
| clusterAgent.collection.kubernetesMetrics | bool | `true` | Enable / disable the cluster agent metrics collection. |
| clusterAgent.collection.kubernetesResources.configmaps | bool | `true` | Enable / disable collection of ConfigMaps. |
| clusterAgent.collection.kubernetesResources.cronjobs | bool | `true` | Enable / disable collection of CronJobs. |
| clusterAgent.collection.kubernetesResources.daemonsets | bool | `true` | Enable / disable collection of DaemonSets. |
| clusterAgent.collection.kubernetesResources.deployments | bool | `true` | Enable / disable collection of Deployments. |
| clusterAgent.collection.kubernetesResources.endpoints | bool | `true` | Enable / disable collection of Endpoints. If endpoints are disabled then StackState won't be able to connect a Service to Pods that serving it |
| clusterAgent.collection.kubernetesResources.ingresses | bool | `true` | Enable / disable collection of Ingresses. |
| clusterAgent.collection.kubernetesResources.jobs | bool | `true` | Enable / disable collection of Jobs. |
| clusterAgent.collection.kubernetesResources.namespaces | bool | `true` | Enable / disable collection of Namespaces. |
| clusterAgent.collection.kubernetesResources.persistentvolumeclaims | bool | `true` | Enable / disable collection of PersistentVolumeClaims. Disabling these will not let StackState connect PersistentVolumes to pods they are attached to |
| clusterAgent.collection.kubernetesResources.persistentvolumes | bool | `true` | Enable / disable collection of PersistentVolumes. |
| clusterAgent.collection.kubernetesResources.replicasets | bool | `true` | Enable / disable collection of ReplicaSets. |
| clusterAgent.collection.kubernetesResources.secrets | bool | `true` | Enable / disable collection of Secrets. |
| clusterAgent.collection.kubernetesResources.statefulsets | bool | `true` | Enable / disable collection of StatefulSets. |
| clusterAgent.collection.kubernetesTimeout | int | `10` | Default timeout (in seconds) when obtaining information from the Kubernetes API. |
| clusterAgent.collection.kubernetesTopology | bool | `true` | Enable / disable the cluster agent topology collection. |
| clusterAgent.config | object | `{"configMap":{"maxDataSize":null},"override":[],"topology":{"collectionInterval":90}}` |  |
| clusterAgent.config.configMap.maxDataSize | string | `nil` | Maximum amount of characters for the data property of a ConfigMap collected by the kubernetes topology check |
| clusterAgent.config.override | list | `[]` | A list of objects containing three keys `name`, `path` and `data`, specifying filenames at specific paths which need to be (potentially) overridden using a mounted configmap |
| clusterAgent.config.topology.collectionInterval | int | `90` | Interval for running topology collection, in seconds |
| clusterAgent.enabled | bool | `true` | Enable / disable the cluster agent. |
| clusterAgent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| clusterAgent.image.repository | string | `"stackstate/stackstate-cluster-agent"` | Base container image repository. |
| clusterAgent.image.tag | string | `"2.17.2"` | Default container image tag. |
| clusterAgent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| clusterAgent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| clusterAgent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| clusterAgent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| clusterAgent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| clusterAgent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| clusterAgent.logLevel | string | `"INFO"` | Logging level for cluster-agent processes. |
| clusterAgent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| clusterAgent.priorityClassName | string | `""` | Priority class for cluster-agent pods. |
| clusterAgent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| clusterAgent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| clusterAgent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| clusterAgent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| clusterAgent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| clusterAgent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| clusterAgent.replicaCount | int | `1` | Number of replicas of the cluster agent to deploy. |
| clusterAgent.resources.limits.cpu | string | `"400m"` | CPU resource limits. |
| clusterAgent.resources.limits.memory | string | `"1024Mi"` | Memory resource limits. |
| clusterAgent.resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| clusterAgent.resources.requests.memory | string | `"256Mi"` | Memory resource requests. |
| clusterAgent.serviceaccount.annotations | object | `{}` | Annotations for the service account for the cluster agent pods |
| clusterAgent.strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| clusterAgent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| clusterChecks.affinity | object | `{}` | Affinity settings for pod assignment. |
| clusterChecks.apm.enabled | bool | `true` | Enable / disable the agent APM module. |
| clusterChecks.checksTagCardinality | string | `"orchestrator"` |  |
| clusterChecks.config | object | `{"override":[]}` |  |
| clusterChecks.config.override | list | `[]` | A list of objects containing three keys `name`, `path` and `data`, specifying filenames at specific paths which need to be (potentially) overridden using a mounted configmap |
| clusterChecks.enabled | bool | `true` | Enable / disable runnning cluster checks in a separately deployed pod |
| clusterChecks.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| clusterChecks.image.repository | string | `"stackstate/stackstate-agent-2"` | Base container image repository. |
| clusterChecks.image.tag | string | `"2.17.2"` | Default container image tag. |
| clusterChecks.kubeStateMetrics.url | string | `""` | URL of the KubeStateMetrics server. This needs to be configured if the KubeStateMetrics server is not enabled by default in this Helm chart. |
| clusterChecks.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| clusterChecks.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| clusterChecks.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| clusterChecks.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| clusterChecks.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| clusterChecks.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| clusterChecks.logLevel | string | `"INFO"` | Logging level for clusterchecks agent processes. |
| clusterChecks.networkTracing.enabled | bool | `true` | Enable / disable the agent network tracing module. |
| clusterChecks.nodeSelector | object | `{}` | Node labels for pod assignment. |
| clusterChecks.priorityClassName | string | `""` | Priority class for clusterchecks agent pods. |
| clusterChecks.processAgent.enabled | bool | `true` | Enable / disable the agent process agent module. |
| clusterChecks.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| clusterChecks.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| clusterChecks.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| clusterChecks.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| clusterChecks.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| clusterChecks.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| clusterChecks.replicas | int | `1` | Number of clusterchecks agent pods to schedule |
| clusterChecks.resources.limits.cpu | string | `"400m"` | Memory resource limits. |
| clusterChecks.resources.limits.memory | string | `"1024Mi"` |  |
| clusterChecks.resources.requests.cpu | string | `"100m"` | Memory resource requests. |
| clusterChecks.resources.requests.memory | string | `"256Mi"` |  |
| clusterChecks.scc.enabled | bool | `false` | Enable / disable the installation of the SecurityContextConfiguration needed for installation on OpenShift |
| clusterChecks.serviceaccount.annotations | object | `{}` | Annotations for the service account for the cluster checks pods |
| clusterChecks.skipSslValidation | bool | `false` | Set to true if self signed certificates are used. |
| clusterChecks.strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| clusterChecks.tolerations | list | `[]` | Toleration labels for pod assignment. |
| dependencies.kubeStateMetrics.enabled | bool | `true` | Whether or not to install the `kube-state-metrics` Deployment along with the StackState cluster agent. Set to `false` if you have `kube-state-metrics` already installed on the cluster. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| global.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| global.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| global.imagePullSecrets | list | `[]` | Secrets / credentials needed for container image registry. |
| kube-state-metrics.image | object | `{"registry":"quay.io","repository":"stackstate/kube-state-metrics","tag":"2.3.0-focal-20220316-r61.20220418.2032"}` | Details about the docker image to be used for this component. This overrides the value in the bitnami chart. |
| kube-state-metrics.image.registry | string | `"quay.io"` | Registry where docker image will be pulled from. This overrides the value in the bitnami chart. |
| kube-state-metrics.image.repository | string | `"stackstate/kube-state-metrics"` | The path inside the registry where the image is hosted. This overrides the value in the bitnami chart. |
| kube-state-metrics.image.tag | string | `"2.3.0-focal-20220316-r61.20220418.2032"` | The version tag of the image to be used during deployment. This overrides the value in the bitnami chart. |
| nameOverride | string | `""` | Override the name of the chart. |
| processAgent.checkIntervals.connections | int | `30` | Override the default value of the connections check interval in seconds. |
| processAgent.checkIntervals.container | int | `30` | Override the default value of the container check interval in seconds. |
| processAgent.checkIntervals.process | int | `30` | Override the default value of the process check interval in seconds. |
| stackstate.apiKey | string | `nil` | **PROVIDE YOUR API KEY HERE** API key to be used by the StackState agent. |
| stackstate.cluster.authToken | string | `""` | Provide a token to enable secure communication between the agent and the cluster agent. |
| stackstate.cluster.name | string | `nil` | **PROVIDE KUBERNETES CLUSTER NAME HERE** Name of the Kubernetes cluster where the agent will be installed. |
| stackstate.url | string | `nil` | **PROVIDE STACKSTATE URL HERE** URL of the StackState installation to receive data from the agent. |
