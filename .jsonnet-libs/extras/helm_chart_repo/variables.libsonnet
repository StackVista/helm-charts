{
  local Variable = self,

  images: {
    stackstate_helm_test: '${DOCKER_PROXY_URL}/stackstate/sts-ci-images:stackstate-helm-test-1f9e87f1',
    stackstate_devops: '${DOCKER_PROXY_URL}/stackstate/sts-ci-images:stackstate-devops-1f9e87f1',
    chart_testing: 'quay.io/helmpack/chart-testing:v3.4.0',
  },
  helm: {
    repositories: {
      bitnami: 'https://charts.bitnami.com/bitnami',
      elastic: 'https://helm.elastic.co',
      stable: 'https://charts.helm.sh/stable',
      stackstate: 'https://helm.stackstate.io',
      stackstate_internal: 'https://helm-internal.stackstate.io',
      stackstate_test: 'https://helm-test.stackstate.io',
      prometheus: 'https://prometheus-community.github.io/helm-charts',
    },
    charts: [
      'aws-amicleaner',
      'aws-nuke',
      'aws-pod-identity-webhook',
      'beacher-job',
      'chartmuseum',
      'cluster-agent',
      'common',
      'elasticsearch',
      'gitlab-runner',
      'gitlab-steward',
      'hbase',
      'helm-reaper',
      'iceman',
      'kommoner-operator',
      'kafkaup-operator',
      'minio',
      'petros-d-kubelet-stats-exporter',
      'pull-secret',
      'sandbox-operator',
      'receiveramplifier',
      'stackstate-standalone',
      'stackstate-monitoring',
      'stackstate',
      'trafficmirror',
    ],
  },
  rules: {
    tag: {
      release_rule: {
        'if': '$CI_COMMIT_TAG =~ /^(\\d+)\\.(\\d+)\\.(\\d+)$/',
        when: 'on_success',
      },

      pre_release_rule: {
        'if': '$CI_COMMIT_TAG =~ /^(\\d+)\\.(\\d+)\\.(\\d+)-pre\\.(\\d+)$/',
        when: 'on_success',
      },

      rc_pre_release_rule: {
        'if': '$CI_COMMIT_TAG =~ /^(\\d+)\\.(\\d+)\\.(\\d+)-pre\\.(\\d+)\\.rc\\.(\\d+)$/',
        when: 'on_success',
      },

      rc_release_rule: {
        'if': '$CI_COMMIT_TAG =~ /^(\\d+)\\.(\\d+)\\.(\\d+)-rc\\.(\\d+)$/',
        when: 'on_success',
      },

      rc_rules: [
        Variable.rules.tag.rc_pre_release_rule,
        Variable.rules.tag.rc_release_rule,
      ],

      published_release_rules: [
        Variable.rules.tag.release_rule,
        Variable.rules.tag.pre_release_rule,
      ],

      all_release_rules: Variable.rules.tag.rc_rules + Variable.rules.tag.published_release_rules,
    },
  },
}
