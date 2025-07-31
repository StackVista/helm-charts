# suse-observability-values

Helm Chart for rendering SUSE Observability Values

Current chart version is `1.2.0`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Usage

This chart is used to generate a SUSE Observability values file that can be used with the SUSE Observability Helm Chart to deploy the SUSE Observability platform.
You can use this by running the following command:

```shell
helm template \
  --name suse-observability \
  --set license=<your-license-key> \
  --set baseUrl=<your-suse-observability-base-url> \
  --set pullSecret.username=<your-docker-registry-username> \
  --set pullSecret.password=<your-docker-registry-password> \
  suse-observability/suse-observability-values
```

The output of this file can be used as the values file for the SUSE Observability Helm Chart.

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `license`
* `baseUrl`

## Optional Values

The following values can be optionally set.

* `receiverApiKey` - If omitted, a random API key will be generated.
* `adminPassword` - If omitted, a random password will be generated and will be output in a separate Yaml document.
* `adminApiPassword` - If omitted, a random password will be generated and will be output in a separate Yaml document.
* `imageRegistry` - If omitted, the default value of `quay.io` will be used.
* `pullSecret.username` - If pullSecret username or password is omitted no pull secret will be configured
* `pullSecret.password` - If pullSecret username or password is omitted no pull secret will be configured

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| adminPassword | string | `""` | The password for the default 'admin' user used for authenticating with the SUSE Observability UI. If not provided a random password is generated.  If the password is not a bcrypt hash, but provided in plaintext, the value will be bcrypt hashed in the output. |
| affinity.nodeAffinity | string | `nil` | Node Affinity settings |
| affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution | bool | `true` | Enable required pod anti-affinity |
| affinity.podAntiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | Topology key for pod anti-affinity |
| baseConfig.generate | bool | `true` | If we want to generate the base configuration |
| baseUrl | string | `""` | The base URL of the SUSE Observability instance. |
| imageRegistry | string | `"registry.rancher.com"` | The registry to pull the SUSE Observability images from. |
| license | string | `nil` | The SUSE Observability license key. |
| pullSecret.password | string | `nil` | The password used for pulling all SUSE Observability images from the registry. |
| pullSecret.username | string | `nil` | The username used for pulling all SUSE Observability images from the registry. |
| receiverApiKey | string | `""` | The SUSE Observability Receiver API Key, used for sending telemetry data to the server. |
| sizing.generate | bool | `true` | If we want to generate the sizing values that match the amount of nodes we are monitoring |
| sizing.profile | string | `""` | Profile. OneOf 10-nonha, 20-nonha, 50-nonha, 100-nonha, 150-ha, 250-ha, 500-ha |
