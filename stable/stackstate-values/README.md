# stackstate-values

Helm Chart for rendering StackState Values

Current chart version is `1.0.0`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Usage

This chart is used to generate a StackState values file that can be used with the StackState Helm Chart to deploy the StackState platform.
You can use this by running the following command:

```shell
helm template \
  --name stackstate \
  --set license=<your-license-key> \
  --set baseUrl=<your-stackstate-base-url> \
  --set pullSecret.username=<your-docker-registry-username> \
  --set pullSecret.password=<your-docker-registry-password> \
  stackstate/stackstate-values
```

The output of this file can be used as the values file for the StackState Helm Chart.

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `license`
* `baseUrl`
* `pullSecret.username`
* `pullSecret.password`

## Optional Values

The following values can be optionally set.

* `receiverApiKey` - If omitted, a random API key will be generated.
* `adminPassword` - If omitted, a random password will be generated and will be output in a separate Yaml document.
* `adminApiPassword` - If omitted, a random password will be generated and will be output in a separate Yaml document.
* `imageRegistry` - If omitted, the default value of `quay.io` will be used.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| adminApiPassword | string | `""` | The password for the default 'admin' user used for authenticating with the StackState Admin API. If not provided a random password is generated. If the password is not a bcrypt hash, but provided in plaintext, the value will be bcrypt hashed in the output. |
| adminPassword | string | `""` | The password for the default 'admin' user used for authenticating with the StackState UI. If not provided a random password is generated.  If the password is not a bcrypt hash, but provided in plaintext, the value will be bcrypt hashed in the output. |
| baseUrl | string | `""` | The base URL of the StackState instance. |
| imageRegistry | string | `"quay.io"` | The registry to pull the StackState images from. |
| license | string | `nil` | The StackState license key. |
| pullSecret.password | string | `""` | The password used for pulling all StackState images from the registry. |
| pullSecret.username | string | `""` | The username used for pulling all StackState images from the registry. |
| receiverApiKey | string | `""` | The StackState Receiver API Key, used for sending telemetry data to StackState's Receiver API. |
