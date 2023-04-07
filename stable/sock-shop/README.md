# sock-shop

This chart deploys [SockShop](https://microservices-demo.github.io/) application. It was forked from [microservices-demo/microservices-demo](https://github.com/microservices-demo/microservices-demo/tree/master/deploy/kubernetes/helm-chart)
and got some improvements.

![Version: 0.2.1](https://img.shields.io/badge/Version-0.2.1-informational?style=flat-square) ![AppVersion: 0.3.5](https://img.shields.io/badge/AppVersion-0.3.5-informational?style=flat-square)
A Helm chart for Sock Shop

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| all.java.options | string | `"-Xms512m -Xmx512m -XX:PermSize=32m -XX:MaxPermSize=64m -XX:+UseG1GC -Djava.security.egd=file:/dev/urandom"` | The common JAVA_OPTS for all Java microservices. |
| all.securityCapabilitiesDisabled | bool | `false` | Disable securtyContext capabilities operations. |
| carts.affinity | object | `{}` | Affinity settings for carts pods. |
| carts.resources | object | `{"limits":{"cpu":"500m","memory":"1000Mi"},"requests":{"cpu":"500m","memory":"1000Mi"}}` | Resource allocation for `carts` pods. |
| carts.version | int | `1` | Custom label (version) value for `carts` pods. |
| cartsDB.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `carts-db` pods. |
| catalogue.annotations | object | `{"vcs":"https://gitlab.com/stackvista/demo/microservices-demo/catalogue/-/commit/e9e5338599dbda30366b38d00794c34aaa4221a7"}` | Custom annotations for `catalogue` pods. |
| catalogue.demoScenarioSimulation.enabled | bool | `false` | Whether the k8s demo scenario should be enabled. |
| catalogue.demoScenarioSimulation.schedule | object | `{"failure":"0 * * * *","fix":"30 * * * *"}` | The cron schedule to trigger the k8s demo scenario. |
| catalogue.demoScenarioSimulation.schedule.failure | string | `"0 * * * *"` | The cron schedule to trigger the faulty k8s demo scenario. |
| catalogue.demoScenarioSimulation.schedule.fix | string | `"30 * * * *"` | The cron schedule to fix the faulty k8s demo scenario. |
| catalogue.image.repository | string | `"quay.io/stackstate/weaveworksdemo-catalogue"` | The container repository for `catalogue` images. |
| catalogue.image.tag | string | `"0.3.5"` | The container image tag. |
| catalogue.resources | object | `{"limits":{"cpu":"100m","memory":"200Mi"},"requests":{"cpu":"100m","memory":"200Mi"}}` | Resource allocation for `catalogue` pods. |
| catalogueDB.image.repository | string | `"quay.io/stackstate/weaveworksdemo-catalogue-db"` | The container repository for `catalogue-db` images. |
| catalogueDB.image.tag | string | `"0.3.1"` | The container image tag. |
| catalogueDB.resources | object | `{"limits":{"cpu":"1000m","memory":"500Mi"},"requests":{"cpu":"500m","memory":"250Mi"}}` | Resource allocation for `catalogue-db` pods. |
| frontend.image.repository | string | `"quay.io/stackstate/weaveworksdemo-front-end"` | The container repository for `frontend` images. |
| frontend.image.tag | string | `"0.3.13"` | The container image tag. |
| frontend.replicas | int | `1` | The number or replicas of `frontend` deployment. |
| frontend.resources | object | `{"limits":{"cpu":"300m","memory":"1000Mi"},"requests":{"cpu":"200m","memory":"1000Mi"}}` | Resource allocation for `frontend` pods. |
| ingress.replicas | int | `1` | The number or replicas of `ingress` deployment. |
| ingress.resources | object | `{"limits":{"cpu":"300m","memory":"1000Mi"},"requests":{"cpu":"200m","memory":"1000Mi"}}` | Resource allocation for `ingress` pods. |
| loadgen.enabled | bool | `false` |  |
| loadgen.image.repository | string | `"quay.io/stackstate/loadgen"` | The container repository for `loadgen` images. |
| loadgen.image.tag | string | `"master-a448026e"` | The container image tag. |
| loadgen.replicas | int | `1` | The number or replicas of `loadgen` deployment. |
| loadgen.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `loadgen` pods. |
| nameOverride | string | `""` | A name to prepend the Helm resources, if not specified the name of the chart is used. |
| orders.backorderPriority | int | `0` | whatever it means. |
| orders.image.repository | string | `"quay.io/stackstate/weaveworksdemo-orders"` | The container repository for `orders` images. |
| orders.image.tag | string | `"master"` | The container image tag. |
| orders.orderPriority | int | `0` | whatever it means. |
| orders.resources | object | `{"limits":{"cpu":"500m","memory":"1000Mi"},"requests":{"cpu":"500m","memory":"1000Mi"}}` | Resource allocation for `orders` pods. |
| orders.shippingPriority | int | `0` | whatever it means. |
| ordersDB.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `orders-db` pods. |
| payment.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `payment` pods. |
| priority | int | `1000000000` | priority for the custom PriorityClass |
| queueMaster.resources | object | `{"limits":{"cpu":"500m","memory":"1000Mi"},"requests":{"cpu":"500m","memory":"1000Mi"}}` | Resource allocation for `queue-master` pods. |
| rabbitmq.resources | object | `{"limits":{"cpu":"100m","memory":"200Mi"},"requests":{"cpu":"100m","memory":"200Mi"}}` | Resource allocation for `rabbitmq` pods. |
| scc.enabled | bool | `false` | Create `SecurityContextConstraints` resource to manage Openshift security constraints for Stackstate. |
| scc.extraServiceAccounts | list | `[]` | Extraccounts from the same namespace to add to SecurityContextConstraints users. |
| sessionDB.limits.cpu | string | `"100m"` |  |
| sessionDB.limits.memory | string | `"100Mi"` |  |
| sessionDB.requests.cpu | string | `"100m"` |  |
| sessionDB.requests.memory | string | `"100Mi"` |  |
| shipping.resources | object | `{"limits":{"cpu":"500m","memory":"1000Mi"},"requests":{"cpu":"500m","memory":"1000Mi"}}` | Resource allocation for `shipping` pods. |
| user.limits.cpu | string | `"300m"` |  |
| user.limits.memory | string | `"1000Mi"` |  |
| user.requests.cpu | string | `"100m"` |  |
| user.requests.memory | string | `"400Mi"` |  |
| userDB.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `user-db` pods. |
