{
  local Variable = self,

  helm: {
    repositories: {
      bitnami: 'https://charts.bitnami.com/bitnami',
      elastic: 'https://helm.elastic.co',
      stable: 'https://charts.helm.sh/stable',
      stackstate: 'https://helm.stackstate.io',
      stackstate_internal: 'https://helm-internal.stackstate.io',
      stackstate_test: 'https://helm-test.stackstate.io',
    },
    charts: [
      'aws-amicleaner',
      'aws-nuke',
      'cluster-agent',
      'common',
      'elasticsearch',
      'gitlab-runner',
      'gitlab-steward',
      'hbase',
      'iceman',
      'stackstate',
      'stackstate-standalone',
      'trafficmirror',
    ],
  },
}
