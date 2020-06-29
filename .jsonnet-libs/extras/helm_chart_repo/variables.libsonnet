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
      'stable/aws-amicleaner',
      'stable/aws-nuke',
      'stable/cluster-agent',
      'stable/common',
      'stable/distributed',
      'stable/elasticsearch',
      'stable/gitlab-runner',
      'stable/gitlab-steward',
      'stable/hbase',
      'stable/iceman',
      'stable/stackstate',
      'stable/stackstate-standalone',
      'stable/trafficmirror',
    ],
  },
}
