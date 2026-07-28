# suse-observability-agent

Helm chart for the SUSE observability Agent.

Current chart version is `1.5.39`

**Homepage:** <https://github.com/StackVista/suse-observability-agent>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../../local/http-header-injector/ | httpHeaderInjectorWebhook(http-header-injector) | * |
| file://../../local/kubernetes-rbac-agent/ | kubernetes-rbac-agent | * |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:

* `stackstate.apiKey`
* `stackstate.cluster.name`
* `stackstate.url`

The parameter `stackstate.cluster.name` is entered when installing the Cluster Agent StackPack.

The recommended namespace for this chart is `suse-observability-agent`.

Install them on the command line on Helm with the following command:

```shell
helm install \
--namespace suse-observability-agent \
--create-namespace \
--set-string 'stackstate.apiKey'='<your-api-key>' \
--set-string 'stackstate.cluster.name'='<your-cluster-name>' \
--set-string 'stackstate.url'='<your-stackstate-url>' \
stackstate/suse-observability-agent
```

## Recommended Values

It is also recommended that you set a value for `stackstate.cluster.authToken`. If it is not provided, a value will be generated for you, but the value will change each time an upgrade is performed.

The command for **also** installing with a set token would be:

```shell
helm install \
--namespace suse-observability-agent \
--create-namespace \
--set-string 'stackstate.apiKey'='<your-api-key>' \
--set-string 'stackstate.cluster.name'='<your-cluster-name>' \
--set-string 'stackstate.cluster.authToken'='<your-cluster-token>' \
--set-string 'stackstate.url'='<your-stackstate-url>' \
stackstate/suse-observability-agent
```

## Integration overlays

The `integrations/` directory contains opt-in values overlays that pre-configure the `k8sResourceCollector` for individual SUSE stackpacks. Each overlay enables the API groups consumed by that stackpack's component- and relation-mappings under both `crdDiscovery.apiGroupFilters.include` and `rbac.crdApiGroups` (the latter only matters when `rbac.useWildcard: false`).

Available overlays:

* `integrations/suse-runtime-enforcer.yaml`
* `integrations/suse-admission-controller.yaml`
* `integrations/suse-virtualization.yaml`
* `integrations/suse-sbom-scanner.yaml`
* `integrations/suse-bundle.yaml` — convenience overlay combining all of the above

Apply one or more with `-f`:

```shell
helm install \
--namespace suse-observability-agent \
--create-namespace \
--set-string 'stackstate.apiKey'='<your-api-key>' \
--set-string 'stackstate.cluster.name'='<your-cluster-name>' \
--set-string 'stackstate.url'='<your-stackstate-url>' \
-f integrations/suse-runtime-enforcer.yaml \
-f integrations/suse-admission-controller.yaml \
stackstate/suse-observability-agent
```

Overlays use map-shaped values, so combining multiple overlays (or layering them on top of your own `apiGroupFilters` / `crdApiGroups` entries) is additive. Listing an API group on a cluster that does not have the corresponding CRDs installed is harmless — the receiver only watches resources that actually exist.

## OTel Prometheus scraping

When `otel.enabled=true` (default=false) and `otel.prometheusScraping.enabled=true` (default=true), the chart deploys an OpenTelemetry Collector based metrics scraper plus a Target Allocator that discovers `ServiceMonitor` and `PodMonitor` resources and distributes their scrape targets across collector pods.

### Opting monitors in

By default the Target Allocator only picks up monitors that carry the `observability.suse.com/agent: scrape` label. Customize via `otel.prometheusScraping.targetAllocator.prometheusCR.serviceMonitorSelector` and `podMonitorSelector`.

### Endpoint auth (basicAuth, bearerTokenSecret, oauth2, tlsConfig)

The Target Allocator has no cluster-wide secrets access. When a `ServiceMonitor` or `PodMonitor` references an endpoint auth secret, you must:

1. List the secret's namespace in `otel.prometheusScraping.targetAllocator.prometheusCR.secretNamespaces`.
2. Deploy a `Role` and `RoleBinding` in that namespace granting the Target Allocator's `ServiceAccount` (which lives in the chart's release namespace) read access to secrets.

#### Securing credentials in transit (mTLS)

By default, `otel.prometheusScraping.targetAllocator.allowInsecureAuthSecrets=false` and `otel.prometheusScraping.targetAllocator.mtlsEnabled=false`. In this state the Target Allocator will not serve auth secrets to collectors at all, so ServiceMonitors and PodMonitors that reference secrets require one of the two options below.

**Option 1 — mTLS (recommended):** Enable mTLS so secrets are served over a mutually authenticated TLS connection. Requires cert-manager.

```yaml
otel:
  prometheusScraping:
    targetAllocator:
      mtlsEnabled: true
      allowInsecureAuthSecrets: false  # default, explicit for clarity
```

**Option 2 — plain HTTP (only for isolated clusters):** Allow secrets to be served over plain HTTP by setting `allowInsecureAuthSecrets=true`. Credentials will travel unencrypted between pods.

```yaml
otel:
  prometheusScraping:
    targetAllocator:
      allowInsecureAuthSecrets: true
```

#### Full example (without mTLS)

Assume the chart is installed as release `suse-observability-agent` in namespace `suse-observability-agent`, and you want to scrape an application in namespace `payments` whose `/metrics` endpoint requires a bearer token:

```yaml
# Helm values
otel:
  prometheusScraping:
    enabled: true
    targetAllocator:
      prometheusCR:
        secretNamespaces:
          - payments
```

```yaml
# Resources you apply in the payments namespace
apiVersion: v1
kind: Secret
metadata:
  name: payments-metrics-auth
  namespace: payments
type: Opaque
stringData:
  token: <bearer-token>
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: payments-api
  namespace: payments
  labels:
    observability.suse.com/agent: scrape
spec:
  selector:
    matchLabels:
      app: payments-api
  endpoints:
    - port: metrics
      path: /metrics
      bearerTokenSecret:
        name: payments-metrics-auth
        key: token
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: suse-observability-agent-otel-target-allocator-secrets
  namespace: payments
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: suse-observability-agent-otel-target-allocator-secrets
  namespace: payments
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: suse-observability-agent-otel-target-allocator-secrets
subjects:
  - kind: ServiceAccount
    name: suse-observability-agent-otel-target-allocator
    namespace: suse-observability-agent
```

