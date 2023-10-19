// Imports
local variables = import '.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet';

// Shortcuts
local repositories = variables.helm.repositories;
local charts = variables.helm.charts;
local public_charts = variables.helm.public_charts;

// resolving deps with deps, ct lint will not resolve that properly;
local update_2nd_degree_chart_deps(chart) = ['yq e \'.dependencies[] | select (.repository == "file*").repository | sub("^file://","")\' stable/' + chart + '/Chart.yaml  | xargs -I % helm dependencies build stable/' + chart + '/%'];

local helm_config_dependencies = [
    'helm repo add %s %s' % [std.strReplace(name, '_', '-'), repositories[name]]
    for name in std.objectFields(repositories)
  ];

local helm_fetch_dependencies = helm_config_dependencies +
  [
    'helm repo update',
  ];

local skip_when_dependency_upgrade = {
  rules: [{
    @'if': '$UPDATE_STACKGRAPH_VERSION',
    when: 'never',
  }, {
    @'if': '$UPDATE_AAD_CHART_VERSION',
    when: 'never',
  }] + super.rules,
};

local sync_charts_template = {
  before_script: helm_fetch_dependencies,
  script: [
    'source .gitlab/aws_auth_setup.sh',
    'sh test/sync-repo.sh',
    ],
  stage: 'build',
  image: variables.images.stackstate_devops,
} + skip_when_dependency_upgrade;


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
      'if [[ "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME}" =~ ^releasing*|^developing* ]] || [[ -n "${CI_COMMIT_TAG}" ]] ; then export VERSION_INCREMENT_CHECK="--check-version-increment=false" ; fi',
      'ct lint --debug --validate-maintainers=false ${VERSION_INCREMENT_CHECK} --excluded-charts stackstate --excluded-charts stackstate-k8s --excluded-charts gitlab-runner --config test/ct.yaml',
      '.gitlab/validate_kubeconform.sh',
    ],
    stage: 'validate',
  } + skip_when_dependency_upgrade,
  validate_stackstate_chart: {
    before_script: ['.gitlab/validate_before_script.sh'] + helm_config_dependencies,
    environment: 'stseuw1-sandbox-main-eks-sandbox/${CI_COMMIT_REF_NAME}',
    rules: [
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        when: 'never',
      },
      { when: 'always' },
    ],
    script:
      ['ct list-changed --config test/ct.yaml'] +
      update_2nd_degree_chart_deps('stackstate') +
      update_2nd_degree_chart_deps('stackstate-k8s') +
      [
        'ct lint --debug --validate-maintainers=false --charts stable/stackstate --charts stable/stackstate-k8s --config test/ct.yaml',
        '.gitlab/validate_kubeconform.sh',
      ],
    stage: 'validate',
  } + skip_when_dependency_upgrade,
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
  image: variables.images.stackstate_helm_test,
  before_script: helm_fetch_dependencies + (
    if chart == 'stackstate' || chart == 'stackstate-k8s' then update_2nd_degree_chart_deps(chart) else []
  ) +
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
} + skip_when_dependency_upgrade;

local itest_chart_job(chart) = {
  image: variables.images.stackstate_helm_test,
  before_script: helm_fetch_dependencies + (
    if chart == 'stackstate' || chart == 'stackstate-k8s' then update_2nd_degree_chart_deps(chart) else []
  ) +
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
} + skip_when_dependency_upgrade;

local push_chart_job_if(chart, repository_url, repository_username, repository_password, rules) = {
  script: (
    if chart == 'stackstate' || chart == 'stackstate-k8s' then update_2nd_degree_chart_deps(chart) else []
  ) + [
    'helm dependencies update ${CHART}',
    'helm cm-push --username ' + repository_username + ' --password ' + repository_password + ' ${CHART} ' + repository_url,
  ],
  image: variables.images.stackstate_devops,
  rules: rules,
  variables: {
    CHART: 'stable/' + chart,
  },
} + skip_when_dependency_upgrade;

local push_chart_job(chart, repository_url, repository_username, repository_password, when, autoTriggerOnCommitMsg) =
  push_chart_job_if(
    chart,
    repository_url,
    repository_username,
    repository_password,
    [
      {
        @'if': '$CI_COMMIT_BRANCH == "master" && $CI_COMMIT_AUTHOR == "stackstate-system-user <ops@stackstate.com>"  && $CI_COMMIT_MESSAGE =~ /\\[' + autoTriggerOnCommitMsg + ']/',
        changes: ['stable/' + chart + '/**/*'],
        when: 'on_success',
      },
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        changes: ['stable/' + chart + '/**/*'],
        when: when,
      },
      {
        @'if': '$CI_COMMIT_TAG =~ /^' + chart + '\\/.*/',
        when: 'on_success',
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
    variables.rules.tag.all_release_rules,
    ) {
    before_script: helm_fetch_dependencies,
    stage: 'push-charts-to-internal',
  },
  push_stackstate_release_to_public: push_chart_job_if(
    'stackstate',
    '${CHARTMUSEUM_URL}',
    '${CHARTMUSEUM_USERNAME}',
    '${CHARTMUSEUM_PASSWORD}',
    [variables.rules.tag.release_rule],
    ) {
    before_script: helm_fetch_dependencies,
    stage: 'push-charts-to-public',
  },
};

