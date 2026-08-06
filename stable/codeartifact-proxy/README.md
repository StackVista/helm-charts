# codeartifact-proxy

![Version: 0.1.4](https://img.shields.io/badge/Version-0.1.4-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.29.0](https://img.shields.io/badge/AppVersion-1.29.0-informational?style=flat-square)

A Helm chart for CodeArtifact Proxy - Nginx-based caching proxy for AWS CodeArtifact Maven/PyPI format endpoints

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchLabels | object | `{}` |  |
| affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey | string | `"kubernetes.io/hostname"` |  |
| affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight | int | `1` |  |
| auth.htpasswd | string | `"test:$apr1$lpJpERMg$rjNCT9/GOewTido8UmQYD.\n"` | Client-facing HTTP Basic Auth credentials for this proxy (PACKAGE_REGISTRY_PROXY_USER/TOKEN). Unrelated to the CodeArtifact authorization token, which never enters Helm/Argo/SOPS. The default below is a chart-test-only credential; the real tooling-main deployment supplies a SOPS-encrypted copy of the current gitlab-proxy htpasswd user set via argocd-apps. |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| codeartifact | object | `{"domain":"stackstate","domainOwner":"","endpoints":{"mavenReleases":"","mavenSnapshots":"","pypi":""},"region":"eu-west-1"}` | Non-secret CodeArtifact coordinates. Endpoint hostnames come from terraform-infra outputs (STAC-25404); the applier/argocd wires the real values. Do not hardcode real hostnames in this chart. |
| codeartifact.domainOwner | string | `""` | AWS account ID owning the CodeArtifact domain (terraform output `codeartifact_domain_owner`). Omitted from the token call when empty. |
| codeartifact.endpoints | object | `{"mavenReleases":"","mavenSnapshots":"","pypi":""}` | Full base URLs, exactly as returned by `aws codeartifact get-repository-endpoint --format <maven|pypi> --repository <repo>`, including the trailing `/maven/<repo>/` or `/pypi/<repo>/` path segment and trailing slash. All three are REQUIRED terraform-infra outputs (STAC-25404); rendering fails while any is unset. |
| codeartifact.endpoints.mavenReleases | string | `""` | `packages` repository Maven format endpoint base URL (terraform output `codeartifact_packages_maven_endpoint`). |
| codeartifact.endpoints.mavenSnapshots | string | `""` | `packages-snapshot` repository Maven format endpoint base URL (terraform output `codeartifact_packages_snapshot_maven_endpoint`). |
| codeartifact.endpoints.pypi | string | `""` | `packages` repository PyPI format endpoint base URL (terraform output `codeartifact_packages_pypi_endpoint`); its `/simple/` suffix comes from the client path, not from this value. |
| fullnameOverride | string | `""` |  |
| gateway.annotations | object | `{}` | Annotations for the `HTTPRoute` object. |
| gateway.backendWeight | string | `nil` | Optional weight for traffic splitting. |
| gateway.enabled | bool | `false` | Enable Gateway API `HTTPRoute` for codeartifact-proxy. Mutually exclusive with ingress.enabled. |
| gateway.filters | list | `[]` | Optional filters for the `HTTPRoute` rule. |
| gateway.hostnames | list | `[]` | List of hostnames for the `HTTPRoute`. If empty, the route matches all hosts. |
| gateway.parentRefs | list | `[]` | List of parent `Gateway` references (required when gateway.enabled is true). |
| gateway.path | string | `"/"` | Path prefix for the `HTTPRoute` rule. |
| gateway.timeouts | object | `{}` | Optional timeouts for the `HTTPRoute` rule. |
| image | object | `{"digest":"sha256:f2a364c3f576d6a25a331f41d20dd4b58883e0fb53d4ba045c49172658921dba","pullPolicy":"IfNotPresent","registry":"dp.apps.rancher.io","repository":"containers/nginx"}` | SUSE Application Collection nginx image (off-the-shelf; no custom codeartifact-proxy image is built). Pinned by digest per the Container and CI Image Policy. |
| image.digest | string | `"sha256:f2a364c3f576d6a25a331f41d20dd4b58883e0fb53d4ba045c49172658921dba"` | REQUIRED. Rendering fails while this is empty or a `sha256:0000…` placeholder. Resolve the published digest for dp.apps.rancher.io/containers/nginx from the SUSE Application Collection catalog (authenticated) and set it here. |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | string | `nil` |  |
| ingress.className | string | `"ingressClass"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"codeartifact.proxy"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
| ingress.tls[0].hosts[0] | string | `"codeartifact.proxy"` |  |
| ingress.tls[0].secretName | string | `"secret-tls"` |  |
| livenessProbe.failureThreshold | int | `6` |  |
| livenessProbe.initialDelaySeconds | int | `30` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.successThreshold | int | `1` |  |
| livenessProbe.tcpSocket.port | string | `"http"` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| nameOverride | string | `""` |  |
| nginx.cache.assetMaxSize | string | `"150g"` | On-disk cap for the versioned-asset cache zone. nginx enforces this; the volume size does not. Keep assetMaxSize + metadataMaxSize below persistence.size. |
| nginx.cache.assetTtl | string | `"30d"` | TTL for immutable versioned assets (jar/whl/sdist/pom): long, since a given coordinate+version+asset never changes content once published. |
| nginx.cache.metadataMaxSize | string | `"1g"` | On-disk cap for the metadata/index cache zone. |
| nginx.cache.metadataTtl | string | `"5m"` | TTL for Maven metadata (maven-metadata.xml) and PyPI simple-index pages: short, because these can change upstream (new snapshot/release versions published). |
| nginx.cache.notFoundTtl | string | `"1m"` | TTL for 404 responses: short, because an artifact that is missing now can be published upstream at any time. Set to 0 to disable negative caching. |
| nginx.cache.snapshotAssetTtl | string | `"5m"` | TTL for Maven snapshot assets: short, because a snapshot version's assets can be republished by a later build of the same branch. |
| nginx.errorLogLevel | string | `"warn"` | nginx error_log level, written to stderr. |
| nginx.listenPort | int | `8080` | Port nginx listens on and the Service targets. Must be above 1024 because the container runs non-root with all capabilities dropped. |
| nginx.reloadPollSeconds | int | `5` | How often the nginx container checks the sidecar's token-generation marker; the reload delay after a token rotation is at most this many seconds. |
| nginx.resolver | string | `"10.0.224.1"` | Cluster-internal DNS resolver nginx uses to resolve the CodeArtifact endpoint hostnames. Mirrors the resolver already used by gitlab-proxy in tooling-main. |
| nginx.tmpMountPath | string | `"/tmp"` | Writable temp directory for the nginx container, backed by an emptyDir. Holds nginx's pid file and its client_body/proxy/fastcgi/uwsgi/scgi temp paths: the chart's nginx.conf points every one of those here, because nginx mkdir()s them at startup and the image defaults (/var/lib/nginx/tmp/* on openSUSE builds) sit on the read-only root filesystem, which fails the process at load time. |
| nginx.tmpSizeLimit | string | `"1Gi"` | Size of that emptyDir. Only holds in-flight request/response spooling for uncached traffic, since both cache zones use use_temp_path=off and write to the cache volume. |
| nginx.workerConnections | int | `1024` | Max simultaneous connections per worker process. |
| nodeSelector | object | `{}` |  |
| persistence | object | `{"accessModes":["ReadWriteOnce"],"mountPath":"/var/cache/nginx","size":"200Gi","storageClass":null}` | Per-replica nginx cache volume (a StatefulSet volumeClaimTemplate). Caching is a primary function of this proxy, not just a latency optimization, so the cache must survive pod restarts and rollouts rather than being discarded with an emptyDir. Sized like gitlab-proxy's cache volume; keep it above the sum of nginx.cache.*MaxSize, which is what actually bounds on-disk usage. |
| persistence.mountPath | string | `"/var/cache/nginx"` | Where the cache volume is mounted; both proxy_cache_path zones live under it. |
| persistence.storageClass | string | `nil` | Storage class for the cache volume. Empty uses the cluster default. |
| podAnnotations | object | `{}` |  |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.fsGroup | int | `1001` |  |
| podSecurityContext.fsGroupChangePolicy | string | `"Always"` |  |
| podSecurityContext.supplementalGroups | list | `[]` |  |
| podSecurityContext.sysctls | list | `[]` |  |
| readinessProbe.failureThreshold | int | `3` |  |
| readinessProbe.httpGet.path | string | `"/health"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.periodSeconds | int | `5` |  |
| readinessProbe.successThreshold | int | `1` |  |
| readinessProbe.timeoutSeconds | int | `3` |  |
| replicaCount | int | `2` |  |
| resources | object | `{"limits":{"ephemeral-storage":"2Gi","memory":"256Mi"},"requests":{"cpu":"250m","ephemeral-storage":"50Mi","memory":"256Mi"}}` | nginx container resources. The cached artifacts live on the persistent cache volume and so do not count here; ephemeral-storage only has to cover nginx.tmpSizeLimit plus logs. |
| revisionHistoryLimit | int | `10` |  |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsGroup":1001,"runAsNonRoot":true,"runAsUser":1001,"seLinuxOptions":{},"seccompProfile":{"type":"RuntimeDefault"}}` | Applies to the nginx container. Non-root, read-only root filesystem, all capabilities dropped; only the mounted volumes are writable, which is why the chart owns the whole nginx.conf and redirects every path nginx writes to (nginx.tmpMountPath, persistence.mountPath). |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `true` | Must stay true: the IRSA/pod-identity webhook mutates this ServiceAccount's token projection, and the sidecar needs it to call CodeArtifact. |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| sidecar | object | `{"image":{"digest":"sha256:5a652068c480d84b709565f9c640d451f29615a5b24a28de83a3c6383160431e","pullPolicy":"IfNotPresent","registry":"dp.apps.rancher.io","repository":"containers/aws-cli"},"refresh":{"maxBackoffRetrySeconds":60,"readinessMarginSeconds":600,"refreshIntervalSeconds":7200,"tokenValiditySeconds":14400},"resources":{"limits":{"cpu":"250m","ephemeral-storage":"256Mi","memory":"128Mi"},"requests":{"cpu":"50m","ephemeral-storage":"32Mi","memory":"64Mi"}},"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsGroup":1001,"runAsNonRoot":true,"runAsUser":1001,"seLinuxOptions":{},"seccompProfile":{"type":"RuntimeDefault"}},"tmpSizeLimit":"64Mi"}` | Token-refresh sidecar, in the same pod as nginx. Runs under the pod's IRSA identity, calls `aws codeartifact get-authorization-token`, and atomically replaces a 0600 nginx `include` file on a memory-backed emptyDir shared with the nginx container, containing exactly:   proxy_set_header Authorization "Basic <base64(aws:token)>"; It never logs the token/header. On refresh failure it keeps the last good include and retries with capped backoff. After a successful refresh it bumps a token-generation marker next to the include; the nginx container's own entrypoint polls that marker and reloads nginx in place, so the pod needs no shared process namespace. |
| sidecar.refresh.maxBackoffRetrySeconds | int | `60` | Cap on retry backoff after a failed refresh. The last good include file is retained (never deleted) while retries continue. |
| sidecar.refresh.readinessMarginSeconds | int | `600` | The sidecar reports Ready only while the current token still has at least this many seconds of validity left. PRODUCTION DEFAULT (600s / 10m). Use 120 only for the manual refresh validation check. |
| sidecar.refresh.refreshIntervalSeconds | int | `7200` | Steady-state interval between successful refreshes. PRODUCTION DEFAULT (7200s / 2h). Use 240 only for the manual refresh validation check. |
| sidecar.refresh.tokenValiditySeconds | int | `14400` | Requested validity for `get-authorization-token`. PRODUCTION DEFAULT (14400s / 4h). Use 900 only for the manual refresh validation check (Verification item 5 / the mandatory user gate), then restore this value before merging. |
| sidecar.tmpSizeLimit | string | `"64Mi"` | Size of the sidecar's writable /tmp emptyDir. The root filesystem is read-only and HOME is set to /tmp, so this holds the AWS CLI's own cache. No token material is written here: the include lives on the memory-backed `codeartifact-auth` volume. |
| tests.image | object | `{"repository":"registry.suse.com/bci/bci-base","tag":"15.7"}` | Image for the `helm test` connection pod. bci-base is the smallest BCI variant that ships curl (bci-micro/bci-busybox do not). |
| tolerations | list | `[]` |  |

