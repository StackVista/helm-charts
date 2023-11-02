# redirector

Redirector can help with redirecting users to their own URL.

Current chart version is `0.0.2`

**Homepage:** <https://gitlab.com/StackVista/platform/redirector>

## Example values file
The defaults should be fine for most cases. Only an ingress configuration needs to be added and possibly a `image.pullSecretName`:

```
pullSecretName: my-pull-secret
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: ingress-nginx-external
    nginx.ingress.kubernetes.io/ingress.class: ingress-nginx-external
    cert-manager.io/cluster-issuer: my-cluster-issuer
  hosts:
    - host: redirector.my-test-cluster.stackstate.io
  tls:
    - hosts:
      - redirector.my-test-cluster.stackstate.io
      secretName: redirector-ecc-tls
```

## Values

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0] | string | `"redirector"` |  |
| affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].topologyKey | string | `"kubernetes.io/hostname"` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the Kommoner operator |
| image.registry | string | `"quay.io"` | Registry containing the image for the Kommoner operator |
| image.repository | string | `"stackstate/redirector"` | Repository containing the image for the Kommoner operator |
| image.tag | string | `"v0.0.2"` | Tag of the image for the Kommoner operator |
| ingress.enabled | bool | `false` |  |
| ingress.path | string | `"/"` |  |
| replicaCount | int | `2` |  |
| resources.limits.cpu | string | `"25m"` |  |
| resources.limits.memory | string | `"32Mi"` |  |
| resources.requests.cpu | string | `"25m"` |  |
| resources.requests.memory | string | `"32Mi"` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `65532` |  |
