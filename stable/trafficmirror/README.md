# trafficmirror

![Version: 2.0.5](https://img.shields.io/badge/Version-2.0.5-informational?style=flat-square) ![AppVersion: 2.2.0](https://img.shields.io/badge/AppVersion-2.2.0-informational?style=flat-square)
Trafficmirror -- mirror traffic to various endpoints.
**Homepage:** <https://github.com/rb3ckers/trafficmirror>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | ops@stackstate.com |  |

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
| image.repository | string | `"quay.io/stackstate/trafficmirror"` | Base container image repository. |
| image.tag | string | `"v2.2.0"` | Default container image tag. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| trafficmirror.failAfterMinutes | int | `30` | Remove a target when it has been failing for this many minutes. |
| trafficmirror.mainUrl | string | `""` | The default URL to receive the mirrored traffic. |
| trafficmirror.mirrorUrls | list | `[]` | The additional URLs that should also receive mirrored traffic. |
| trafficmirror.password | string | `""` | Basic auth password for the Trafficmirror service. |
| trafficmirror.retryAfterMinutes | int | `1` | After 5 successive failures a target is temporarily disabled, it will be retried after this many minutes. |
| trafficmirror.username | string | `""` | Basic auth username for the Trafficmirror service. |
