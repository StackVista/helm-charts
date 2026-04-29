{
  local Variable = self,

  images: {
    stackstate_helm_test: '${REGISTRY_QUAY_URL}/stackstate/sts-ci-images:stackstate-helm-test-4423bdaa',
    stackstate_devops: '${REGISTRY_QUAY_URL}/stackstate/sts-ci-images:stackstate-devops-fd4b135a',
    chart_testing: 'quay.io/helmpack/chart-testing:v3.10.1',
    container_tools_dev: 'quay.io/stackstate/container-tools:1.8.6_dev-fa52bb17-main-4',
  },
  helm: {
    repositories: {
      deliveryhero: 'https://charts.deliveryhero.io',
      opentelemetry: 'https://open-telemetry.github.io/opentelemetry-helm-charts',
      stackstate: 'https://helm.stackstate.io',
      victoriametrics: 'https://victoriametrics.github.io/helm-charts',
    },
    public_charts: [
      'suse-observability-agent',
      'artifactory-cleaner',
      'anomaly-detection',
      'clickhouse',
      'common',
      'elasticsearch',
      'hbase',
      'kafka',
      'kafkaup-operator',
      'kubernetes-rbac-agent',
      'prometheus-elasticsearch-exporter',
      'pull-secret',
      'suse-observability',
      'suse-observability-values',
      'stackstate-monitoring',
      'trafficmirror',
      'victoria-metrics-single',
      'zookeeper',
    ],
    charts: [
      'aws-pod-identity-webhook',
      'beacher-job',
      'ci-test',
      'chartmuseum',
      'gitlab-proxy',
      'gitlab-steward',
      'helm-reaper',
      'k8s-spot-termination-handler',
      'kommoner-operator',
      'notification-operator',
      'otel-demo',
      'opentelemetry-collector',
      'petros-d-kubelet-stats-exporter',
      'redirector',
      'receiveramplifier',
      'sock-shop',
      'tenantprovisioning',
      'tenantmanagement',
      'tenant-cleanup-controller',
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
