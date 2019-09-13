trafficmirror
=============
Trafficmirror -- mirror traffic to various endpoints.

Current chart version is `0.1.3`

Source code can be found [here](https://github.com/rb3ckers/trafficmirror)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.stackstate.io/ | common | 0.1.3 |

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.repository | string | `"docker.io/stackstate/trafficmirror"` | Base container image repository. |
| image.tag | string | `"652621a6e6ace12819dbddfb43ff26cda45bda28"` | Default container image tag. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| trafficmirror.mainUrl | string | `""` | The default URL for receiving the mirrored traffic. |
| trafficmirror.password | string | `""` | Basic auth password for the Trafficmirror service. |
| trafficmirror.username | string | `""` | Basic auth username for the Trafficmirror service. |
