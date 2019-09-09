common
======
Common chartbuilding components and helpers

Current chart version is `0.1.1`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts)



## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| configmap.annotations | object | `{}` | Annotations for `ConfigMap` objects. |
| configmap.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `ConfigMap` object. |
| container.livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| container.livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| container.livenessProbe.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the liveness probe. |
| container.livenessProbe.periodSeconds | int | `10` | `periodSeconds` for the liveness probe. |
| container.livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| container.livenessProbe.timeoutSeconds | int | `2` | `timeoutSeconds` for the liveness probe. |
| container.readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| container.readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| container.readinessProbe.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the readiness probe. |
| container.readinessProbe.periodSeconds | int | `10` | `periodSeconds` for the readiness probe. |
| container.readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| container.readinessProbe.timeoutSeconds | int | `2` | `timeoutSeconds` for the readiness probe. |
| container.resources | object | `{}` | Container resource requests / limits. |
| deployment.affinity | object | `{}` | Affinity settings for pod assignment. |
| deployment.annotations | object | `{}` | Annotations for `Deployment` objects. |
| deployment.antiAffinity.strategy | string | `""` | Spread pods using simple `antiAffinity`; valid values are `hard` or `soft`. |
| deployment.antiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | The `topologyKey` to use for simple `antiAffinity` rule. |
| deployment.apiVersion | string | `"apps/v1"` | Kubernetes apiVersion to use with a `Deployment` object. |
| deployment.nodeSelector | object | `{}` | Node labels for pod assignment. |
| deployment.replicaCount | int | `1` | Amount of replicas to create for the `Deployment` object. |
| deployment.tolerations | list | `[]` | Toleration labels for pod assignment. |
| extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| gitlab.app | string | `""` | If CI is GitLab, specify the `app` for annotations. |
| gitlab.env | string | `""` | If CI is GitLab, specify the `env` for annotations. |
| global | object | `{}` |  |
| image.pullPolicy | string | `"Always"` | The pull policy for the Docker image. |
| image.repository | string | `"nginx"` | (string) Repository of the Docker image. |
| image.tag | string | `"latest"` | (string) Tag of the Docker image. |
| ingress.annotations | object | `{}` | Annotations for `Ingress` objects. |
| ingress.apiVersion | string | `"extensions/v1beta1"` | Kubernetes apiVersion to use with an `Ingress` object. |
| ingress.hosts | list | `[]` | List of ingress hostnames. |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| persistentvolumeclaim.annotations | object | `{}` | Annotations for `PersistentVolumeClaim` objects. |
| persistentvolumeclaim.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `PersistentVolumeClaim` object. |
| pod.annotations | object | `{}` | Annotations for `Pod` objects. |
| secret.annotations | object | `{}` | Annotations for `Secret` objects. |
| secret.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `Secret` object. |
| service.annotations | object | `{}` | Annotations for `Service` objects. |
| service.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `Service` object. |
| service.clusterIP | string | `""` |  |
| service.externalIPs | list | `[]` | List of external IP addresses to map to the `Service` object. |
| service.externalTrafficPolicy | string | `"Cluster"` | Denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints; only enabled when `service.type` is set to `LoadBalancer` or `NodePort`. |
| service.loadBalancerIP | string | `""` | Specify the source IP range to allow traffic from; only enabled when `service.type` is set to `LoadBalancer`. |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.ports | list | `[]` | List of ports to apply to the `Service` object. |
| service.type | string | `"ClusterIP"` | Kubernetes 'Service' type to use. |
