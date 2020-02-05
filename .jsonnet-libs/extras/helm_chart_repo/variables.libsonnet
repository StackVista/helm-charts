{
  local Variable = self,

  helm: {
    repositories: {
      aws: 'https://aws.github.io/eks-charts',
      bitnami: 'https://charts.bitnami.com/bitnami',
      elastic: 'https://helm.elastic.co',
      emberstack: 'https://emberstack.github.io/helm-charts',
      gitlab: 'https://charts.gitlab.io',
      incubator: 'https://kubernetes-charts-incubator.storage.googleapis.com',
      jetstack: 'https://charts.jetstack.io',
      jfrog: 'https://charts.jfrog.io',
      kube_eagle: 'https://raw.githubusercontent.com/cloudworkz/kube-eagle-helm-chart/master',
      stable: 'https://kubernetes-charts.storage.googleapis.com',
      stackstate: 'https://helm.stackstate.io',
      stackstate_test: 'https://helm-test.stackstate.io',
      stakater: 'https://stakater.github.io/stakater-charts',
      vmware_tanzu: 'https://vmware-tanzu.github.io/helm-charts',
    },
  },
}
