{
  local Variable = self,

  images: {
    stackstate_helm_test: '${DOCKER_PROXY_URL}/stackstate/sts-ci-images:stackstate-helm-test-7f8a4dd6',
    stackstate_devops: '${DOCKER_PROXY_URL}/stackstate/sts-ci-images:stackstate-devops-0ccefe18',
    chart_testing: 'quay.io/helmpack/chart-testing:v3.6.0',
  },
  helm: {
    repositories: {
      bitnami: 'https://charts.bitnami.com/bitnami',
      'bitnami-pre-2022': 'https://raw.githubusercontent.com/bitnami/charts/eb5f9a9513d987b519f0ecd732e7031241c50328/bitnami',
      'bitnami-full-archive': 'https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami',
      elastic: 'https://helm.elastic.co',
      stable: 'https://charts.helm.sh/stable',
      stackstate: 'https://helm.stackstate.io',
      stackstate_internal: 'https://helm-internal.stackstate.io',
      stackstate_test: 'https://helm-test.stackstate.io',
      prometheus: 'https://prometheus-community.github.io/helm-charts',
    },
    public_charts: [
      'stackstate-agent',
      'stackstate-k8s-agent',
      'artifactory-cleaner',
      'common',
      'elasticsearch',
      'hbase',
      'kafkaup-operator',
      'minio',
      'pull-secret',
      'stackstate',
      'stackstate-k8s',
      'stackstate-monitoring',
      'stackstate-standalone',
      'trafficmirror',
      'victoria-metrics-single',
    ],
    charts: [
      'aws-amicleaner',
      'aws-nuke',
      'aws-pod-identity-webhook',
      'beacher-job',
      'chartmuseum',
      'gitlab-steward',
      'helm-reaper',
      'iceman',
      'k8s-image-list-to-s3',
      'k8s-spot-termination-handler',
      'kommoner-operator',
      'petros-d-kubelet-stats-exporter',
      'redirector',
      'receiveramplifier',
      'sandbox-operator',
      'sock-shop',
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
