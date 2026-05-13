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
      'suse-observability',
      'suse-observability-agent',
      'suse-observability-values',
    ],
    charts: [
      'artifactory-cleaner',
      'aws-pod-identity-webhook',
      'anomaly-detection',
      'beacher-job',
      'ci-test',
      'chartmuseum',
      'clickhouse',
      'common',
      'elasticsearch',
      'gitlab-proxy',
      'gitlab-steward',
      'hbase',
      'helm-reaper',
      'kafka',
      'kafkaup-operator',
      'k8s-spot-termination-handler',
      'kommoner-operator',
      'kubernetes-rbac-agent',
      'notification-operator',
      'otel-demo',
      'opentelemetry-collector',
      'petros-d-kubelet-stats-exporter',
      'pull-secret',
      'prometheus-elasticsearch-exporter',
      'redirector',
      'receiveramplifier',
      'sock-shop',
      'stackstate-monitoring',
      'tenantprovisioning',
      'tenantmanagement',
      'tenant-cleanup-controller',
      'trafficmirror',
      'victoria-metrics-single',
      'zookeeper',
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
