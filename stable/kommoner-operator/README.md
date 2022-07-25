# kommoner-operator

The StackState Common Objects Operator

Current chart version is `0.2.2`

**Homepage:** <https://gitlab.com/StackVista/devops/kommoner>

## Example values file

````
rbac:
- apiGroup: networking.k8s.io
  objects:
  - networkpolicies
commonObjects:
- apiVersion: networking.k8s.io/v1
  kind: NetworkPolicy
  metadata:
    name: allow-traffic
    labels:
    environment: production
  spec:
    egress:
    - {}
    ingress:
    - {}
    podSelector: {}
    policyTypes:
    - Ingress
    - Egress
````

## Values

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| commonObjects | list | `[]` | Common objects to be installed in all namespaces |
| image.pullPolicy | string | `"IfNotPresent"` | Pull policy for the image for the Kommoner operator |
| image.registry | string | `"quay.io"` | Registry containing the image for the Kommoner operator |
| image.repository | string | `"stackstate/kommoner"` | Repository containing the image for the Kommoner operator |
| image.tag | string | `"v0.2.0"` | Tag of the image for the Kommoner operator |
| namespaces.exclude | list | `["default","kube-system","kube-public","kube-node-lease"]` | Namespaces to exclude from processing by the Kommoner |
| rbac | list | `[]` | Types of Kubernetes objects that are created by Kommoner, these are added to an RBAC ClusterRole so that Kommoner may manipulate them |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"64Mi"` |  |
| resources.requests.cpu | string | `"25m"` |  |
| resources.requests.memory | string | `"32Mi"` |  |
| securityContext | object | `{"runAsNonRoot":true,"runAsUser":65532}` | SecurityContext for the Kommoner pod |
