# trafficmirror

![Version: 2.1.1](https://img.shields.io/badge/Version-2.1.1-informational?style=flat-square) ![AppVersion: 2.5.4](https://img.shields.io/badge/AppVersion-2.5.4-informational?style=flat-square)
Trafficmirror -- mirror traffic to various endpoints.
**Homepage:** <https://github.com/rb3ckers/trafficmirror>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common/ | common | * |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| container.livenessProbeDefaults.enabled | bool | `true` | Use defaults for the `livenessProbe` from the upstream `common` chart. |
| container.readinessProbeDefaults.enabled | bool | `true` | Use defaults for the `readinessProbe` from the upstream `common` chart. |
| deployment.securityContext.runAsNonRoot | bool | `true` |  |
| deployment.securityContext.runAsUser | int | `65534` |  |
| image.repository | string | `"ghcr.io/rb3ckers/trafficmirror"` | Base container image repository. |
| image.tag | string | `"v2.5.4"` | Default container image tag. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| trafficmirror.enablePProf | bool | `false` | Enable pprof profiling |
| trafficmirror.failAfterMinutes | int | `30` | Remove a target when it has been failing for this many minutes. |
| trafficmirror.mainTargetDelayMs | int | `200` | Delay executions to main target, this gives the mirror time to catch up, and increases parallelism. |
| trafficmirror.mainUrl | string | `""` | The default URL to receive the mirrored traffic. |
| trafficmirror.maxQueuedRequests | int | `3000` | Max requests that gets queued per mirror target. |
| trafficmirror.mirrorUrls | list | `[]` | The additional URLs that should also receive mirrored traffic. |
| trafficmirror.password | string | `""` | Basic auth password for the Trafficmirror service. |
| trafficmirror.retryAfterMinutes | int | `1` | After 5 successive failures a target is temporarily disabled, it will be retried after this many minutes. |
| trafficmirror.username | string | `""` | Basic auth username for the Trafficmirror service. |
