# otel-demo

![Version: 0.0.16](https://img.shields.io/badge/Version-0.0.16-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

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
| featureflags.resources.limits.cpu | string | `"200m"` |  |
| featureflags.resources.limits.memory | string | `"200Mi"` |  |
| featureflags.resources.requests.cpu | string | `"100m"` |  |
| featureflags.resources.requests.memory | string | `"100Mi"` |  |
| opentelemetry-demo.components.accountingService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.accountingService.resources.limits.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.accountingService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.accountingService.resources.requests.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.adService.imageOverride.repository | string | `"quay.io/stackstate/opentelemetry-demo"` |  |
| opentelemetry-demo.components.adService.imageOverride.tag | string | `"dev-bb5b07ef-adservice"` |  |
| opentelemetry-demo.components.adService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.adService.resources.limits.memory | string | `"300Mi"` |  |
| opentelemetry-demo.components.adService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.adService.resources.requests.memory | string | `"300Mi"` |  |
| opentelemetry-demo.components.cartService.podAnnotations."monitor.kubernetes-v2.stackstate.io/pod-span-error-ratio" | string | `"{ \"threshold\": 0.02 }"` |  |
| opentelemetry-demo.components.cartService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.cartService.resources.limits.memory | string | `"160Mi"` |  |
| opentelemetry-demo.components.cartService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.cartService.resources.requests.memory | string | `"160Mi"` |  |
| opentelemetry-demo.components.checkoutService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.checkoutService.resources.limits.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.checkoutService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.checkoutService.resources.requests.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.currencyService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.currencyService.resources.limits.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.currencyService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.currencyService.resources.requests.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.emailService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.emailService.resources.limits.memory | string | `"100Mi"` |  |
| opentelemetry-demo.components.emailService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.emailService.resources.requests.memory | string | `"100Mi"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].name | string | `"DISABLE_FEATURE_FLAGS"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].value | string | `"true"` |  |
| opentelemetry-demo.components.featureflagService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.featureflagService.resources.limits.memory | string | `"250Mi"` |  |
| opentelemetry-demo.components.featureflagService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.featureflagService.resources.requests.memory | string | `"250Mi"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].data."99-ffs_update.sql" | string | `"UPDATE public.featureflags SET enabled = 1 WHERE name = 'adServiceFailure';\n"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].mountPath | string | `"/docker-entrypoint-initdb.d/99-ffs_update.sql"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].name | string | `"init-scripts"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].subPath | string | `"99-ffs_update.sql"` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsGroup | int | `70` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.components.ffsPostgres.podSecurityContext.runAsUser | int | `70` |  |
| opentelemetry-demo.components.ffsPostgres.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.ffsPostgres.resources.limits.memory | string | `"120Mi"` |  |
| opentelemetry-demo.components.ffsPostgres.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.ffsPostgres.resources.requests.memory | string | `"120Mi"` |  |
| opentelemetry-demo.components.frauddetectionService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.frauddetectionService.resources.limits.memory | string | `"200Mi"` |  |
| opentelemetry-demo.components.frauddetectionService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.frauddetectionService.resources.requests.memory | string | `"200Mi"` |  |
| opentelemetry-demo.components.frontend.podAnnotations."monitor.kubernetes-v2.stackstate.io/pod-span-error-ratio" | string | `"{ \"threshold\": 0.25 }"` |  |
| opentelemetry-demo.components.frontend.resources.limits.cpu | string | `"500m"` |  |
| opentelemetry-demo.components.frontend.resources.limits.memory | string | `"200Mi"` |  |
| opentelemetry-demo.components.frontend.resources.requests.cpu | string | `"150m"` |  |
| opentelemetry-demo.components.frontend.resources.requests.memory | string | `"200Mi"` |  |
| opentelemetry-demo.components.frontend.service.annotations."monitor.kubernetes-v2.stackstate.io/http-error-ratio-for-service" | string | `"{\n  \"criticalThreshold\": 0.05,\n  \"deviatingThreshold\": 0.001\n}\n"` |  |
| opentelemetry-demo.components.frontend.service.annotations."monitor.kubernetes-v2.stackstate.io/k8s-service-span-error-ratio" | string | `"{ \"threshold\": 0.25 }"` |  |
| opentelemetry-demo.components.frontendProxy.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.frontendProxy.resources.limits.memory | string | `"50Mi"` |  |
| opentelemetry-demo.components.frontendProxy.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.frontendProxy.resources.requests.memory | string | `"50Mi"` |  |
| opentelemetry-demo.components.kafka.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.kafka.resources.limits.memory | string | `"500Mi"` |  |
| opentelemetry-demo.components.kafka.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.kafka.resources.requests.memory | string | `"500Mi"` |  |
| opentelemetry-demo.components.loadgenerator.imageOverride.repository | string | `"quay.io/stackstate/opentelemetry-demo"` |  |
| opentelemetry-demo.components.loadgenerator.imageOverride.tag | string | `"dev-7a35e404-loadgenerator"` |  |
| opentelemetry-demo.components.paymentService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.paymentService.resources.limits.memory | string | `"120Mi"` |  |
| opentelemetry-demo.components.paymentService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.paymentService.resources.requests.memory | string | `"120Mi"` |  |
| opentelemetry-demo.components.productCatalogService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.productCatalogService.resources.limits.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.productCatalogService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.productCatalogService.resources.requests.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.quoteService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.quoteService.resources.limits.memory | string | `"40Mi"` |  |
| opentelemetry-demo.components.quoteService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.quoteService.resources.requests.memory | string | `"40Mi"` |  |
| opentelemetry-demo.components.recommendationService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.recommendationService.resources.limits.memory | string | `"500Mi"` |  |
| opentelemetry-demo.components.recommendationService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.recommendationService.resources.requests.memory | string | `"500Mi"` |  |
| opentelemetry-demo.components.redis.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.redis.resources.limits.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.redis.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.redis.resources.requests.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.shippingService.resources.limits.cpu | string | `"300m"` |  |
| opentelemetry-demo.components.shippingService.resources.limits.memory | string | `"20Mi"` |  |
| opentelemetry-demo.components.shippingService.resources.requests.cpu | string | `"50m"` |  |
| opentelemetry-demo.components.shippingService.resources.requests.memory | string | `"20Mi"` |  |
| opentelemetry-demo.default.podSecurityContext.runAsGroup | int | `65534` |  |
| opentelemetry-demo.default.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.default.podSecurityContext.runAsUser | int | `65534` |  |
| opentelemetry-demo.opentelemetry-collector.config.connectors.spanmetrics.exemplars.enabled | bool | `false` |  |
| resourceQuota.enabled | bool | `true` | Whether the ResourceQuota should be created for the demo namespace. |
| resourceQuota.hard | object | `{"limits.cpu":"30","limits.memory":"50Gi","requests.cpu":"10","requests.memory":"20Gi"}` | Hard is the set of desired hard limits for each named resource. |

