# otel-demo

![Version: 0.0.6](https://img.shields.io/badge/Version-0.0.6-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

Helm chart for Opentelemetry Demo

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://open-telemetry.github.io/opentelemetry-helm-charts | opentelemetry-demo | 0.29.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| featureflags.demoScenarioSimulation.enabled | bool | `true` | Whether the k8s demo scenario should be enabled. |
| featureflags.demoScenarioSimulation.schedule | object | `{"failure":"0 * * * *","fix":"30 * * * *"}` | The cron schedule to trigger the k8s demo scenario. |
| featureflags.demoScenarioSimulation.schedule.failure | string | `"0 * * * *"` | The cron schedule to trigger the faulty k8s demo scenario. |
| featureflags.demoScenarioSimulation.schedule.fix | string | `"30 * * * *"` | The cron schedule to fix the faulty k8s demo scenario. |
| opentelemetry-demo.components.featureflagService.envOverrides[0].name | string | `"DISABLE_FEATURE_FLAGS"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].value | string | `"false"` |  |
| opentelemetry-demo.components.featureflagService.imageOverride.repository | string | `"quay.io/stackstate/opentelemetry-demo"` |  |
| opentelemetry-demo.components.featureflagService.imageOverride.tag | string | `"dev-11b1c878-featureflagservice"` |  |
| opentelemetry-demo.components.featureflagService.resources.limits.memory | string | `nil` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].data."99-ffs_update.sql" | string | `"UPDATE public.featureflags SET enabled = 1 WHERE name = 'adServiceFailure';\n"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].mountPath | string | `"/docker-entrypoint-initdb.d"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].name | string | `"init-scripts"` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsGroup | int | `70` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsUser | int | `70` |  |
| opentelemetry-demo.default.podSecurityContext.runAsGroup | int | `65534` |  |
| opentelemetry-demo.default.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.default.podSecurityContext.runAsUser | int | `65534` |  |
| opentelemetry-demo.opentelemetry-collector.config.connectors.spanmetrics.exemplars.enabled | bool | `false` |  |

