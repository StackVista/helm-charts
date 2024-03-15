# otel-demo

![Version: 0.0.15](https://img.shields.io/badge/Version-0.0.15-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

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
| featureflags.fixImage | string | `"quay.io/stackstate/opentelemetry-demo:dev-11cf2533-featureflagservice"` | Image for the featureflags service that fixes any of the issues triggered by feature flags (i.e. it ignores all feature flags) |
| opentelemetry-demo.components.adService.imageOverride.repository | string | `"quay.io/stackstate/opentelemetry-demo"` |  |
| opentelemetry-demo.components.adService.imageOverride.tag | string | `"dev-bb5b07ef-adservice"` |  |
| opentelemetry-demo.components.cartService.podAnnotations."monitor.kubernetes-v2.stackstate.io/pod-span-error-ratio" | string | `"{ \"threshold\": 0.02 }"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].name | string | `"DISABLE_FEATURE_FLAGS"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].value | string | `"true"` |  |
| opentelemetry-demo.components.featureflagService.resources.limits.memory | string | `nil` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].data."99-ffs_update.sql" | string | `"UPDATE public.featureflags SET enabled = 1 WHERE name = 'adServiceFailure';\n"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].mountPath | string | `"/docker-entrypoint-initdb.d/99-ffs_update.sql"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].name | string | `"init-scripts"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].subPath | string | `"99-ffs_update.sql"` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsGroup | int | `70` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsUser | int | `70` |  |
| opentelemetry-demo.components.frontend.podAnnotations."monitor.kubernetes-v2.stackstate.io/pod-span-error-ratio" | string | `"{ \"threshold\": 0.25 }"` |  |
| opentelemetry-demo.components.frontend.service.annotations."monitor.kubernetes-v2.stackstate.io/http-error-ratio-for-service" | string | `"{\n  \"criticalThreshold\": 0.05,\n  \"deviatingThreshold\": 0.001\n}\n"` |  |
| opentelemetry-demo.components.frontend.service.annotations."monitor.kubernetes-v2.stackstate.io/k8s-service-span-error-ratio" | string | `"{ \"threshold\": 0.25 }"` |  |
| opentelemetry-demo.components.loadgenerator.imageOverride.repository | string | `"quay.io/stackstate/opentelemetry-demo"` |  |
| opentelemetry-demo.components.loadgenerator.imageOverride.tag | string | `"dev-7a35e404-loadgenerator"` |  |
| opentelemetry-demo.default.podSecurityContext.runAsGroup | int | `65534` |  |
| opentelemetry-demo.default.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.default.podSecurityContext.runAsUser | int | `65534` |  |
| opentelemetry-demo.opentelemetry-collector.config.connectors.spanmetrics.exemplars.enabled | bool | `false` |  |

