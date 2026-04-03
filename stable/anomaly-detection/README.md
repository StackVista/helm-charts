# anomaly-detection

A Helm chart for Anomaly Detection

Current chart version is `5.2.0-snapshot.180`

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.stackstate.io | common | 0.4.17 |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `stackstate.instance`
* `global.receiverApiKey`

Install them on the command line on Helm with the following command:

```
helm install \
  --set stackstate.instance=<instance> \
  --set global.receiverApiKey=<API-KEY> \
  --values ./values.yaml \
  stable/anomaly-detection
```

The overriding default etc/* configuration is possible using property overrides below:
* `etcoverride.task_logging_conf`
* `etcoverride.anomaly_detect_yaml`
* `etcoverride.anomaly_models.yaml`
* `etcoverride.anomaly_train_yaml`
* `etcoverride.spotlight_yaml`
* `etcoverride.streams_stsl`
* `etcoverride.views_stsl`

See example, below:
```
helm template . --values values.yaml --set-file etcoverride.streams_stsl=etc/streams.stsl
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| cluster-role.enabled | bool | `true` | Deploy the ClusterRoleBinding(s) together with the chart. Can be disabled if these need to be installed by an administrator of the Kubernetes cluster. |
| cpu.limit | int | `4` | CPU resource limit |
| cpu.request | int | `4` | CPU resource request |
| enabled | bool | `true` | Enables anomaly detection chart |
| ephemeralStorage.limit | string | `"2Gi"` | Ephemeral storage resource limit |
| ephemeralStorage.request | string | `"1Mi"` | Ephemeral storage resource request |
| global.commonLabels | object | `{}` | Common labels to be applied to Deployments and their pods. |
| global.receiverApiKey | string | `nil` | **Required API key used by the Receiver. |
| image.imagePullPolicy | string | `"Always"` | The default pullPolicy used for anomaly detection pods. |
| image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| image.pullSecretPassword | string | `nil` |  |
| image.pullSecretUsername | string | `nil` | Password used to login to the registry to pull Docker images of all pods. |
| image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| image.spotlightRepository | string | `"stackstate/spotlight"` | Repository of the spotlight Docker image. |
| image.tag | string | `"5.0.0"` | the chart image tag, e.g. 4.1.0-latest |
| ingress | object | `{"annotations":{},"enabled":false,"hostname":null,"hosts":[],"port":8090,"tls":null}` | Status interface ingress |
| ingress.enabled | bool | `false` | Enables ingress controller for status interface |
| ingress.hostname | string | `nil` | Status interface hostname e.g. spotlight.local.domain |
| manager.affinity | object | `{}` |  |
| manager.cpu.limit | int | `1` |  |
| manager.cpu.request | int | `1` |  |
| manager.ephemeralStorage.limit | string | `"2Gi"` |  |
| manager.ephemeralStorage.request | string | `"1Mi"` |  |
| manager.memory.limit | string | `"1Gi"` |  |
| manager.memory.request | string | `"1Gi"` |  |
| manager.nodeSelector | object | `{}` |  |
| manager.persistentStorage.size | string | `"10Gi"` |  |
| manager.persistentStorage.storageClass | string | `nil` |  |
| manager.tolerations | list | `[]` |  |
| memory.limit | string | `"6Gi"` | Memory resource limit |
| memory.request | string | `"6Gi"` | Memory resource request |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| pdb.maxUnavailable | int | `0` | PodDisruptionBudget settings for `anomaly-detection` pods. |
| replicas | int | `1` |  |
| securityContext | object | `{"enabled":true,"fsGroup":65534,"runAsGroup":65534,"runAsNonRoot":true,"runAsUser":65534}` | Pod Security Context |
| securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| securityContext.fsGroup | int | `65534` | The GID (group ID) used to mount volumes |
| securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| stackstate.apiToken | string | `nil` | Stackstate Api token that used by spotlight for authentication, it is expected to be set only in case if authType = "api-token" |
| stackstate.authType | string | `"token"` | Type of authentication. There are three options 1) "token" - with service account token (default), 2) "api-token" - with Stackstate API Token, 3) "cookie" - username, password based auth. |
| stackstate.instance | string | `nil` | **Required Stackstate instance URL, e.g http://stackstate-headless:7070 |
| stackstate.password | string | `nil` | Stackstate Password that used by spotlight for authentication, it is expected to be set only in case if authType = "cookie" |
| stackstate.username | string | `nil` | Stackstate Username that used by spotlight for authentication, it is expected to be set only in case if authType = "cookie" |
| strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| threadWorkers | int | `8` | The number of worker threads. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
