# otel-demo

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

Helm chart for Opentelemetry Demo

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://open-telemetry.github.io/opentelemetry-helm-charts | opentelemetry-demo | 0.28.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| featureflags.demoScenarioSimulation.enabled | bool | `true` | Whether the k8s demo scenario should be enabled. |
| featureflags.demoScenarioSimulation.schedule | object | `{"failure":"0 * * * *","fix":"30 * * * *"}` | The cron schedule to trigger the k8s demo scenario. |
| featureflags.demoScenarioSimulation.schedule.failure | string | `"0 * * * *"` | The cron schedule to trigger the faulty k8s demo scenario. |
| featureflags.demoScenarioSimulation.schedule.fix | string | `"30 * * * *"` | The cron schedule to fix the faulty k8s demo scenario. |
| opentelemetry-demo.components.adService.imageOverride.repository | string | `"wolverminion/demo"` |  |
| opentelemetry-demo.components.adService.imageOverride.tag | string | `"1.7.0-adservice"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].name | string | `"DISABLE_FEATURE_FLAGS"` |  |
| opentelemetry-demo.components.featureflagService.envOverrides[0].value | string | `"false"` |  |
| opentelemetry-demo.components.featureflagService.imageOverride.repository | string | `"quay.io/stackstate/opentelemetry-demo"` |  |
| opentelemetry-demo.components.featureflagService.imageOverride.tag | string | `"dev-d1314aa7-featureflagservice"` |  |
| opentelemetry-demo.components.featureflagService.resources.limits.memory | string | `nil` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].data."10-ffs_schema.sql" | string | `"CREATE TABLE public.featureflags (\n    name character varying(255),\n    description character varying(255),\n    enabled double precision DEFAULT 0.0 NOT NULL\n);\nALTER TABLE ONLY public.featureflags ADD CONSTRAINT featureflags_pkey PRIMARY KEY (name);\nCREATE UNIQUE INDEX featureflags_name_index ON public.featureflags USING btree (name);\n"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].data."20-ffs_data.sql" | string | `"-- Feature Flags created and initialized on startup\nINSERT INTO public.featureflags (name, description, enabled)\nVALUES\n    ('productCatalogFailure', 'Fail product catalog service on a specific product', 0),\n    ('recommendationCache', 'Cache recommendations', 0),\n    ('adServiceFailure', 'Fail ad service requests', 0),\n    ('cartServiceFailure', 'Fail cart service requests', 0);\n"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].mountPath | string | `"/docker-entrypoint-initdb.d"` |  |
| opentelemetry-demo.components.ffsPostgres.mountedConfigMaps[0].name | string | `"init-scripts"` |  |
| opentelemetry-demo.components.frauddetectionService.imageOverride.repository | string | `"wolverminion/demo"` |  |
| opentelemetry-demo.components.frauddetectionService.imageOverride.tag | string | `"1.7.0-frauddetectionservice"` |  |
| opentelemetry-demo.default.podSecurityContext.runAsGroup | int | `65534` |  |
| opentelemetry-demo.default.podSecurityContext.runAsNonRoot | bool | `true` |  |
| opentelemetry-demo.default.podSecurityContext.runAsUser | int | `65534` |  |
| opentelemetry-demo.opentelemetry-collector.config.connectors.spanmetrics.exemplars.enabled | bool | `false` |  |
| opentelemetry-demo.opentelemetry-collector.config.exporters.otlphttp/trafficmirror.endpoint | string | `"https://trafficmirror.gke-demo-dev.gcp.stackstate.io/otlp"` |  |
| opentelemetry-demo.opentelemetry-collector.config.processors.resource/addk8scluster.attributes[0].action | string | `"insert"` |  |
| opentelemetry-demo.opentelemetry-collector.config.processors.resource/addk8scluster.attributes[0].key | string | `"k8s.cluster.name"` |  |
| opentelemetry-demo.opentelemetry-collector.config.processors.resource/addk8scluster.attributes[0].value | string | `"gke-demo-dev"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.exporters[0] | string | `"otlp"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.exporters[1] | string | `"debug"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.exporters[2] | string | `"spanmetrics"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.exporters[3] | string | `"otlphttp/trafficmirror"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.processors[0] | string | `"memory_limiter"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.processors[1] | string | `"resource"` |  |
| opentelemetry-demo.opentelemetry-collector.config.service.pipelines.traces.processors[2] | string | `"batch"` |  |

