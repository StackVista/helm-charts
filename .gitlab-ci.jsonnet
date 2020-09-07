// Imports
local variables = import '.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet';

// Shortcuts
local repositories = variables.helm.repositories;
local charts = variables.helm.charts;

local helm_fetch_dependencies = [
    'helm repo add %s %s' % [std.strReplace(name, '_', '-'), repositories[name]]
    for name in std.objectFields(repositories)
  ] +
  [
    'helm repo update',
  ];

local sync_charts_template = {
  before_script:
  ['.gitlab/push_before_script.sh'] + helm_fetch_dependencies,
  script: ['sh test/sync-repo.sh'],
  stage: 'build',
};

local validate_and_push_jobs = {
  validate_charts: {
    before_script: ['.gitlab/validate_before_script.sh'],
    environment: 'stseuw1-sandbox-main-eks-sandbox/${CI_COMMIT_REF_NAME}',
    rules: [
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        when: 'never',
      },
      { when: 'always' },
    ],
    script: [
      'ct list-changed --config test/ct.yaml',
      'ct lint --debug --validate-maintainers=false --config test/ct.yaml',
      '.gitlab/validate_kubeval.sh',
    ],
    stage: 'validate',
  },
  push_test_charts: sync_charts_template {
    rules: [
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        when: 'never',
      },
      {
        @'if': '$CI_COMMIT_TAG',
        when: 'never',
      },
      { when: 'always' },
    ],
    variables: {
      AWS_BUCKET: 's3://helm-test.stackstate.io',
      REPO_URL: 'https://helm-test.stackstate.io/',
    },
  },
};

local test_chart_job(chart) = {
  image: 'stackstate/stackstate-ci-images:stackstate-helm-test-e8e8e526',
  before_script: helm_fetch_dependencies +
  ['helm dependencies update ${CHART}'],
  script: [
    'go test ./stable/' + chart + '/test/...',
  ],
  stage: 'test',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: ['stable/' + chart + '/**/*'],
      exists: ['stable/' + chart + '/test/*.go'],
    },
  ],
  variables: {
    CHART: 'stable/' + chart,
    CGO_ENABLED: 0,
  },
};

local itest_chart_job(chart) = {
  image: 'stackstate/stackstate-ci-images:stackstate-helm-test-e8e8e526',
  before_script: helm_fetch_dependencies +
  ['helm dependencies update ${CHART}'],
  script: [
    'go test ./stable/' + chart + '/itest/...',
  ],
  stage: 'test',
  rules: [
    {
      @'if': '$CI_COMMIT_TAG',
      changes: ['stable/' + chart + '/**/*'],
      exists: ['stable/' + chart + '/itest/*.go'],
    },
  ],
  variables: {
    CHART: 'stable/' + chart,
    CGO_ENABLED: 0,
  },
};

local push_chart_job_if(chart, repository_url, repository_username, repository_password, rules) = {
  script: [
    'helm plugin install https://github.com/chartmuseum/helm-push.git',
    'helm dependencies update ${CHART}',
    'helm push --username ' + repository_username + ' --password ' + repository_password + ' ${CHART} ' + repository_url,
  ],
  rules: rules,
  variables: {
    CHART: 'stable/' + chart,
  },
};

local push_chart_job(chart, repository_url, repository_username, repository_password, when) =
  push_chart_job_if(
    chart,
    repository_url,
    repository_username,
    repository_password,
    [
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        changes: ['stable/' + chart + '/**/*'],
        when: when,
      },
    ]
  );

local push_stackstate_chart_releases =
{
 push_stackstate_release_to_internal: push_chart_job_if(
    'stackstate',
    '${CHARTMUSEUM_INTERNAL_URL}',
    '${CHARTMUSEUM_INTERNAL_USERNAME}',
    '${CHARTMUSEUM_INTERNAL_PASSWORD}',
    [
          {
            @'if': '$CI_COMMIT_TAG =~ /^release-\\d+\\.\\d+\\.\\d+(-.+)?\\+rc\\d+$/',
            when: 'on_success',
          },
          {
            @'if': '$CI_COMMIT_TAG =~ /^sts-private-v\\d+\\.\\d+\\.\\d+-[a-z]+$/',
            when: 'on_success',
          },
          {
            @'if': '$CI_COMMIT_TAG =~ /^sts-v\\d+\\.\\d+\\.\\d+$/',
            when: 'on_success',
          },
        ]
    ) {
    stage: 'push-charts-to-internal',
  },
  push_stackstate_release_to_public: push_chart_job_if(
    'stackstate',
    '${CHARTMUSEUM_URL}',
    '${CHARTMUSEUM_USERNAME}',
    '${CHARTMUSEUM_PASSWORD}',
    [
          {
            @'if': '$CI_COMMIT_TAG =~ /^sts-v\\d+\\.\\d+\\.\\d+$/',
            when: 'on_success',
          },
        ]
    ) {
    stage: 'push-charts-to-public',
  },
};

local test_chart_jobs = {
  ['test_%s' % chart]: (test_chart_job(chart))
  for chart in charts
};

local itest_stackstate = {
  integration_test_stackstate: itest_chart_job('stackstate'),
};

local push_charts_to_internal_jobs = {
  ['push_%s_to_internal' % chart]: (push_chart_job(chart,
      '${CHARTMUSEUM_INTERNAL_URL}',
'${CHARTMUSEUM_INTERNAL_USERNAME}',
'${CHARTMUSEUM_INTERNAL_PASSWORD}',
'on_success') {
    stage: 'push-charts-to-internal',
  })
  for chart in charts
};

local push_charts_to_public_jobs = {
  ['push_%s_to_public' % chart]: (push_chart_job(chart,
      '${CHARTMUSEUM_URL}',
'${CHARTMUSEUM_USERNAME}',
'${CHARTMUSEUM_PASSWORD}',
'manual') {
    stage: 'push-charts-to-public',

    needs: ['push_%s_to_internal' % chart],
  })
  for chart in charts
};

// Main
{
  // Only run for merge requests, tags, or the default (master) branch
  workflow: {
    rules: [
      { @'if': '$CI_MERGE_REQUEST_IID' },
      { @'if': '$CI_COMMIT_TAG' },
      { @'if': '$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH' },
    ],
  },
  image: 'quay.io/helmpack/chart-testing:v3.0.0-beta.2',
  stages: ['validate', 'test', 'build', 'push-charts-to-internal', 'push-charts-to-public'],

  variables: {
    HELM_VERSION: 'v3.1.2',
    KUBEVAL_SCHEMA_LOCATION: 'https://raw.githubusercontent.com/instrumenta/kubernetes-json-schema/master',
  },
}
+ test_chart_jobs
+ push_charts_to_internal_jobs
+ push_charts_to_public_jobs
+ validate_and_push_jobs
+ push_stackstate_chart_releases
+ itest_stackstate
