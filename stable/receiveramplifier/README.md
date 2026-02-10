# receiveramplifier

![Version: 0.1.8](https://img.shields.io/badge/Version-0.1.8-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)
Receiver amplifier to increase the load on an installation.
**Homepage:** <https://gitlab.com/stackvista/stackstate.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Alejandro Acevedo | <aacevedo@stackstate.com> |  |
| Bram Schuur | <bschuur@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common/ | common | * |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| container.livenessProbeDefaults.enabled | bool | `true` | Use defaults for the `livenessProbe` from the upstream `common` chart. |
| container.readinessProbeDefaults.enabled | bool | `true` | Use defaults for the `readinessProbe` from the upstream `common` chart. |
| global.imagePullSecrets | list | `[]` | List of image pull secret names to be used by all images across all charts. |
| image.pullPolicy | string | `"IfNotPresent"` | Default image pull policy. |
| image.registry | string | `"quay.io"` | REgistry |
| image.repository | string | `"stackstate/stackstate-receiver-amplifier"` | Base container image repository. |
| image.tag | string | `"master"` | Default container image tag. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| metrics.agentAnnotationsEnabled | bool | `true` |  |
| metrics.enabled | bool | `true` | Enable metrics port. |
| metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| metrics.servicemonitor.enabled | bool | `false` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| receiveramplifier.additionalConfig | string | `""` | Additional configuration settings appended to the end of the amplifier config file |
| receiveramplifier.amplifierFactor | int | `1` | Amplification factor. |
| receiveramplifier.amplifierFactorPeak | int | `1` | Amplification factor during peak hours |
| receiveramplifier.dailyPeaks | list | `[]` | Daily peak hours (multiple is possible) defined by startTime and endTime |
| receiveramplifier.targetUrl | string | `nil` | The target URL for sending the amplified intake requests. |
