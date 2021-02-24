{
  local Variable = self,

  helm: {
    repositories: {
      bitnami: 'https://charts.bitnami.com/bitnami',
      elastic: 'https://helm.elastic.co',
      minio: 'https://helm.min.io/',
      stable: 'https://charts.helm.sh/stable',
      stackstate: 'https://helm.stackstate.io',
      stackstate_internal: 'https://helm-internal.stackstate.io',
      stackstate_test: 'https://helm-test.stackstate.io',
    },
    charts: [
      'aws-amicleaner',
      'aws-nuke',
      'aws-pod-identity-webhook',
      'chartmuseum',
      'cluster-agent',
      'common',
      'elasticsearch',
      'gitlab-runner',
      'gitlab-steward',
      'hbase',
      'iceman',
      'kafka',
      'sandbox-operator',
      'stackstate',
      'stackstate-standalone',
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
