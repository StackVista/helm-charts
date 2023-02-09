To make a new release of this helm chart, follow the following steps:

- Create a branch from master
- Set the latest tags for the docker images, for each of these images in values.yaml:
  * https://quay.io/repository/stackstate/stackstate-k8s-cluster-agent
    * [clusterAgent.image.tag]
  * https://quay.io/repository/stackstate/stackstate-k8s-agent
    * [nodeAgent.containers.agent.image.tag]
    * [checksAgent.image.tag]
  * https://quay.io/repository/stackstate/stackstate-k8s-process-agent
    * [nodeAgent.containers.processAgent.image.tag]
- Bump the version of the chart
- Merge the mr and hit the public release button on the ci pipeline
- Manually test the newly released stackstate/stackstate-k8s-agent chart
