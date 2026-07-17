# keycloak

![Version: 0.1.5](https://img.shields.io/badge/Version-0.1.5-informational?style=flat-square) ![AppVersion: 26.7.0](https://img.shields.io/badge/AppVersion-26.7.0-informational?style=flat-square)

Keycloak is a high performance Java-based identity and access management solution, packaged for SUSE Observability on top of the upstream Keycloak image.

**Homepage:** <https://github.com/StackVista/helm-charts-internal>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| SUSE Observability |  | <https://github.com/StackVista> |

## Source Code

* <https://github.com/StackVista/helm-charts-internal/tree/main/stable/keycloak>
* <https://www.keycloak.org>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| oci://registry-1.docker.io/bitnamicharts | common | 2.24.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for pod assignment (overrides the presets above) |
| args | list | `[]` | Override the default container args. Defaults to ["start", "--optimized"] when empty. |
| auth.adminPassword | string | `""` | Bootstrap admin password (maps to KC_BOOTSTRAP_ADMIN_PASSWORD). Auto-generated when empty and no existingSecret is set. |
| auth.adminUser | string | `"admin"` | Bootstrap admin username (maps to KC_BOOTSTRAP_ADMIN_USERNAME) |
| auth.annotations | object | `{}` | Additional annotations for the chart-managed admin secret |
| auth.existingSecret | string | `""` | Name of an existing secret containing the admin password. When set, adminPassword is ignored. |
| auth.passwordSecretKey | string | `""` | Key inside auth.existingSecret holding the admin password (defaults to "admin-password") |
| automountServiceAccountToken | bool | `true` | Mount the service account token in the pod (needed for KUBE_PING/DNS discovery RBAC) |
| autoscaling.behavior.scaleDown.policies | list | `[{"periodSeconds":300,"type":"Pods","value":1}]` | Scaling policies when scaling down |
| autoscaling.behavior.scaleDown.selectPolicy | string | `"Max"` | Policy selection when scaling down |
| autoscaling.behavior.scaleDown.stabilizationWindowSeconds | int | `300` | Stabilization window when scaling down |
| autoscaling.behavior.scaleUp.policies | list | `[]` | Scaling policies when scaling up |
| autoscaling.behavior.scaleUp.selectPolicy | string | `"Max"` | Policy selection when scaling up |
| autoscaling.behavior.scaleUp.stabilizationWindowSeconds | int | `120` | Stabilization window when scaling up |
| autoscaling.enabled | bool | `false` | Enable the HorizontalPodAutoscaler |
| autoscaling.maxReplicas | int | `11` | Maximum number of replicas |
| autoscaling.minReplicas | int | `1` | Minimum number of replicas |
| autoscaling.targetCPU | string | `""` | Target CPU utilization percentage |
| autoscaling.targetMemory | string | `""` | Target memory utilization percentage |
| cache.authOwnersCount | string | `""` | Number of owners for the auth/loginFailures/actionTokens caches (defaults to ownersCount) |
| cache.configuration | string | `""` | Inline cache-ispn.xml content. Overrides the generated file when set. |
| cache.enabled | bool | `true` | Enable the distributed Infinispan cache (KC_CACHE=ispn) for clustering. Set false for a single-node local cache. |
| cache.existingConfigmap | string | `""` | Name of an existing ConfigMap providing cache-ispn.xml. Overrides configuration when set. |
| cache.ownersCount | int | `2` | Number of owners for the distributed session caches (rendered into cache-ispn.xml) |
| cache.stackName | string | `"kubernetes"` | JGroups stack for discovery (KC_CACHE_STACK). "kubernetes" uses DNS_PING against the headless service. |
| clusterDomain | string | `"cluster.local"` | Default Kubernetes cluster domain (used to build the JGroups DNS query) |
| command | list | `[]` | Override the default container command (image ENTRYPOINT is kc.sh; leave empty to use it) |
| commonAnnotations | object | `{}` | Annotations to add to all deployed objects |
| commonLabels | object | `{}` | Labels to add to all deployed objects |
| containerPorts.http | int | `8080` | Keycloak HTTP container port (KC_HTTP_PORT) |
| containerPorts.management | int | `9000` | Keycloak management port for metrics/health (KC_HTTP_MANAGEMENT_PORT) |
| containerSecurityContext.allowPrivilegeEscalation | bool | `false` | Allow privilege escalation |
| containerSecurityContext.capabilities.drop | list | `["ALL"]` | Linux capabilities to drop |
| containerSecurityContext.enabled | bool | `true` | Enable the container security context |
| containerSecurityContext.privileged | bool | `false` | Run in privileged mode |
| containerSecurityContext.readOnlyRootFilesystem | bool | `true` | Mount the root filesystem read-only (conf and data are writable emptyDir mounts) |
| containerSecurityContext.runAsGroup | int | `0` | runAsGroup (gid 0 / root group, as the image expects) |
| containerSecurityContext.runAsNonRoot | bool | `true` | Run as a non-root user |
| containerSecurityContext.runAsUser | int | `1000` | runAsUser. The image runs as uid 1000. |
| containerSecurityContext.seLinuxOptions | object | `{}` | SELinux options |
| containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` | Seccomp profile type |
| customLivenessProbe | object | `{}` | Custom liveness probe overriding the default |
| customReadinessProbe | object | `{}` | Custom readiness probe overriding the default |
| customStartupProbe | object | `{}` | Custom startup probe overriding the default |
| diagnosticMode.args | list | `["infinity"]` | Args override applied to the container in diagnostic mode |
| diagnosticMode.command | list | `["sleep"]` | Command override applied to the container in diagnostic mode |
| diagnosticMode.enabled | bool | `false` | Enable diagnostic mode (all probes disabled and the container command is overridden) |
| dnsConfig | object | `{}` | DNS Configuration for pod |
| dnsPolicy | string | `""` | DNS Policy for pod |
| enableServiceLinks | bool | `true` | If set to false, disable Kubernetes service links in the pod spec |
| externalDatabase.annotations | object | `{}` | Additional annotations for the chart-managed external database secret |
| externalDatabase.database | string | `"keycloak"` | Database name (KC_DB_URL_DATABASE) |
| externalDatabase.existingSecret | string | `""` | Name of an existing secret holding database credentials |
| externalDatabase.existingSecretDatabaseKey | string | `""` | Key in existingSecret holding the database name |
| externalDatabase.existingSecretHostKey | string | `""` | Key in existingSecret holding the database host (sources KC_DB_URL_HOST from the secret when set) |
| externalDatabase.existingSecretPasswordKey | string | `""` | Key in existingSecret holding the database password (defaults to "db-password") |
| externalDatabase.existingSecretPortKey | string | `""` | Key in existingSecret holding the database port |
| externalDatabase.existingSecretUserKey | string | `""` | Key in existingSecret holding the database user |
| externalDatabase.host | string | `""` | Database host (KC_DB_URL_HOST). Evaluated as a template. |
| externalDatabase.password | string | `""` | Database password (KC_DB_PASSWORD). Stored in a chart-managed secret when no existingSecret is set. |
| externalDatabase.port | int | `5432` | Database port (KC_DB_URL_PORT) |
| externalDatabase.type | string | `"postgres"` | Database vendor (KC_DB). The image is build-optimized for postgres; do not change without rebuilding the image. |
| externalDatabase.user | string | `"keycloak"` | Database username (KC_DB_USERNAME) |
| extraContainerPorts | list | `[]` | Extra ports to expose on the Keycloak container |
| extraDeploy | list | `[]` | Array of extra objects to deploy with the release (rendered as templates). A consumer injects an extra Secret via this. |
| extraEnvVars | list | `[]` | Extra environment variables for the Keycloak container (list of name/value). The StackState image reads KEYCLOAK_BASE_URL and USERCREATE_EVENT_LISTENER_CLIENT_NAME_PREFIX from here. |
| extraEnvVarsCM | string | `""` | Name of an existing ConfigMap with extra env vars (mounted via envFrom) |
| extraEnvVarsConfigMapData | object | `{}` | Extra key/value pairs merged into the chart-managed env-vars ConfigMap |
| extraEnvVarsSecret | string | `""` | Name of an existing Secret with extra env vars (mounted via envFrom) |
| extraVolumeMounts | list | `[]` | Extra volume mounts for the Keycloak container |
| extraVolumes | list | `[]` | Extra volumes for the Keycloak pods |
| fullnameOverride | string | `""` | String to fully override common.names.fullname |
| global.compatibility.openshift.adaptSecurityContext | string | `"auto"` | Adapt securityContext to be compatible with Openshift restricted-v2 SCC. Values: auto, force, disabled |
| global.defaultStorageClass | string | `""` | Global default StorageClass for Persistent Volume(s) |
| global.imagePullSecrets | list | `[]` | Global Docker registry secret names as an array |
| global.imageRegistry | string | `""` | Global Docker image registry (overrides image.registry when set) |
| global.storageClass | string | `""` | DEPRECATED: use global.defaultStorageClass instead |
| hostAliases | list | `[]` | Pod host aliases |
| hostname | string | `""` | Public hostname/URL of Keycloak (KC_HOSTNAME). Formerly KEYCLOAK_BASE_URL. Evaluated as a template. |
| hostnameAdmin | string | `""` | Separate admin console hostname/URL (KC_HOSTNAME_ADMIN). Evaluated as a template. Leave empty to reuse hostname. |
| hostnameStrict | bool | `false` | Disable dynamic hostname resolution from request headers (KC_HOSTNAME_STRICT) |
| httpEnabled | bool | `true` | Serve plain HTTP (KC_HTTP_ENABLED). Required behind a TLS-terminating proxy edge. |
| httpRelativePath | string | `"/"` | Path relative to '/' for serving resources (KC_HTTP_RELATIVE_PATH) |
| image.digest | string | `""` | Keycloak image digest in the way sha256:aa.... Overrides the tag when set |
| image.pullPolicy | string | `"IfNotPresent"` | Keycloak image pull policy |
| image.pullSecrets | list | `["registry-credentials"]` | Docker registry secret names as an array (e.g. registry-credentials) |
| image.registry | string | `"quay.io"` | Keycloak image registry |
| image.repository | string | `"stackstate/keycloak"` | Keycloak image repository (StackState build FROM the upstream/Rancher Keycloak image) |
| image.tag | string | `"26.7.0-main-af703171"` | Keycloak image tag |
| ingress.annotations | object | `{}` | Additional annotations for the Ingress (e.g. cert-manager or a Traefik middleware annotation set by the consumer) |
| ingress.apiVersion | string | `""` | Force Ingress API version (auto-detected if not set) |
| ingress.controller | string | `"default"` | Ingress controller type. Values: default or gce |
| ingress.enabled | bool | `false` | Enable ingress record generation for Keycloak |
| ingress.extraHosts | list | `[]` | Additional hostnames to cover with the ingress record |
| ingress.extraPaths | list | `[]` | Additional arbitrary paths added under the main host |
| ingress.extraRules | list | `[]` | Additional rules for the ingress record |
| ingress.extraTls | list | `[]` | TLS configuration for additional hostnames |
| ingress.hostname | string | `"keycloak.local"` | Default host for the ingress record (evaluated as a template) |
| ingress.ingressClassName | string | `""` | IngressClass used to implement the Ingress |
| ingress.labels | object | `{}` | Additional labels for the Ingress resource |
| ingress.path | string | `"{{ .Values.httpRelativePath }}"` | Default path for the ingress record (evaluated as a template) |
| ingress.pathType | string | `"ImplementationSpecific"` | Ingress path type |
| ingress.secrets | list | `[]` | Custom certificates provided as secrets |
| ingress.selfSigned | bool | `false` | Create a self-signed TLS secret for this ingress via Helm |
| ingress.servicePort | string | `"http"` | Backend service port name to use (http) |
| ingress.tls | bool | `false` | Enable TLS for ingress.hostname (TLS terminates at the ingress) |
| initContainers | list | `[]` | Additional init containers for the Keycloak pods |
| kubeVersion | string | `""` | Force target Kubernetes version (using Helm capabilities if not set) |
| lifecycleHooks | object | `{}` | Container lifecycle hooks |
| livenessProbe.enabled | bool | `true` | Enable the liveness probe (GET /health/live on the management port) |
| livenessProbe.failureThreshold | int | `3` | Failure threshold for the liveness probe |
| livenessProbe.initialDelaySeconds | int | `60` | Initial delay for the liveness probe |
| livenessProbe.periodSeconds | int | `10` | Period for the liveness probe |
| livenessProbe.successThreshold | int | `1` | Success threshold for the liveness probe |
| livenessProbe.timeoutSeconds | int | `5` | Timeout for the liveness probe |
| logging.level | string | `"INFO"` | Root log level (KC_LOG_LEVEL): FATAL, ERROR, WARN, INFO, DEBUG, TRACE, ALL, OFF |
| logging.output | string | `"default"` | Console log format (KC_LOG_CONSOLE_OUTPUT): default or json |
| metrics.enabled | bool | `false` | Enable Keycloak metrics (KC_METRICS_ENABLED) on the management port |
| metrics.healthEnabled | bool | `true` | Enable Keycloak health endpoints (KC_HEALTH_ENABLED) on the management port. Required by the probes. |
| metrics.prometheusRule.enabled | bool | `false` | Create a PrometheusRule resource |
| metrics.prometheusRule.groups | list | `[]` | Alert rule groups |
| metrics.prometheusRule.labels | object | `{}` | Additional labels so Prometheus discovers the PrometheusRule |
| metrics.prometheusRule.namespace | string | `""` | Namespace where the PrometheusRule is created |
| metrics.service.annotations | object | `{"prometheus.io/port":"{{ .Values.metrics.service.ports.management }}","prometheus.io/scrape":"true"}` | Annotations for the metrics service (enables prometheus scraping) |
| metrics.service.extraPorts | list | `[]` | Extra ports for the metrics service |
| metrics.service.ports.management | int | `9000` | Metrics service port targeting the management container port (9000) |
| metrics.serviceMonitor.enabled | bool | `false` | Create a ServiceMonitor for the Prometheus Operator |
| metrics.serviceMonitor.endpoints | list | `[{"path":"/metrics"},{"path":"/health"}]` | Scrape endpoints. Defaults scrape /metrics and /health on the management port. |
| metrics.serviceMonitor.honorLabels | bool | `false` | Honor target labels on collisions |
| metrics.serviceMonitor.interval | string | `"30s"` | Scrape interval |
| metrics.serviceMonitor.jobLabel | string | `""` | Service label to use as the job name in Prometheus |
| metrics.serviceMonitor.labels | object | `{}` | Additional labels so Prometheus discovers the ServiceMonitor |
| metrics.serviceMonitor.metricRelabelings | list | `[]` | MetricRelabelConfigs applied before ingestion |
| metrics.serviceMonitor.namespace | string | `""` | Namespace where the ServiceMonitor is created |
| metrics.serviceMonitor.path | string | `""` | Single scrape path. Deprecated: use endpoints instead. |
| metrics.serviceMonitor.port | string | `"management"` | Service port name the ServiceMonitor scrapes (management) |
| metrics.serviceMonitor.relabelings | list | `[]` | RelabelConfigs applied before scraping |
| metrics.serviceMonitor.scrapeTimeout | string | `""` | Scrape timeout |
| metrics.serviceMonitor.selector | object | `{}` | Prometheus instance selector labels |
| minReadySeconds | int | `0` | Seconds a pod must be ready before killing the next during an update |
| nameOverride | string | `""` | String to partially override common.names.fullname |
| namespaceOverride | string | `""` | String to fully override common.names.namespace |
| networkPolicy.allowExternal | bool | `true` | When false, only pods with the matching client label may reach Keycloak |
| networkPolicy.allowExternalEgress | bool | `true` | Allow the pod to reach any destination |
| networkPolicy.enabled | bool | `true` | Create a NetworkPolicy for Keycloak |
| networkPolicy.extraEgress | list | `[]` | Extra egress rules for the NetworkPolicy |
| networkPolicy.extraIngress | list | `[]` | Extra ingress rules for the NetworkPolicy |
| networkPolicy.ingressNSMatchLabels | object | `{}` | Namespace labels allowed to reach Keycloak |
| networkPolicy.ingressNSPodMatchLabels | object | `{}` | Pod labels (in allowed namespaces) allowed to reach Keycloak |
| networkPolicy.kubeAPIServerPorts | list | `[443,6443,8443]` | Possible kube-apiserver endpoints for restricted egress |
| nodeAffinityPreset.key | string | `""` | Node label key to match. Ignored if affinity is set. |
| nodeAffinityPreset.type | string | `""` | Node affinity preset type. Ignored if affinity is set. Values: soft or hard |
| nodeAffinityPreset.values | list | `[]` | Node label values to match. Ignored if affinity is set. |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| pdb.create | bool | `true` | Create a PodDisruptionBudget |
| pdb.maxUnavailable | string | `""` | Maximum number/percentage of pods that may be unavailable |
| pdb.minAvailable | string | `""` | Minimum number/percentage of pods that must remain available |
| podAffinityPreset | string | `""` | Pod affinity preset. Ignored if affinity is set. Values: soft or hard |
| podAnnotations | object | `{}` | Annotations for Keycloak pods |
| podAntiAffinityPreset | string | `"soft"` | Pod anti-affinity preset. Ignored if affinity is set. Values: soft or hard |
| podLabels | object | `{}` | Extra labels for Keycloak pods |
| podManagementPolicy | string | `"Parallel"` | Pod management policy for the StatefulSet |
| podSecurityContext.enabled | bool | `true` | Enable the pods' security context |
| podSecurityContext.fsGroup | int | `0` | Pod fsGroup. The image runs as gid 0. |
| podSecurityContext.fsGroupChangePolicy | string | `"Always"` | Filesystem group change policy |
| podSecurityContext.supplementalGroups | list | `[]` | Extra filesystem groups |
| podSecurityContext.sysctls | list | `[]` | Kernel settings using the sysctl interface |
| priorityClassName | string | `""` | Priority class name for Keycloak pods |
| proxyHeaders | string | `"xforwarded"` | Reverse-proxy header mode (KC_PROXY_HEADERS). Use "xforwarded" for an edge proxy; leave empty to unset. |
| rbac.create | bool | `true` | Create the Role/RoleBinding granting pod get/list for JGroups KUBE_PING/DNS discovery |
| rbac.rules | list | `[]` | Custom RBAC rules appended to the Role |
| readinessProbe.enabled | bool | `true` | Enable the readiness probe (GET /health/ready on the management port) |
| readinessProbe.failureThreshold | int | `3` | Failure threshold for the readiness probe |
| readinessProbe.initialDelaySeconds | int | `30` | Initial delay for the readiness probe |
| readinessProbe.periodSeconds | int | `10` | Period for the readiness probe |
| readinessProbe.successThreshold | int | `1` | Success threshold for the readiness probe |
| readinessProbe.timeoutSeconds | int | `1` | Timeout for the readiness probe |
| replicaCount | int | `1` | Number of Keycloak replicas to deploy |
| resources | object | `{}` | Explicit container requests/limits (recommended for production) |
| resourcesPreset | string | `"small"` | Container resource preset (none, nano, micro, small, medium, large, xlarge, 2xlarge). Ignored when resources is set. |
| revisionHistoryLimitCount | int | `10` | Number of controller revisions to keep |
| schedulerName | string | `""` | Alternate scheduler name |
| service.annotations | object | `{}` | Additional annotations for the Keycloak service |
| service.clusterIP | string | `""` | Static clusterIP for the service |
| service.externalTrafficPolicy | string | `"Cluster"` | External traffic policy for NodePort/LoadBalancer |
| service.extraPorts | list | `[]` | Extra ports to expose on the Keycloak service |
| service.headless.annotations | object | `{}` | Annotations for the headless service |
| service.headless.extraPorts | list | `[]` | Extra ports to expose on the headless service |
| service.loadBalancerIP | string | `""` | loadBalancerIP for the service (cloud specific) |
| service.loadBalancerSourceRanges | list | `[]` | Addresses allowed when service is LoadBalancer |
| service.nodePorts.http | string | `""` | nodePort for the HTTP port when service type is NodePort/LoadBalancer |
| service.ports.http | int | `80` | Keycloak service HTTP port |
| service.sessionAffinity | string | `"None"` | Session affinity. Values: ClientIP or None |
| service.sessionAffinityConfig | object | `{}` | Additional settings for sessionAffinity |
| service.type | string | `"ClusterIP"` | Kubernetes service type |
| serviceAccount.annotations | object | `{}` | Additional annotations for the ServiceAccount |
| serviceAccount.automountServiceAccountToken | bool | `false` | Auto-mount the service account token on the ServiceAccount object |
| serviceAccount.create | bool | `true` | Create a ServiceAccount for Keycloak pods |
| serviceAccount.extraLabels | object | `{}` | Additional labels for the ServiceAccount |
| serviceAccount.name | string | `""` | Name of the created ServiceAccount (generated when empty) |
| sidecars | list | `[]` | Additional sidecar containers for the Keycloak pods |
| startupProbe.enabled | bool | `false` | Enable the startup probe (GET /health/ready on the management port) |
| startupProbe.failureThreshold | int | `60` | Failure threshold for the startup probe |
| startupProbe.initialDelaySeconds | int | `30` | Initial delay for the startup probe |
| startupProbe.periodSeconds | int | `5` | Period for the startup probe |
| startupProbe.successThreshold | int | `1` | Success threshold for the startup probe |
| startupProbe.timeoutSeconds | int | `1` | Timeout for the startup probe |
| statefulsetAnnotations | object | `{}` | Extra annotations for the StatefulSet resource |
| terminationGracePeriodSeconds | string | `""` | Seconds the pod needs to terminate gracefully |
| tolerations | list | `[]` | Tolerations for pod assignment |
| topologySpreadConstraints | list | `[]` | Topology spread constraints for pod assignment |
| updateStrategy.rollingUpdate | object | `{}` | StatefulSet rolling update configuration |
| updateStrategy.type | string | `"RollingUpdate"` | StatefulSet update strategy type |

