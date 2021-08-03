# beacher

The StackState Beacher

Current chart version is `0.1.0`

**Homepage:** <https://gitlab.com/StackVista/devops/sts-toolbox>

## Example values file

```
scaleInterval: 72h
systemNamespaces:
  - aws-pod-identity-webhook
  - cert-manager
  - default
  - ingress-nginx-external
  - ingress-nginx-internal
  - kube-node-lease
  - kube-public
  - kube-system
  - logging
  - monitoring
  - trafficmirror
  - velero
  - beacher
  - kommoner
```

## Values

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the Beacher cron job |
| image.registry | string | `"quay.io"` | Registry containing the image for the Beacher cron job |
| image.repository | string | `"stackstate/sts-toolbox"` | Repository containing the image for the Beacher cron job |
| image.tag | string | `"v1.3.11"` | Tag of the image for the Beacher cron job |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"64Mi"` |  |
| resources.requests.cpu | string | `"25m"` |  |
| resources.requests.memory | string | `"32Mi"` |  |
| scaleInterval | string | `"24h"` | Interval after which a namespace will be scaled down. |
| schedule | string | `"0 1 * * *"` | The cron schedule for the Beacher cron job. |
| systemNamespaces | list | `["kube-system","kube-public","kube-node-lease"]` | Namespaces that are considered system and off-limits to the Beacher cron job. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
