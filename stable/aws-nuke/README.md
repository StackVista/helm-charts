aws-nuke
========
AWS Nuke -- Clean an entire AWS account

Current chart version is `0.2.0`

Source code can be found [here](https://github.com/rebuy-de/aws-nuke)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.stackstate.io/ | common | 0.3.0 |

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| awsNuke.config | string | `nil` | **REQUIRED** AWS Nuke configuration file that will be used with the `--config` flag. |
| awsNuke.extraArgs | object | `{}` | Extra command-line options to pass. |
| awsNuke.noDryRun | bool | `false` | If specified, it actually deletes found resources. Otherwise it just lists all candidates. |
| awsNuke.profile | string | `""` | Name of the AWS profile name for accessing the AWS API. Cannot be used together with `--access-key-id` and `--secret-access-key`. |
| awsNuke.quiet | bool | `false` | Don't show filtered resources. |
| awsNuke.verbose | bool | `false` | Enables debug output. |
| concurrencyPolicy | string | `"Forbid"` |  |
| failedJobsHistoryLimit | int | `5` |  |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/rebuy/aws-nuke"` | Base container image registry. |
| image.tag | string | `"v2.14.0"` | Default container image tag. |
| schedule | string | `"*/10 * * * *"` |  |
| successfulJobsHistoryLimit | int | `5` |  |
