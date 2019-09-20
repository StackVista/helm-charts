aws-nuke
========
AWS Nuke -- clean an entire AWS account

Current chart version is `0.1.0`

Source code can be found [here](https://github.com/rebuy-de/aws-nuke)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.stackstate.io/ | common | 0.1.4 |

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| awsNuke.config | string | `nil` | **REQUIRED** AWS Nuke configuration file that will be used with the `--config` flag |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/rebuy/aws-nuke@sha256"` | Base container image registry. |
| image.tag | string | `"f77b1f248cd7a036411ee8205d928dee5ac4f9cb357b267da30931a4ae75cc37"` | Default container image tag. |
