# kafkaup-operator

![Version: 0.1.15](https://img.shields.io/badge/Version-0.1.15-informational?style=flat-square) ![AppVersion: v0.1.0](https://img.shields.io/badge/AppVersion-v0.1.0-informational?style=flat-square)

The StackState Kafka upgrade operator

Current chart version is `0.1.15`

**Homepage:** <https://gitlab.com/StackVista/platform/kafkaup-operator>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Bram Schuur | <bschuur@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common/ | common | * |
| file://../suse-observability-sizing | suse-observability-sizing | 0.1.3 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| global.commonLabels | object | `{}` | Labels to apply to all resources |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the KafkaUp operator |
| image.registry | string | `"quay.io"` | Registry containing the image for the KafkaUp operator |
| image.repository | string | `"stackstate/kafkaup-operator"` | Repository containing the image for the KafkaUp operator |
| image.tag | string | `"0.0.1"` | Tag of the image for the KafkaUp operator |
| kafkaSelectors.podLabel | object | `{"key":"app.kubernetes.io/component","value":"kafka"}` | pod label of kafka pods to operate on |
| kafkaSelectors.statefulSetName | string | `"kafka"` | name of the statefulSet to operate on |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"64Mi"` |  |
| resources.requests.cpu | string | `"25m"` |  |
| resources.requests.memory | string | `"32Mi"` |  |
| securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| securityContext.fsGroup | int | `65534` | The GID (group ID) used to mount volumes |
| securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| startVersion | string | `nil` | Version to use if no version is set. Allow going from a non-operated to operated situation |
| strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
