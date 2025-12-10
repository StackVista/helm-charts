# pull-secret

Helm chart for deploying a pull-secret for StackState.

> **⚠️ DEPRECATION NOTICE**
> This chart is deprecated when used as a subchart of `suse-observability`.
> The `suse-observability` chart now supports pull secret configuration natively
> via the `global.suseObservability.pullSecret` section.
> Please use the integrated pull secret configuration instead.
> See the [suse-observability chart documentation](../suse-observability/README.md) for details.

Current chart version is `1.0.2`

**Homepage:** <https://github.com/StackVista/stackstate>

## Values

## Values

| Key         | Type | Default | Description                                                                   |
| ----------- | ---- | ------- | ----------------------------------------------------------------------------- |
| credentials | list | `[]`    | Array of Registries and their credentials (i.e. username, password, registry) |
