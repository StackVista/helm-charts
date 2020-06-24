{
  local Variable = self,

  helm: {
    repositories: {
      bitnami: 'https://charts.bitnami.com/bitnami',
      elastic: 'https://helm.elastic.co',
      stable: 'https://kubernetes-charts.storage.googleapis.com',
      stackstate: 'https://helm.stackstate.io',
      stackstate_test: 'https://helm-test.stackstate.io',
    },
    charts: [
      'stable/hbase',
      'stable/stackstate',
    ],
  },
}
