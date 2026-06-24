# trafficmirror

![Version: 3.2.0](https://img.shields.io/badge/Version-3.2.0-informational?style=flat-square) ![AppVersion: 2.5.4](https://img.shields.io/badge/AppVersion-2.5.4-informational?style=flat-square)
Trafficmirror -- mirror traffic to various endpoints.
**Homepage:** <https://github.com/rb3ckers/trafficmirror>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../../local/common/ | common | * |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| container.livenessProbeDefaults.enabled | bool | `true` | Use defaults for the `livenessProbe` from the upstream `common` chart. |
| container.readinessProbeDefaults.enabled | bool | `true` | Use defaults for the `readinessProbe` from the upstream `common` chart. |
| deployment.securityContext.runAsNonRoot | bool | `true` |  |
| deployment.securityContext.runAsUser | int | `65534` |  |
| gateway.additionalGateways | list | `[]` | Additional route objects (e.g. a GRPCRoute alongside an HTTPRoute). Each entry takes the same fields as this block. |
| gateway.annotations | object | `{}` | Annotations for the route object. |
| gateway.backendWeight | string | `nil` | Optional weight for traffic splitting. |
| gateway.enabled | bool | `false` | Enable Gateway API routes (HTTPRoute/GRPCRoute) for trafficmirror. Mutually exclusive with ingress.enabled. |
| gateway.filters | list | `[]` | Optional filters for the route rule. |
| gateway.hostnames | list | `[]` | List of hostnames for the route. If empty, the route matches all hosts. |
| gateway.kind | string | `"HTTPRoute"` | Route kind to create. Valid values: "HTTPRoute" or "GRPCRoute". |
| gateway.parentRefs | list | `[]` | List of parent `Gateway` references (required when gateway.enabled is true). |
| gateway.path | string | `"/"` | Path prefix for the HTTPRoute rule (ignored for GRPCRoute). |
| gateway.pathType | string | `"PathPrefix"` | Path match type for the HTTPRoute rule (ignored for GRPCRoute). |
| gateway.timeouts | object | `{}` | Optional timeouts for the HTTPRoute rule (ignored for GRPCRoute). |
| image.repository | string | `"ghcr.io/rb3ckers/trafficmirror"` | Base container image repository. |
| image.tag | string | `"v2.5.4"` | Default container image tag. |
| ingress.annotations | object | `{}` | Annotations for `Ingress` objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames. |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| service.appProtocol | string | `""` | Optional appProtocol for the service port. Set to `kubernetes.io/h2c` for gRPC / HTTP2 cleartext backends so a Gateway API `HTTPRoute` forwards HTTP/2 upstream. |
| trafficmirror.enablePProf | bool | `false` | Enable pprof profiling |
| trafficmirror.failAfterMinutes | int | `30` | Remove a target when it has been failing for this many minutes. |
| trafficmirror.mainTargetDelayMs | int | `200` | Delay executions to main target, this gives the mirror time to catch up, and increases parallelism. |
| trafficmirror.mainUrl | string | `""` | The default URL to receive the mirrored traffic. |
| trafficmirror.maxQueuedRequests | int | `3000` | Max requests that gets queued per mirror target. |
| trafficmirror.mirrorUrls | list | `[]` | The additional URLs that should also receive mirrored traffic. |
| trafficmirror.password | string | `""` | Basic auth password for the Trafficmirror service. |
| trafficmirror.retryAfterMinutes | int | `1` | After 5 successive failures a target is temporarily disabled, it will be retried after this many minutes. |
| trafficmirror.username | string | `""` | Basic auth username for the Trafficmirror service. |
