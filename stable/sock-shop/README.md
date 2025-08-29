# sock-shop

This chart deploys [SockShop](https://microservices-demo.github.io/) application. It was forked from [microservices-demo/microservices-demo](https://github.com/microservices-demo/microservices-demo/tree/master/deploy/kubernetes/helm-chart)
and got some improvements.

![Version: 0.2.12](https://img.shields.io/badge/Version-0.2.12-informational?style=flat-square) ![AppVersion: 0.3.5](https://img.shields.io/badge/AppVersion-0.3.5-informational?style=flat-square)
A Helm chart for Sock Shop

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.deliveryhero.io | locust | 0.31.1 |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| all.java.options | string | `"-Xms512m -Xmx512m -XX:PermSize=32m -XX:MaxPermSize=64m -XX:+UseG1GC -Djava.security.egd=file:/dev/urandom"` | The common JAVA_OPTS for all Java microservices. |
| all.securityCapabilitiesDisabled | bool | `false` | Disable securtyContext capabilities operations. |
| carts.affinity | object | `{}` | Affinity settings for carts pods. |
| carts.annotations | object | `{}` |  |
| carts.enabled | bool | `true` |  |
| carts.image.repository | string | `"weaveworksdemos/carts"` |  |
| carts.image.tag | string | `"0.4.3"` |  |
| carts.nodeSelector | object | `{}` |  |
| carts.resources | object | `{"limits":{"cpu":"500m","memory":"1500Mi"},"requests":{"cpu":"500m","memory":"1500Mi"}}` | Resource allocation for `carts` pods. |
| carts.tolerations | list | `[]` |  |
| carts.version | int | `1` | Custom label (version) value for `carts` pods. |
| cartsDB.annotations."monitor.kubernetes-v2.stackstate.io/pod-cpu-throttling" | string | `"{\"enabled\":false}"` |  |
| cartsDB.nodeSelector | object | `{}` |  |
| cartsDB.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `carts-db` pods. |
| cartsDB.tolerations | list | `[]` |  |
| catalogue.annotations | object | `{"vcs":"https://gitlab.com/stackvista/demo/microservices-demo/catalogue/-/commit/e9e5338599dbda30366b38d00794c34aaa4221a7"}` | Custom annotations for `catalogue` pods. |
| catalogue.demoScenarioSimulation.enabled | bool | `false` | Whether the k8s demo scenario should be enabled. |
| catalogue.demoScenarioSimulation.image.repository | string | `"bitnamilegacy/kubectl"` | The container repository for the kubectl image used in demo scenario jobs. |
| catalogue.demoScenarioSimulation.image.tag | string | `"latest"` | The container image tag for the kubectl image. |
| catalogue.demoScenarioSimulation.imagePullSecrets | list | `[]` | Image pull secrets for the demo scenario jobs. |
| catalogue.demoScenarioSimulation.schedule | object | `{"failure":"0 * * * *","fix":"30 * * * *"}` | The cron schedule to trigger the k8s demo scenario. |
| catalogue.demoScenarioSimulation.schedule.failure | string | `"0 * * * *"` | The cron schedule to trigger the faulty k8s demo scenario. |
| catalogue.demoScenarioSimulation.schedule.fix | string | `"30 * * * *"` | The cron schedule to fix the faulty k8s demo scenario. |
| catalogue.enabled | bool | `true` |  |
| catalogue.image.repository | string | `"quay.io/stackstate/weaveworksdemo-catalogue"` | The container repository for `catalogue` images. |
| catalogue.image.tag | string | `"0.3.5"` | The container image tag. |
| catalogue.nodeSelector | object | `{}` |  |
| catalogue.resources | object | `{"limits":{"cpu":"100m","memory":"200Mi"},"requests":{"cpu":"100m","memory":"200Mi"}}` | Resource allocation for `catalogue` pods. |
| catalogue.tolerations | list | `[]` |  |
| catalogueDB.annotations."monitor.kubernetes-v2.stackstate.io/pod-cpu-throttling" | string | `"{\"enabled\":false}"` |  |
| catalogueDB.image.repository | string | `"quay.io/stackstate/weaveworksdemo-catalogue-db"` | The container repository for `catalogue-db` images. |
| catalogueDB.image.tag | string | `"0.3.1"` | The container image tag. |
| catalogueDB.nodeSelector | object | `{}` |  |
| catalogueDB.resources | object | `{"limits":{"cpu":"1000m","memory":"500Mi"},"requests":{"cpu":"500m","memory":"250Mi"}}` | Resource allocation for `catalogue-db` pods. |
| catalogueDB.tolerations | list | `[]` |  |
| frontend.annotations | object | `{}` |  |
| frontend.components.cartsHost | string | `"carts"` |  |
| frontend.components.catalogueHost | string | `"catalogue"` |  |
| frontend.components.ordersHost | string | `"orders"` |  |
| frontend.components.userHost | string | `"user"` |  |
| frontend.enabled | bool | `true` |  |
| frontend.image.repository | string | `"quay.io/stackstate/weaveworksdemo-front-end"` | The container repository for `frontend` images. |
| frontend.image.tag | string | `"0.3.13"` | The container image tag. |
| frontend.nodeSelector | object | `{}` |  |
| frontend.replicas | int | `1` | The number or replicas of `frontend` deployment. |
| frontend.resources | object | `{"limits":{"cpu":"300m","memory":"1000Mi"},"requests":{"cpu":"200m","memory":"1000Mi"}}` | Resource allocation for `frontend` pods. |
| frontend.tolerations | list | `[]` |  |
| global.annotations | object | `{}` |  |
| ingress.annotations | object | `{}` |  |
| ingress.replicas | int | `1` | The number or replicas of `ingress` deployment. |
| ingress.resources | object | `{"limits":{"cpu":"300m","memory":"1000Mi"},"requests":{"cpu":"200m","memory":"1000Mi"}}` | Resource allocation for `ingress` pods. |
| loadgen.annotations | object | `{}` |  |
| loadgen.enabled | bool | `false` |  |
| loadgen.image.repository | string | `"quay.io/stackstate/loadgen"` | The container repository for `loadgen` images. |
| loadgen.image.tag | string | `"master-f65782ce"` | The container image tag. |
| loadgen.nodeSelector | object | `{}` |  |
| loadgen.replicas | int | `1` | The number or replicas of `loadgen` deployment. |
| loadgen.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `loadgen` pods. |
| loadgen.tolerations | list | `[]` |  |
| locust.enabled | bool | `false` |  |
| locust.loadtest.headless | bool | `false` |  |
| locust.loadtest.locust_host | string | `"http://ingress"` |  |
| locust.loadtest.locust_lib_configmap | string | `"loadtest-lib"` |  |
| locust.loadtest.locust_locustfile_configmap | string | `"loadtest-locustfile"` |  |
| locust.loadtest.name | string | `"catalogue"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/name"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0] | string | `"locust"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[1].key | string | `"component"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[1].operator | string | `"In"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[1].values[0] | string | `"worker"` |  |
| locust.master.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].topologyKey | string | `"kubernetes.io/hostname"` |  |
| locust.master.args[0] | string | `"-u"` |  |
| locust.master.args[1] | string | `"5"` |  |
| locust.master.args[2] | string | `"-r"` |  |
| locust.master.args[3] | string | `"1"` |  |
| locust.master.args[4] | string | `"--autostart"` |  |
| locust.master.command[0] | string | `"sh"` |  |
| locust.master.command[1] | string | `"/config/docker-entrypoint.sh"` |  |
| locust.podSecurityContext.runAsUser | int | `1000` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/name"` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0] | string | `"locust"` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[1].key | string | `"component"` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[1].operator | string | `"In"` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[1].values[0] | string | `"master"` |  |
| locust.worker.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].topologyKey | string | `"kubernetes.io/hostname"` |  |
| locust.worker.args | list | `[]` |  |
| locust.worker.command[0] | string | `"sh"` |  |
| locust.worker.command[1] | string | `"/config/docker-entrypoint.sh"` |  |
| nameOverride | string | `""` | A name to prepend the Helm resources, if not specified the name of the chart is used. |
| orders.annotations | object | `{}` |  |
| orders.backorderPriority | int | `0` | whatever it means. |
| orders.components.paymentHost | string | `"payment"` |  |
| orders.components.shippingHost | string | `"shipping"` |  |
| orders.enabled | bool | `true` |  |
| orders.image.repository | string | `"quay.io/stackstate/weaveworksdemo-orders"` | The container repository for `orders` images. |
| orders.image.tag | string | `"master"` | The container image tag. |
| orders.nodeSelector | object | `{}` |  |
| orders.orderPriority | int | `0` | whatever it means. |
| orders.resources | object | `{"limits":{"cpu":"500m","memory":"1000Mi"},"requests":{"cpu":"500m","memory":"1000Mi"}}` | Resource allocation for `orders` pods. |
| orders.shippingPriority | int | `0` | whatever it means. |
| orders.tolerations | list | `[]` |  |
| ordersDB.annotations | object | `{}` |  |
| ordersDB.nodeSelector | object | `{}` |  |
| ordersDB.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `orders-db` pods. |
| ordersDB.tolerations | list | `[]` |  |
| payment.annotations | object | `{}` |  |
| payment.enabled | bool | `true` |  |
| payment.image.repository | string | `"weaveworksdemos/payment"` |  |
| payment.image.tag | string | `"0.4.3"` |  |
| payment.ingress.annotations | object | `{}` |  |
| payment.ingress.enabled | bool | `false` | Whether to deploy Ingress resource for payment service |
| payment.ingress.hostname | string | `""` |  |
| payment.ingress.ingressClassName | string | `"ingress-nginx"` |  |
| payment.ingress.tls | bool | `false` |  |
| payment.nodeSelector | object | `{}` |  |
| payment.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `payment` pods. |
| payment.tolerations | list | `[]` |  |
| priority | int | `1000000000` | priority for the custom PriorityClass |
| queueMaster.annotations | object | `{}` |  |
| queueMaster.image.repository | string | `"weaveworksdemos/queue-master"` |  |
| queueMaster.image.tag | string | `"0.3.1"` |  |
| queueMaster.nodeSelector | object | `{}` |  |
| queueMaster.resources | object | `{"limits":{"cpu":"500m","memory":"1500Mi"},"requests":{"cpu":"500m","memory":"1500Mi"}}` | Resource allocation for `queue-master` pods. |
| queueMaster.tolerations | list | `[]` |  |
| rabbitmq.annotations | object | `{}` |  |
| rabbitmq.nodeSelector | object | `{}` |  |
| rabbitmq.resources | object | `{"limits":{"cpu":"100m","memory":"200Mi"},"requests":{"cpu":"100m","memory":"200Mi"}}` | Resource allocation for `rabbitmq` pods. |
| rabbitmq.tolerations | list | `[]` |  |
| scc.enabled | bool | `false` | Create `SecurityContextConstraints` resource to manage Openshift security constraints for Stackstate. |
| scc.extraServiceAccounts | list | `[]` | Extraccounts from the same namespace to add to SecurityContextConstraints users. |
| sessionDB.annotations | object | `{}` |  |
| sessionDB.limits.cpu | string | `"100m"` |  |
| sessionDB.limits.memory | string | `"100Mi"` |  |
| sessionDB.nodeSelector | object | `{}` |  |
| sessionDB.requests.cpu | string | `"100m"` |  |
| sessionDB.requests.memory | string | `"100Mi"` |  |
| sessionDB.tolerations | list | `[]` |  |
| shipping.annotations | object | `{}` |  |
| shipping.enabled | bool | `true` |  |
| shipping.image.repository | string | `"weaveworksdemos/shipping"` |  |
| shipping.image.tag | string | `"0.4.8"` |  |
| shipping.ingress.annotations | object | `{}` |  |
| shipping.ingress.enabled | bool | `false` |  |
| shipping.ingress.hostname | string | `""` |  |
| shipping.ingress.ingressClassName | string | `"ingress-nginx"` |  |
| shipping.ingress.tls | bool | `false` |  |
| shipping.nodeSelector | object | `{}` |  |
| shipping.resources | object | `{"limits":{"cpu":"500m","memory":"1500Mi"},"requests":{"cpu":"500m","memory":"1500Mi"}}` | Resource allocation for `shipping` pods. |
| shipping.tolerations | list | `[]` |  |
| user.annotations | object | `{}` |  |
| user.enabled | bool | `true` |  |
| user.image.repository | string | `"weaveworksdemos/user"` |  |
| user.image.tag | string | `"0.4.7"` |  |
| user.limits.cpu | string | `"300m"` |  |
| user.limits.memory | string | `"1000Mi"` |  |
| user.nodeSelector | object | `{}` |  |
| user.requests.cpu | string | `"100m"` |  |
| user.requests.memory | string | `"400Mi"` |  |
| user.tolerations | list | `[]` |  |
| userDB.annotations | object | `{}` |  |
| userDB.image.repository | string | `"weaveworksdemos/user-db"` |  |
| userDB.image.tag | string | `"0.4.0"` |  |
| userDB.nodeSelector | object | `{}` |  |
| userDB.resources | object | `{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}` | Resource allocation for `user-db` pods. |
| userDB.tolerations | list | `[]` |  |