Repeat the `Role`+`RoleBinding` per namespace listed in `secretNamespaces`. The `ServiceAccount` subject always lives in the chart's release namespace; only the `Role`/`RoleBinding` move to the secret's namespace.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| all.fullPrivilegesMode.enabled | bool | `false` | All agent pods run in full privileges mode (mainly used for debugging). |
| all.image.registry | string | `nil` | The image registry to use. |
| checksAgent.affinity | object | `{}` | Affinity settings for pod assignment. |
| checksAgent.checksTagCardinality | string | `"orchestrator"` | low, orchestrator or high. Orchestrator level adds pod_name, high adds display_container_name |
| checksAgent.config | object | `{"override":[]}` |  |
| checksAgent.config.override | list | `[]` | A list of objects containing three keys `name`, `path` and `data`, specifying filenames at specific paths which need to be (potentially) overridden using a mounted configmap |
| checksAgent.enabled | bool | `true` | Enable / disable runnning cluster checks in a separately deployed pod |
| checksAgent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| checksAgent.image.repository | string | `"stackstate/stackstate-k8s-agent"` | Base container image repository. |
| checksAgent.image.tag | string | `"21f456fd"` | Default container image tag. |
| checksAgent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| checksAgent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| checksAgent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| checksAgent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| checksAgent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| checksAgent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| checksAgent.logLevel | string | `"INFO"` | Logging level for clusterchecks agent processes. |
| checksAgent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| checksAgent.priorityClassName | string | `""` | Priority class for clusterchecks agent pods. |
| checksAgent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| checksAgent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| checksAgent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| checksAgent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| checksAgent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| checksAgent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| checksAgent.replicas | int | `1` | Number of clusterchecks agent pods to schedule |
| checksAgent.resources.limits.cpu | string | `"400m"` | CPU resource limits. |
| checksAgent.resources.limits.memory | string | `"600Mi"` | Memory resource limits. |
| checksAgent.resources.requests.cpu | string | `"20m"` | CPU resource requests. |
| checksAgent.resources.requests.memory | string | `"512Mi"` | Memory resource requests. |
| checksAgent.serviceaccount.annotations | object | `{}` | Annotations for the service account for the cluster checks pods |
| checksAgent.skipSslValidation | bool | `false` | Set to true if self signed certificates are used. |
| checksAgent.strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| checksAgent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| clusterAgent.affinity | object | `{}` | Affinity settings for pod assignment. |
| clusterAgent.collection.kubeStateMetrics.annotationsAsTags | object | `{}` | Extra annotations to collect from resources and to turn into tags. |
| clusterAgent.collection.kubeStateMetrics.clusterCheck | bool | `false` | For large clusters where the Kubernetes State Metrics Check Core needs to be distributed on dedicated workers. |
| clusterAgent.collection.kubeStateMetrics.enabled | bool | `true` | Enable / disable the cluster agent kube-state-metrics collection. |
| clusterAgent.collection.kubeStateMetrics.labelsAsTags | object | `{}` | Extra labels to collect from resources and to turn into StackState tag. # It has the following structure: # labelsAsTags: #   <resource1>:        # can be pod, deployment, node, etc. #     <label1>: <tag1>  # where <label1> is the kubernetes label and <tag1> is the StackState tag #     <label2>: <tag2> #   <resource2>: #     <label3>: <tag3> # # Warning: the label must match the transformation done by kube-state-metrics, # for example tags.example/version becomes tags_example_version. |
| clusterAgent.collection.kubernetesEvents | bool | `true` | Enable / disable the cluster agent events collection. |
| clusterAgent.collection.kubernetesMetrics | bool | `true` | Enable / disable the cluster agent metrics collection. |
| clusterAgent.collection.kubernetesResources.configmaps | bool | `true` | Enable / disable collection of ConfigMaps. |
| clusterAgent.collection.kubernetesResources.cronjobs | bool | `true` | Enable / disable collection of CronJobs. |
| clusterAgent.collection.kubernetesResources.daemonsets | bool | `true` | Enable / disable collection of DaemonSets. |
| clusterAgent.collection.kubernetesResources.deployments | bool | `true` | Enable / disable collection of Deployments. |
| clusterAgent.collection.kubernetesResources.endpoints | bool | `true` | Enable / disable collection of Endpoints. If endpoints are disabled it is not possible to connect a Service to Pods that serving it |
| clusterAgent.collection.kubernetesResources.horizontalpodautoscalers | bool | `true` | Enable / disable collection of HorizontalPodAutoscalers. |
| clusterAgent.collection.kubernetesResources.ingresses | bool | `true` | Enable / disable collection of Ingresses. |
| clusterAgent.collection.kubernetesResources.jobs | bool | `true` | Enable / disable collection of Jobs. |
| clusterAgent.collection.kubernetesResources.limitranges | bool | `true` | Enable / disable collection of LimitRanges. |
| clusterAgent.collection.kubernetesResources.namespaces | bool | `true` | Enable / disable collection of Namespaces. |
| clusterAgent.collection.kubernetesResources.persistentvolumeclaims | bool | `true` | Enable / disable collection of PersistentVolumeClaims. Disabling these makes it impossible to create relations between PersistentVolumes and pods |
| clusterAgent.collection.kubernetesResources.persistentvolumes | bool | `true` | Enable / disable collection of PersistentVolumes. |
| clusterAgent.collection.kubernetesResources.poddisruptionbudgets | bool | `true` | Enable / disable collection of PodDisruptionBudgets. |
| clusterAgent.collection.kubernetesResources.replicasets | bool | `true` | Enable / disable collection of ReplicaSets. |
| clusterAgent.collection.kubernetesResources.replicationcontrollers | bool | `true` | Enable / disable collection of ReplicationControllers. |
| clusterAgent.collection.kubernetesResources.resourcequotas | bool | `true` | Enable / disable collection of ResourceQuotas. |
| clusterAgent.collection.kubernetesResources.secrets | bool | `true` | Enable / disable collection of Secrets. |
| clusterAgent.collection.kubernetesResources.statefulsets | bool | `true` | Enable / disable collection of StatefulSets. |
| clusterAgent.collection.kubernetesResources.storageclasses | bool | `true` | Enable / disable collection of StorageClasses. |
| clusterAgent.collection.kubernetesResources.volumeattachments | bool | `true` | Enable / disable collection of Volume Attachments. Used to bind Nodes to Persistent Volumes. |
| clusterAgent.collection.kubernetesTimeout | int | `10` | Default timeout (in seconds) when obtaining information from the Kubernetes API. |
| clusterAgent.collection.kubernetesTopology | bool | `true` | Enable / disable the cluster agent topology collection. |
| clusterAgent.config | object | `{"configMap":{"maxDataSize":null},"events":{"categories":{}},"override":[],"topology":{"collectionInterval":90}}` |  |
| clusterAgent.config.configMap.maxDataSize | string | `nil` | Maximum amount of characters for the data property of a ConfigMap collected by the kubernetes topology check |
| clusterAgent.config.events.categories | object | `{}` | Custom mapping from Kubernetes event reason to StackState event category. Categories allowed: Alerts, Activities, Changes, Others |
| clusterAgent.config.override | list | `[]` | A list of objects containing three keys `name`, `path` and `data`, specifying filenames at specific paths which need to be (potentially) overridden using a mounted configmap |
| clusterAgent.config.topology.collectionInterval | int | `90` | Interval for running topology collection, in seconds |
| clusterAgent.enabled | bool | `true` | Enable / disable the cluster agent. |
| clusterAgent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| clusterAgent.image.repository | string | `"stackstate/stackstate-k8s-cluster-agent"` | Base container image repository. |
| clusterAgent.image.tag | string | `"21f456fd"` | Default container image tag. |
| clusterAgent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| clusterAgent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| clusterAgent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| clusterAgent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| clusterAgent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| clusterAgent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| clusterAgent.logLevel | string | `"INFO"` | Logging level for stackstate-k8s-agent processes. |
| clusterAgent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| clusterAgent.priorityClassName | string | `""` | Priority class for stackstate-k8s-agent pods. |
| clusterAgent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| clusterAgent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| clusterAgent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| clusterAgent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| clusterAgent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| clusterAgent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| clusterAgent.replicaCount | int | `1` | Number of replicas of the cluster agent to deploy. |
| clusterAgent.resources.limits.cpu | string | `"400m"` | CPU resource limits. |
| clusterAgent.resources.limits.memory | string | `"800Mi"` | Memory resource limits. |
| clusterAgent.resources.requests.cpu | string | `"70m"` | CPU resource requests. |
| clusterAgent.resources.requests.memory | string | `"512Mi"` | Memory resource requests. |
| clusterAgent.service.port | int | `5005` | Change the Cluster Agent service port |
| clusterAgent.service.targetPort | int | `5005` | Change the Cluster Agent service targetPort |
| clusterAgent.serviceaccount.annotations | object | `{}` | Annotations for the service account for the cluster agent pods |
| clusterAgent.skipSslValidation | bool | `false` | If true, ignores the server certificate being signed by an unknown authority. |
| clusterAgent.strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| clusterAgent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| global.apiKey.fromSecret | string | `"{{ include \"stackstate-k8s-agent.secret.internal.name\" . }}"` | The secret from which the receiver api key is taken. Will execute as a template. Overriding this will allow setting the api key from an externally provided secret. The api key will be picked form the STS_API_KEY value |
| global.clusterAgentAuthToken.fromSecret | string | `"{{ include \"stackstate-k8s-agent.secret.internal.name\" . }}"` | The secret from from which the token for authenticating between node and cluster agent will be taken. Overriding this will allow setting the api key from an externally provided secret. The api key will be picked form the STS_CLUSTER_AGENT_AUTH_TOKEN value |
| global.customCertificates | object | `{"configMapName":"","enabled":false,"pemData":""}` | Custom certificates for HTTPS endpoints |
| global.customCertificates.configMapName | string | `""` | Name of existing ConfigMap containing certificates (exclusive with pemData) |
| global.customCertificates.enabled | bool | `false` | Enable custom certificate injection |
| global.customCertificates.pemData | string | `""` | PEM-encoded certificate data (exclusive with configMapName), will be stored as tls.pem |
| global.extraAnnotations | object | `{}` | Extra annotations added ta all resources created by the helm chart |
| global.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| global.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| global.extraLabels | object | `{}` | Extra labels added ta all resources created by the helm chart |
| global.features.experimentalStackpacks | bool | `false` | Enable StackPacks 2.0 to signal to all components that they should support the StackPacks 2.0 spec. When enabled, the cluster collector (CRD discovery) is automatically activated. |
| global.imagePullCredentials | object | `{}` | Globally define credentials for pulling images. |
| global.imagePullSecrets | list | `[]` | Secrets / credentials needed for container image registry. |
| global.imageRegistry | string | `"quay.io"` | The image registry to use. |
| global.proxy.url | string | `""` | Proxy for all traffic to stackstate |
| global.skipSslValidation | bool | `false` | Enable tls validation from client |
| httpHeaderInjectorWebhook.certificatePrehook.image.repository | string | `"stackstate/container-tools"` |  |
| httpHeaderInjectorWebhook.certificatePrehook.image.tag | string | `"1.8.6-so17"` |  |
| httpHeaderInjectorWebhook.enabled | bool | `false` | Enable the webhook for injection http header injection sidecar proxy |
| httpHeaderInjectorWebhook.proxy.image.repository | string | `"stackstate/http-header-injector-proxy"` |  |
| httpHeaderInjectorWebhook.proxy.image.tag | string | `"1.38.3-so2"` |  |
| httpHeaderInjectorWebhook.proxyInit.image.repository | string | `"stackstate/http-header-injector-proxy-init"` |  |
| httpHeaderInjectorWebhook.proxyInit.image.tag | string | `"1.0.0-so4"` |  |
| httpHeaderInjectorWebhook.sidecarInjector.image.repository | string | `"stackstate/generic-sidecar-injector"` |  |
| httpHeaderInjectorWebhook.sidecarInjector.image.tag | string | `"b8ac81b7-1025-release"` |  |
| kubernetes-rbac-agent.clusterName.fromConfigMap | string | `"{{ include \"stackstate-k8s-agent.clusterName.configmap.internal.name\" . }}"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.affinity | object | `{}` | Set affinity |
| kubernetes-rbac-agent.containers.rbacAgent.env | object | `{}` | Additional environment variables |
| kubernetes-rbac-agent.containers.rbacAgent.image.repository | string | `"stackstate/kubernetes-rbac-agent"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.image.tag | string | `"4ec35439-1023-release"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.nodeSelector | object | `{}` | Set a nodeSelector |
| kubernetes-rbac-agent.containers.rbacAgent.podAnnotations | object | `{}` | Additional annotations on the pod |
| kubernetes-rbac-agent.containers.rbacAgent.podLabels | object | `{}` | Additional labels on the pod |
| kubernetes-rbac-agent.containers.rbacAgent.priorityClassName | string | `""` | Set priorityClassName |
| kubernetes-rbac-agent.containers.rbacAgent.resources.limits.memory | string | `"40Mi"` | Memory resource limits. |
| kubernetes-rbac-agent.containers.rbacAgent.resources.requests.memory | string | `"25Mi"` | Memory resource requests. |
| kubernetes-rbac-agent.containers.rbacAgent.tolerations | list | `[]` | Set tolerations |
| kubernetes-rbac-agent.enabled | bool | `true` |  |
| kubernetes-rbac-agent.roleType | string | `"scope"` |  |
| kubernetes-rbac-agent.url.fromConfigMap | string | `"{{ include \"stackstate-k8s-agent.url.configmap.internal.name\" . }}"` |  |
| logsAgent.affinity | object | `{}` | Affinity settings for pod assignment. |
| logsAgent.enabled | bool | `true` | Enable / disable k8s pod log collection |
| logsAgent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| logsAgent.image.repository | string | `"stackstate/promtail"` | Base container image repository. |
| logsAgent.image.tag | string | `"3.6.11-so10"` | Default container image tag. |
| logsAgent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| logsAgent.priorityClassName | string | `""` | Priority class for logsAgent pods. |
| logsAgent.resources.limits.cpu | string | `"1300m"` | CPU resource limits. |
| logsAgent.resources.limits.memory | string | `"192Mi"` | Memory resource limits. |
| logsAgent.resources.requests.cpu | string | `"20m"` | CPU resource requests. |
| logsAgent.resources.requests.memory | string | `"100Mi"` | Memory resource requests. |
| logsAgent.serviceaccount.annotations | object | `{}` | Annotations for the service account for the daemonset pods |
| logsAgent.skipSslValidation | bool | `false` | If true, ignores the server certificate being signed by an unknown authority. |
| logsAgent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| logsAgent.updateStrategy | object | `{"rollingUpdate":{"maxUnavailable":100},"type":"RollingUpdate"}` | The update strategy for the DaemonSet object. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeAgent.affinity | object | `{}` | Affinity settings for pod assignment. |
| nodeAgent.autoScalingEnabled | bool | `false` | Enable / disable autoscaling for the node agent pods. |
| nodeAgent.checksTagCardinality | string | `"orchestrator"` | low, orchestrator or high. Orchestrator level adds pod_name, high adds display_container_name |
| nodeAgent.config | object | `{"override":[]}` |  |
| nodeAgent.config.override | list | `[]` | A list of objects containing three keys `name`, `path` and `data`, specifying filenames at specific paths which need to be (potentially) overridden using a mounted configmap |
| nodeAgent.containerRuntime.customSocketPath | string | `""` | If the container socket path does not match the default for CRI-O, Containerd or Docker, supply a custom socket path. |
| nodeAgent.containerRuntime.hostProc | string | `"/proc"` |  |
| nodeAgent.containers.agent.env | object | `{}` | Additional environment variables for the agent container |
| nodeAgent.containers.agent.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| nodeAgent.containers.agent.image.repository | string | `"stackstate/stackstate-k8s-agent"` | Base container image repository. |
| nodeAgent.containers.agent.image.tag | string | `"21f456fd"` | Default container image tag. |
| nodeAgent.containers.agent.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| nodeAgent.containers.agent.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| nodeAgent.containers.agent.livenessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the liveness probe. |
| nodeAgent.containers.agent.livenessProbe.periodSeconds | int | `15` | `periodSeconds` for the liveness probe. |
| nodeAgent.containers.agent.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| nodeAgent.containers.agent.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| nodeAgent.containers.agent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off # If not set, fall back to the value of agent.logLevel. |
| nodeAgent.containers.agent.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| nodeAgent.containers.agent.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| nodeAgent.containers.agent.readinessProbe.initialDelaySeconds | int | `15` | `initialDelaySeconds` for the readiness probe. |
| nodeAgent.containers.agent.readinessProbe.periodSeconds | int | `15` | `periodSeconds` for the readiness probe. |
| nodeAgent.containers.agent.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| nodeAgent.containers.agent.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| nodeAgent.containers.agent.resources.limits.cpu | string | `"270m"` | CPU resource limits. |
| nodeAgent.containers.agent.resources.limits.memory | string | `"420Mi"` | Memory resource limits. |
| nodeAgent.containers.agent.resources.requests.cpu | string | `"20m"` | CPU resource requests. |
| nodeAgent.containers.agent.resources.requests.memory | string | `"180Mi"` | Memory resource requests. |
| nodeAgent.containers.processAgent.enabled | bool | `true` | Enable / disable the process agent container. |
| nodeAgent.containers.processAgent.env | object | `{}` | Additional environment variables for the process-agent container |
| nodeAgent.containers.processAgent.image.pullPolicy | string | `"IfNotPresent"` | Process-agent container image pull policy. |
| nodeAgent.containers.processAgent.image.registry | string | `nil` |  |
| nodeAgent.containers.processAgent.image.repository | string | `"stackstate/stackstate-k8s-process-agent"` | Process-agent container image repository. |
| nodeAgent.containers.processAgent.image.tag | string | `"b9f7fe00"` | Default process-agent container image tag. |
| nodeAgent.containers.processAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off # If not set, fall back to the value of agent.logLevel. |
| nodeAgent.containers.processAgent.procVolumeReadOnly | bool | `true` | Configure whether /host/proc is read only for the process agent container |
| nodeAgent.containers.processAgent.resources.limits.cpu | string | `"125m"` | CPU resource limits. |
| nodeAgent.containers.processAgent.resources.limits.memory | string | `"400Mi"` | Memory resource limits. |
| nodeAgent.containers.processAgent.resources.requests.cpu | string | `"25m"` | CPU resource requests. |
| nodeAgent.containers.processAgent.resources.requests.memory | string | `"128Mi"` | Memory resource requests. |
| nodeAgent.httpTracing.enabled | bool | `true` | Enable / disable the process-agent HTTP tracing. |
| nodeAgent.logLevel | string | `"INFO"` | Logging level for agent processes. |
| nodeAgent.nodeSelector | object | `{}` | Node labels for pod assignment. |
| nodeAgent.priorityClassName | string | `""` | Priority class for nodeAgent pods. |
| nodeAgent.protocolInspection.enabled | bool | `true` | Enable / disable the nodeAgent protocol inspection. |
| nodeAgent.scaling.autoscalerLimits.agent.maximum.cpu | string | `"200m"` | Maximum CPU resource limits for main agent. |
| nodeAgent.scaling.autoscalerLimits.agent.maximum.memory | string | `"450Mi"` | Maximum memory resource limits for main agent. |
| nodeAgent.scaling.autoscalerLimits.agent.minimum.cpu | string | `"20m"` | Minimum CPU resource limits for main agent. |
| nodeAgent.scaling.autoscalerLimits.agent.minimum.memory | string | `"180Mi"` | Minimum memory resource limits for main agent. |
| nodeAgent.scaling.autoscalerLimits.processAgent.maximum.cpu | string | `"200m"` | Maximum CPU resource limits for process agent. |
| nodeAgent.scaling.autoscalerLimits.processAgent.maximum.memory | string | `"500Mi"` | Maximum memory resource limits for process agent. |
| nodeAgent.scaling.autoscalerLimits.processAgent.minimum.cpu | string | `"25m"` | Minimum CPU resource limits for process agent. |
| nodeAgent.scaling.autoscalerLimits.processAgent.minimum.memory | string | `"100Mi"` | Minimum memory resource limits for process agent. |
| nodeAgent.scc.enabled | bool | `false` | Enable / disable the installation of the SecurityContextConfiguration needed for installation on OpenShift. |
| nodeAgent.service | object | `{"annotations":{},"loadBalancerSourceRanges":["10.0.0.0/8"],"type":"ClusterIP"}` | The Kubernetes service for the agent |
| nodeAgent.service.annotations | object | `{}` | Annotations for the service |
| nodeAgent.service.loadBalancerSourceRanges | list | `["10.0.0.0/8"]` | The IP4 CIDR allowed to reach LoadBalancer for the service. For LoadBalancer type of service only. |
| nodeAgent.service.type | string | `"ClusterIP"` | Type of Kubernetes service: ClusterIP, LoadBalancer, NodePort |
| nodeAgent.serviceaccount.annotations | object | `{}` | Annotations for the service account for the agent daemonset pods |
| nodeAgent.skipKubeletTLSVerify | bool | `false` | Set to true if you want to skip kubelet tls verification. |
| nodeAgent.skipSslValidation | bool | `false` | Set to true if self signed certificates are used. |
| nodeAgent.tolerations | list | `[]` | Toleration labels for pod assignment. |
| nodeAgent.updateStrategy | object | `{"rollingUpdate":{"maxUnavailable":100},"type":"RollingUpdate"}` | The update strategy for the DaemonSet object. |
| nodeAgent.useHostNetwork | bool | `true` | Set to true if you want to deploy the node agent in the host network namespace. |
| nodeAgent.useHostPID | bool | `true` | Set to true if you want to deploy the node agent in the host PID namespace. |
| openShiftLogging.installSecret | bool | `false` | Install a secret for logging on openshift |
| otel.enabled | bool | `false` | Master switch for all OTel components. Set to true to activate OpenTelemetry based features. |
| otel.k8sResourceCollector.affinity | object | `{}` | Affinity settings for pod assignment. |
| otel.k8sResourceCollector.crDiscovery.apiGroups.exclude | object | `{}` | Map of API group patterns (key) -> bool (enabled). Empty by default. |
| otel.k8sResourceCollector.crDiscovery.apiGroups.include | object | `{}` | Map of API group patterns (key) -> bool (enabled). Used for both CR filtering and restricted RBAC. Supports wildcards like "*.suse.com" for filtering, but restricted RBAC (rbac.useWildcard=false) can only render exact API groups or "*". Set a key to false in an override values file to disable a default. Must have at least one truthy entry when discoveryMode is "api_groups". Common SUSE integration API groups are added by the enabled integrations below. |
| otel.k8sResourceCollector.crDiscovery.discoveryMode | string | `"api_groups"` | CR discovery mode: "api_groups" (filtered custom resources) or "all" (watch all custom resources). CRDs are always watched and forwarded. |
| otel.k8sResourceCollector.crDiscovery.snapshotInterval | string | `"5m"` | Interval for periodic snapshot emission from the informer cache (default: 5m, min: 1m) |
| otel.k8sResourceCollector.dataLimits.maxCrTotalDataSizeBytes | int | `10485760` | Total serialized payload budget, in bytes, for CR-discovered Custom Resources per collection cycle. Default: 10 MiB. CRs are considered smallest-first, then by stable identity; CRs that do not fit are dropped and counted in receiver metrics. |
| otel.k8sResourceCollector.dataLimits.maxObjectTotalDataSizeBytes | int | `10485760` | Total serialized payload budget, in bytes, for statically configured Kubernetes object watches per collection cycle. Default: 10 MiB. Objects are considered smallest-first, then by stable identity; objects that do not fit are dropped and counted in receiver metrics. |
| otel.k8sResourceCollector.debug | object | `{"enabled":false,"pipelines":["logs"],"verbosity":"basic"}` | Optional debug exporter for troubleshooting. When enabled, the upstream OTel `debug` exporter is wired into the listed pipelines so payloads are written to the collector log. Leave disabled in production. |
| otel.k8sResourceCollector.debug.enabled | bool | `false` | Enable the debug exporter for this collector. |
| otel.k8sResourceCollector.debug.pipelines | list | `["logs"]` | Pipelines (by signal) to attach the debug exporter to. Must be a subset of {traces, logs, metrics}. |
| otel.k8sResourceCollector.debug.verbosity | string | `"basic"` | Debug exporter verbosity: basic, normal, or detailed. |
| otel.k8sResourceCollector.deniedObjects | object | `{}` | Map of resource name (plural, used as the key) -> spec extending the built-in denylist (core Secrets, ConfigMaps). Spec needs only `group`. Resources listed here must not appear under otel.k8sResourceCollector.objects. Use to block third-party resources with sensitive contents. |
| otel.k8sResourceCollector.enabled | bool | `true` | Enable / disable the OpenTelemetry cluster collector for CRD discovery. Requires otel.enabled=true. |
| otel.k8sResourceCollector.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| otel.k8sResourceCollector.image.repository | string | `"stackstate/sts-opentelemetry-collector"` | Base container image repository. |
| otel.k8sResourceCollector.image.tag | string | `"v0.0.53-agent"` | Container image tag for 'opentelemetry-collector' containers. |
| otel.k8sResourceCollector.integrations.rancher | bool | `false` | Enable Rancher Manager URL enrichment. When true, the collector reads CATTLE_SERVER from the cattle-cluster-agent Deployment in cattle-system and attaches it as a resource attribute on all emitted log records. Defaults to false: on non-Rancher clusters the cattle-system namespace is absent so the necessary Role and RoleBinding cannot be created. |
| otel.k8sResourceCollector.integrations.suseAdmissionController | bool | `true` | Enable pre-configured API group filters for the SUSE Admission Controller (Kubewarden) stackpack. |
| otel.k8sResourceCollector.integrations.suseRuntimeEnforcer | bool | `true` | Enable pre-configured API group filters for the SUSE Runtime Enforcer stackpack. |
| otel.k8sResourceCollector.integrations.suseSbomScanner | bool | `false` | Enable pre-configured API group filters for the SUSE SBOM Scanner stackpack. |
| otel.k8sResourceCollector.integrations.suseVirtualization | bool | `true` | Enable pre-configured API group filters for the SUSE Virtualization (KubeVirt) stackpack. |
| otel.k8sResourceCollector.leaderElection.enabled | bool | `true` | Enable the k8s_leader_elector extension and peer-to-peer cache sync. When enabled, only the leader actively watches CRDs/CRs, and cache state is synced to replicas for fast failover. |
| otel.k8sResourceCollector.leaderElection.leaseDuration | string | `"15s"` | Duration a leader holds the lease before it must renew. |
| otel.k8sResourceCollector.leaderElection.leaseName | string | `"k8sresourcereceiver"` | Name of the Lease object. Must be unique per collector deployment. |
| otel.k8sResourceCollector.leaderElection.renewDeadline | string | `"10s"` | Deadline for the leader to renew the lease. Must be less than leaseDuration. |
| otel.k8sResourceCollector.leaderElection.retryPeriod | string | `"2s"` | How often non-leaders retry acquiring the lease. Must be less than renewDeadline. |
| otel.k8sResourceCollector.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| otel.k8sResourceCollector.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| otel.k8sResourceCollector.livenessProbe.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the liveness probe. |
| otel.k8sResourceCollector.livenessProbe.periodSeconds | int | `10` | `periodSeconds` for the liveness probe. |
| otel.k8sResourceCollector.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| otel.k8sResourceCollector.livenessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the liveness probe. |
| otel.k8sResourceCollector.logLevel | string | `"info"` | Logging level for OpenTelemetry collector (debug, info, warn, error) |
| otel.k8sResourceCollector.nodeSelector | object | `{}` | Node labels for pod assignment. |
| otel.k8sResourceCollector.objects | object | `{}` | Map of resource name (plural, used as the key) -> spec for additional Kubernetes resources to watch alongside CRD-discovered custom resources. Spec fields: group (empty for core), version (preferred if empty), namespaces (cluster-wide if empty), labelSelector, fieldSelector. Set a value to null or false in an override values file to disable a default entry. Entries that overlap a CRD covered by cr_api_groups are rejected at startup. When useWildcard=false, get/list/watch RBAC for each entry is auto-derived (deduped per group). |
| otel.k8sResourceCollector.peerSync.port | int | `4319` | Port for peer-to-peer cache sync HTTP server. Each replica serves its cache on this port. |
| otel.k8sResourceCollector.podAnnotations | object | `{}` | Additional annotations for cluster collector pods. |
| otel.k8sResourceCollector.podLabels | object | `{}` | Additional labels for cluster collector pods. |
| otel.k8sResourceCollector.pprof.enabled | bool | `true` | Enable the pprof extension for profiling/debugging. Opt-out: enabled by default. The pprof endpoint is reachable inside the pod (port 1777) via kubectl port-forward. |
| otel.k8sResourceCollector.priorityClassName | string | `""` | Priority class for cluster collector pods. |
| otel.k8sResourceCollector.rbac.useWildcard | bool | `true` | Use wildcard permissions for watching all custom resources. Set to false for restricted RBAC with specific API groups |
| otel.k8sResourceCollector.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| otel.k8sResourceCollector.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| otel.k8sResourceCollector.readinessProbe.initialDelaySeconds | int | `5` | `initialDelaySeconds` for the readiness probe. |
| otel.k8sResourceCollector.readinessProbe.periodSeconds | int | `5` | `periodSeconds` for the readiness probe. |
| otel.k8sResourceCollector.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| otel.k8sResourceCollector.readinessProbe.timeoutSeconds | int | `5` | `timeoutSeconds` for the readiness probe. |
| otel.k8sResourceCollector.replicaCount | int | `2` | Number of cluster collector pods to schedule. Use 2+ with leaderElection for HA. |
| otel.k8sResourceCollector.resources.limits.cpu | string | `"500m"` | CPU resource limits. |
| otel.k8sResourceCollector.resources.limits.memory | string | `"512Mi"` | Memory resource limits. |
| otel.k8sResourceCollector.resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| otel.k8sResourceCollector.resources.requests.memory | string | `"128Mi"` | Memory resource requests. |
| otel.k8sResourceCollector.serviceaccount.annotations | object | `{}` | Annotations for the service account for the cluster collector pods |
| otel.k8sResourceCollector.skipSslValidation | bool | `false` | If true, ignores the server certificate being signed by an unknown authority. |
| otel.k8sResourceCollector.strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| otel.k8sResourceCollector.tolerations | list | `[]` | Toleration labels for pod assignment. |
| otel.platformGrpcOtlpEndpoint | string | `""` | Override the OTLP endpoint with a gRPC OTLP endpoint (format must be `host:port`, no scheme) platformHttpOtlpEndpoint takes precedence when both overrides are defined. When empty, derived by appending /otel to stackstate.url. |
| otel.platformHttpOtlpEndpoint | string | `""` | Override the OTLP endpoint with an HTTP(S) OTLP endpoint (format must be `http(s)://<host>:<port>`), takes precedence over the platformGrpcOtlpEndpoint when both overrides are defined. When empty, derived by appending /otel to stackstate.url. |
| otel.prometheusScraping.collector.affinity | object | `{}` | Affinity settings for pod assignment. |
| otel.prometheusScraping.collector.debug | object | `{"enabled":false,"pipelines":["metrics"],"verbosity":"basic"}` | Optional debug exporter for troubleshooting. When enabled, the upstream OTel `debug` exporter is wired into the listed pipelines so payloads are written to the collector log. Leave disabled in production. |
| otel.prometheusScraping.collector.debug.enabled | bool | `false` | Enable the debug exporter for this collector. |
| otel.prometheusScraping.collector.debug.pipelines | list | `["metrics"]` | Pipelines (by signal) to attach the debug exporter to. Must be a subset of {traces, logs, metrics}. |
| otel.prometheusScraping.collector.debug.verbosity | string | `"basic"` | Debug exporter verbosity: basic, normal, or detailed. |
| otel.prometheusScraping.collector.image.pullPolicy | string | `"IfNotPresent"` | Container image pull policy for the Prometheus scraper collector. |
| otel.prometheusScraping.collector.image.repository | string | `"stackstate/sts-opentelemetry-collector"` | Base container image repository for the Prometheus scraper collector. Shares the SUSE Observability collector image with the k8sResourceCollector component but keeps its own tag/pullPolicy so the two can be overridden independently. |
| otel.prometheusScraping.collector.image.tag | string | `"v0.0.53-agent"` | Container image tag for the Prometheus scraper collector. Uses the strict agent collector BOM image (the "-agent" suffixed tag). |
| otel.prometheusScraping.collector.nodeSelector | object | `{}` | Node labels for pod assignment. |
| otel.prometheusScraping.collector.podAnnotations | object | `{}` | Additional annotations for Prometheus scraper collector pods. |
| otel.prometheusScraping.collector.podLabels | object | `{}` | Additional labels for Prometheus scraper collector pods. |
| otel.prometheusScraping.collector.pprof.enabled | bool | `true` | Enable the pprof extension for profiling/debugging. Opt-out: enabled by default. The pprof endpoint is reachable inside the pod (port 1777) via kubectl port-forward. |
| otel.prometheusScraping.collector.priorityClassName | string | `nil` | Priority class for Prometheus scraper collector pods. |
| otel.prometheusScraping.collector.replicaCount | int | `1` | Number of Prometheus scraper collector pods to schedule. |
| otel.prometheusScraping.collector.resources.limits.cpu | string | `"500m"` | CPU resource limits. |
| otel.prometheusScraping.collector.resources.limits.memory | string | `"1Gi"` | Memory resource limits. A bit more headroom than k8sResourceCollector to absorb scrape batching from many targets. |
| otel.prometheusScraping.collector.resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| otel.prometheusScraping.collector.resources.requests.memory | string | `"256Mi"` | Memory resource requests. |
| otel.prometheusScraping.collector.tolerations | list | `[]` | Toleration labels for pod assignment. |
| otel.prometheusScraping.enabled | bool | `true` | Enable / disable OpenTelemetry Collector based Prometheus scraping via ServiceMonitor and PodMonitor resources. Requires otel.enabled=true. |
| otel.prometheusScraping.monitorCrds.enabled | bool | `false` | Install and upgrade ServiceMonitor and PodMonitor CRDs when Prometheus scraping is enabled. |
| otel.prometheusScraping.monitorCrds.keep | bool | `true` | Annotate the installed ServiceMonitor and PodMonitor CRDs with `helm.sh/resource-policy: keep` so they (and any custom resources users have created against them) survive a `helm uninstall` of this chart. Only takes effect when `monitorCrds.enabled` is true. |
| otel.prometheusScraping.skipSslValidation | bool | `false` | If true, ignores the server certificate being signed by an unknown authority when sending OTLP to the platform. |
| otel.prometheusScraping.targetAllocator.affinity | object | `{}` | Affinity settings for pod assignment. |
| otel.prometheusScraping.targetAllocator.allocationStrategy | string | `"consistent-hashing"` | Target Allocator strategy for distributing scrape targets across collectors. |
| otel.prometheusScraping.targetAllocator.allowInsecureAuthSecrets | bool | `false` | Allow the Target Allocator to serve ServiceMonitor and PodMonitor auth secrets to collectors over plain HTTP. When false (default), auth secrets are not served; enable mTLS with mtlsEnabled=true to serve them securely, or set to true to allow serving over plain HTTP. |
| otel.prometheusScraping.targetAllocator.filterStrategy | string | `"relabel-config"` | Target Allocator filtering strategy for generated scrape configs. |
| otel.prometheusScraping.targetAllocator.image.pullPolicy | string | `"IfNotPresent"` | Container image pull policy for the Target Allocator. |
| otel.prometheusScraping.targetAllocator.image.registry | string | `nil` | Override registry for the Target Allocator image. Defaults to global.imageRegistry. |
| otel.prometheusScraping.targetAllocator.image.repository | string | `"stackstate/opentelemetry-target-allocator"` | SUSE Observability Target Allocator image repository, rebuilt from the OpenTelemetry Operator source on SUSE BCI so we can patch Go-stdlib and golang.org/x/* CVEs without waiting for an upstream operator release. |
| otel.prometheusScraping.targetAllocator.image.tag | string | `"0.153.0-so6"` | SUSE Observability Target Allocator image tag (<upstream-version>-so<release-increment>). |
| otel.prometheusScraping.targetAllocator.mtlsEnabled | bool | `false` | Enable mTLS between scraper collectors and the Target Allocator. When true, credentials referenced by ServiceMonitors and PodMonitors are fetched over a mutually authenticated TLS connection. Requires cert-manager to be installed. See the README for details. |
| otel.prometheusScraping.targetAllocator.nodeSelector | object | `{}` | Node labels for pod assignment. |
| otel.prometheusScraping.targetAllocator.podAnnotations | object | `{}` | Additional annotations for Target Allocator pods. |
| otel.prometheusScraping.targetAllocator.podLabels | object | `{}` | Additional labels for Target Allocator pods. |
| otel.prometheusScraping.targetAllocator.priorityClassName | string | `nil` | Priority class for Target Allocator pods. |
| otel.prometheusScraping.targetAllocator.prometheusCR.allowNamespaces | list | `[]` | Namespaces where monitor resources are allowed. Mutually exclusive with denyNamespaces. |
| otel.prometheusScraping.targetAllocator.prometheusCR.denyFSAccessThroughSMs | bool | `true` | Drop monitor endpoints that reference arbitrary files on the collector filesystem. |
| otel.prometheusScraping.targetAllocator.prometheusCR.denyNamespaces | list | `[]` | Namespaces where monitor resources are denied. Mutually exclusive with allowNamespaces. |
| otel.prometheusScraping.targetAllocator.prometheusCR.podMonitorNamespaceSelector | object | `{}` | Labels or full LabelSelector selecting PodMonitor namespaces. |
| otel.prometheusScraping.targetAllocator.prometheusCR.podMonitorSelector | object | `{"observability.suse.com/agent":"scrape"}` | Labels or full LabelSelector selecting PodMonitor resources to scrape. |
| otel.prometheusScraping.targetAllocator.prometheusCR.secretNamespaces | list | `[]` | Namespaces where referenced monitor auth secrets can be read. |
| otel.prometheusScraping.targetAllocator.prometheusCR.serviceMonitorNamespaceSelector | object | `{}` | Labels or full LabelSelector selecting ServiceMonitor namespaces. |
| otel.prometheusScraping.targetAllocator.prometheusCR.serviceMonitorSelector | object | `{"observability.suse.com/agent":"scrape"}` | Labels or full LabelSelector selecting ServiceMonitor resources to scrape. |
| otel.prometheusScraping.targetAllocator.replicaCount | int | `1` | Number of Target Allocator pods to schedule. |
| otel.prometheusScraping.targetAllocator.resources.limits.cpu | string | `"200m"` | CPU resource limits. The allocator only watches monitor CRDs and distributes targets across collectors, so this can stay small. |
| otel.prometheusScraping.targetAllocator.resources.limits.memory | string | `"256Mi"` | Memory resource limits. |
| otel.prometheusScraping.targetAllocator.resources.requests.cpu | string | `"50m"` | CPU resource requests. |
| otel.prometheusScraping.targetAllocator.resources.requests.memory | string | `"64Mi"` | Memory resource requests. |
| otel.prometheusScraping.targetAllocator.tolerations | list | `[]` | Toleration labels for pod assignment. |
| otel.telemetryGateway.affinity | object | `{}` | Affinity settings for pod assignment. |
| otel.telemetryGateway.debug | object | `{"enabled":false,"pipelines":["traces","logs","metrics"],"verbosity":"basic"}` | Optional debug exporter for troubleshooting. When enabled, the upstream OTel `debug` exporter is wired into the listed pipelines so payloads are written to the collector log. Leave disabled in production. |
| otel.telemetryGateway.debug.enabled | bool | `false` | Enable the debug exporter for this collector. |
| otel.telemetryGateway.debug.pipelines | list | `["traces","logs","metrics"]` | Pipelines (by signal) to attach the debug exporter to. Must be a subset of {traces, logs, metrics}. |
| otel.telemetryGateway.debug.verbosity | string | `"basic"` | Debug exporter verbosity: basic, normal, or detailed. |
| otel.telemetryGateway.enabled | bool | `false` | Enable the telemetry gateway for OTLP push-based telemetry. Metrics and traces are forwarded to SUSE Observability; log signals are accepted for debug visibility only and written to the gateway pod logs. Requires otel.enabled=true. |
| otel.telemetryGateway.image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| otel.telemetryGateway.image.repository | string | `"stackstate/sts-opentelemetry-collector"` | Base container image repository. |
| otel.telemetryGateway.image.tag | string | `"v0.0.53-agent"` | Container image tag for the telemetry gateway. Uses the strict agent collector BOM image (the "-agent" suffixed tag). |
| otel.telemetryGateway.nodeSelector | object | `{}` | Node labels for pod assignment. |
| otel.telemetryGateway.podAnnotations | object | `{}` | Additional annotations for gateway pods. |
| otel.telemetryGateway.podLabels | object | `{}` | Additional labels for gateway pods. |
| otel.telemetryGateway.pprof.enabled | bool | `true` | Enable the pprof extension for profiling/debugging. Opt-out: enabled by default. The pprof endpoint is reachable inside the pod (port 1777) via kubectl port-forward. |
| otel.telemetryGateway.priorityClassName | string | `""` | Priority class for gateway pods. |
| otel.telemetryGateway.replicaCount | int | `1` | Number of gateway pods. 1 is sufficient for small/medium clusters. Use 2+ with a PDB for HA. |
| otel.telemetryGateway.resources.limits.cpu | string | `"1"` | CPU resource limits. |
| otel.telemetryGateway.resources.limits.memory | string | `"2Gi"` | Memory resource limits. |
| otel.telemetryGateway.resources.requests.cpu | string | `"250m"` | CPU resource requests. |
| otel.telemetryGateway.resources.requests.memory | string | `"1Gi"` | Memory resource requests. |
| otel.telemetryGateway.service.annotations | object | `{}` | Annotations for the ClusterIP Service. |
| otel.telemetryGateway.serviceaccount.annotations | object | `{}` | Annotations for the service account for the gateway pods. |
| otel.telemetryGateway.skipSslValidation | bool | `false` | Skip TLS validation when exporting to the platform. |
| otel.telemetryGateway.spanMetrics.aggregationCardinalityLimit | int | `5000` | Maximum number of unique span-metric aggregation series held by the gateway. Prevents high-cardinality spans from exhausting memory. |
| otel.telemetryGateway.strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| otel.telemetryGateway.tolerations | list | `[]` | Toleration labels for pod assignment. |
| otel.telemetryGateway.traceSampling.maxTotalSpansPerSecond | int | `500` | Maximum traces spans per second exported by the gateway. |
| processAgent.checkIntervals.connections | int | `30` | Override the default value of the connections check interval in seconds. |
| processAgent.checkIntervals.process | int | `32` | Override the default value of the process check interval in seconds. |
| processAgent.disabledProtocols | list | `[]` | List of protocols to disable for protocol inspection. Supported protocols are http, http2, mongo, amqp, postgres, tls. If nothing is provided all protocols will be enabled. |
| processAgent.podCorrelation.attributes | list | `[]` | The attributes to be added to all exported metrics. If nothing is provided a default set will be used. |
| processAgent.podCorrelation.enabled | bool | `false` | [Experimental] Enable / disable pod correlation. |
| processAgent.podCorrelation.exporter.endpoint | string | `""` | Override the default endpoint to which the exporter will send metrics (e.g. "otel-collector-service:4317"). |
| processAgent.podCorrelation.exporter.interval | int | `30` | The interval at which the exporter will send metrics (in seconds). |
| processAgent.podCorrelation.exporter.type | string | `""` | The type of the exporter to use for metrics. Possible values ("otlp", "stdout") |
| processAgent.podCorrelation.partialCorrelation | bool | `false` | Enable / disable partial pod correlation. if false the agent will export only pod<->pod metrics. Both extremities of the connection must be pods. |
| processAgent.podCorrelation.protocolMetrics | bool | `false` | Enable / disable exporting protocol metrics. If false the agent will export only metrics for network connections. |
| processAgent.podCorrelation.remoteCache | bool | `false` | When true, the chart will deploy a remote kube cache service and populate the process-agent environment with the service address (serviceName.namespace:grpcPort). Set to `false` to use a local informer inside the process-agent and avoid deploying the remote cache. |
| processAgent.softMemoryLimit.goMemLimit | string | `"340MiB"` | Soft-limit for golang heap allocation, for sanity, must be around 85% of nodeAgent.containers.processAgent.resources.limits.cpu. |
| processAgent.softMemoryLimit.httpObservationsBufferSize | int | `40000` | Sets a maximum for the number of http observations to keep in memory between check runs, to use 40k requires around ~400Mib of memory. |
| processAgent.softMemoryLimit.httpStatsBufferSize | int | `40000` | Sets a maximum for the number of http stats to keep in memory between check runs, to use 40k requires around ~400Mib of memory. |
| remoteKubeCache.affinity | object | `{}` | Affinity settings for pod assignment. |
| remoteKubeCache.nodeSelector | object | `{}` | Node labels for pod assignment. |
| remoteKubeCache.resources.limits.cpu | string | `"200m"` | CPU resource limits. |
| remoteKubeCache.resources.limits.memory | string | `"400Mi"` | Memory resource limits. |
| remoteKubeCache.resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| remoteKubeCache.resources.requests.memory | string | `"200Mi"` | Memory resource requests. |
| remoteKubeCache.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.apiKey | string | `nil` | **PROVIDE YOUR API KEY HERE** API key to be used by the agent. |
| stackstate.cluster.authToken | string | `""` | Provide a token to enable secure communication between the agent and the cluster agent. |
| stackstate.cluster.name | string | `nil` | **PROVIDE KUBERNETES CLUSTER NAME HERE** Name of the Kubernetes cluster where the agent will be installed. |
| stackstate.customApiKeySecretKey | string | `"sts-api-key"` | Key in the secret containing the receiver API key. |
| stackstate.customClusterAuthTokenSecretKey | string | `"sts-cluster-auth-token"` | Key in the secret containing the cluster auth token. |
| stackstate.customSecretName | string | `""` | Name of the secret containing the receiver API key. |
| stackstate.manageOwnSecrets | bool | `false` | Set to true if you don't want this helm chart to create secrets for you. |
| stackstate.url | string | `nil` | **PROVIDE STACKSTATE URL HERE** URL of the StackState installation to receive data from the agent. |
| targetSystem | string | `"linux"` | Target OS for this deployment (possible values: linux) |
