stackstate-standalone
=====================
Helm chart for StackState standlone -- all components running inside a single container.

This chart's source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)


## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | {} | Affinity settings for pod assignment. |
| fullnameOverride | string | "" | Override the fullname of the chart. |
| image.pullPolicy | string | "IfNotPresent" | Default container image pull policy. |
| image.repository | string | "508573134510.dkr.ecr.eu-west-1.amazonaws.com/stackstate" | Base container image registry. |
| image.tag | string | "master" | Default container image tag. |
| imagePullSecrets | list | [] | Extra secrets / credentials needed for container image registry. |
| ingress.annotations | object | {} | Annotations for ingress objects. |
| ingress.enabled | bool | false | Enable use of ingress controllers. |
| ingress.hosts | list | [] | List of ingress hostnames |
| ingress.tls | list | [] | List of ingress TLS certificates to use. |
| nameOverride | string | "" | Override the name of the chart. |
| nodeSelector | object | {} | Node labels for pod assignment. |
| resources.limits.cpu | int | 2 | CPU resource limits. |
| resources.limits.memory | string | "8Gi" | Memory resource limits. |
| resources.requests.cpu | int | 1 | CPU resource requests. |
| resources.requests.memory | string | "2Gi" | Memory resource requests. |
| service.administration.port | int | 7071 | The default port for the StackState Administration area. |
| service.receiver.port | int | 7077 | The default port for the StackState Receiver. |
| service.type | string | "ClusterIP" | The Kubernetes 'Service' type to use. |
| service.ui.port | int | 7070 | The default port for the StackState UI. |
| stackstate.license.key | string | \<nil\> | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. |
| stackstate.receiver.apiKey | string | \<nil\> | **PROVIDE YOUR API KEY HERE** API key to be used by all StackState agents. |
| stackstate.receiver.baseUrl | string | \<nil\> | **PROVIDE YOUR BASEURL HERE** Externally visible baseUrl of the StackState endpoints. |
| tolerations | list | [] | Toleration labels for pod assignment. |