local test_chart_jobs = {
  ['test_%s' % chart]: (test_chart_job(chart))
  for chart in (charts + public_charts)
};

local itest_stackstate = {
  integration_test_stackstate: itest_chart_job('stackstate'),
  integration_test_stackstate_k8s: itest_chart_job('stackstate-k8s'),
};

local push_charts_to_internal_jobs = {
  ['push_%s_to_internal' % chart]: (push_chart_job(chart,
      '${CHARTMUSEUM_INTERNAL_URL}',
'${CHARTMUSEUM_INTERNAL_USERNAME}',
'${CHARTMUSEUM_INTERNAL_PASSWORD}',
'on_success',
 if chart == 'stackstate-k8s-agent' then 'publish-k8s-agent' else 'publish-' + chart) + {
    stage: 'push-charts-to-internal',
  } + (
  if chart == 'stackstate' then
  { before_script: helm_fetch_dependencies + ['.gitlab/bump_sts_chart_master_version.sh stackstate-internal ' + chart] }
  else if chart == 'stackstate-k8s' then
  { before_script: helm_fetch_dependencies + ['.gitlab/configure_git.sh', '.gitlab/bump_sts_chart_master_version_v2.sh stable/' + chart + " $(echo $CI_COMMIT_TAG | sed -E 's/^" + chart + "/(.*)$/\\1/')"] }  //It extracts version from tag, e.g. stackstate-k8s/1.3.2 => 1.3.2
  else {}
  ))
  for chart in (charts + public_charts)
};

local push_charts_to_public_jobs = {
  ['push_%s_to_public' % chart]: (push_chart_job(chart,
      '${CHARTMUSEUM_URL}',
'${CHARTMUSEUM_USERNAME}',
'${CHARTMUSEUM_PASSWORD}',
'manual',
if chart == 'stackstate-k8s-agent' then 'publish-k8s-agent' else 'publish-' + chart) + {
    stage: 'push-charts-to-public',

    needs: ['push_%s_to_internal' % chart],
  } + (
    if chart == 'stackstate' || chart == 'stackstate-k8s' then
  { before_script: helm_fetch_dependencies + ['.gitlab/bump_sts_chart_master_version.sh stackstate ' + chart] }
  else {}
  ))
  for chart in public_charts
  if chart != 'stackstate'
};

local update_sg_version = {
  update_stackgraph_version: {
    image: variables.images.stackstate_helm_test,
    stage: 'update',
    variables: {
      GIT_AUTHOR_EMAIL: 'sts-admin@stackstate.com',
      GIT_AUTHOR_NAME: 'stackstate-system-user',
      GIT_COMMITTER_EMAIL: 'sts-admin@stackstate.com',
      GIT_COMMITTER_NAME: 'stackstate-system-user',
    },
    before_script: helm_fetch_dependencies,
    rules: [
      {
        @'if': '$UPDATE_STACKGRAPH_VERSION',
        when: 'always',
      },
    ],
    script: [
      '.gitlab/update_sg_version.sh stable/hbase ""',
      '.gitlab/update_sg_version.sh stable/stackstate-k8s "hbase."',
      '.gitlab/update_chart_version.sh stable/stackstate-k8s hbase local:stable/hbase',
      '.gitlab/commit_changes_and_push.sh StackGraph $UPDATE_STACKGRAPH_VERSION',
    ],
  },
};

local update_aad_chart_version = {
  update_aad_chart_version: {
    image: variables.images.stackstate_helm_test,
    stage: 'update',
    variables: {
      GIT_AUTHOR_EMAIL: 'sts-admin@stackstate.com',
      GIT_AUTHOR_NAME: 'stackstate-system-user',
      GIT_COMMITTER_EMAIL: 'sts-admin@stackstate.com',
      GIT_COMMITTER_NAME: 'stackstate-system-user',
    },
    before_script: helm_fetch_dependencies,
    rules: [
      {
        @'if': '$UPDATE_AAD_CHART_VERSION',
        when: 'always',
      },
    ],
    script: [
      '.gitlab/update_chart_version.sh stable/stackstate-k8s anomaly-detection $UPDATE_AAD_CHART_VERSION',
      '.gitlab/commit_changes_and_push.sh anomaly-detection $UPDATE_AAD_CHART_VERSION',
    ],
  },
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
  image: variables.images.chart_testing,
  stages: ['validate', 'test', 'update', 'build', 'push-charts-to-internal', 'push-charts-to-public'],

  variables: {
    HELM_VERSION: 'v3.1.2',
    PROMOTION_DRY_RUN: 'no',
  },
}
+ test_chart_jobs
+ push_charts_to_internal_jobs
+ push_charts_to_public_jobs
+ validate_and_push_jobs
+ push_stackstate_chart_releases
+ itest_stackstate
+ update_sg_version
+ update_aad_chart_version
