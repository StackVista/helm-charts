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
  }, {
    @'if': '$UPDATE_STACKSTATE_DOCKER_VERSION',
    when: 'never',
  }, {
    @'if': '$UPDATE_STACKPACKS_DOCKER_VERSION',
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
      'ct lint --debug --validate-maintainers=false ${VERSION_INCREMENT_CHECK} --excluded-charts stackstate --excluded-charts suse-observability --excluded-charts gitlab-runner --config test/ct.yaml',
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
      update_2nd_degree_chart_deps('suse-observability') +
      [
        'ct lint --debug --validate-maintainers=false --charts stable/stackstate --charts stable/suse-observability --config test/ct.yaml',
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
    if chart == 'stackstate' || chart == 'suse-observability' then update_2nd_degree_chart_deps(chart) else []
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
    if chart == 'stackstate' || chart == 'suse-observability' then update_2nd_degree_chart_deps(chart) else []
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

local push_chart_job_if(chart, script, rules) = {
  script: script,
  image: variables.images.stackstate_devops,
  rules: rules,
  variables: {
    CHART: 'stable/' + chart,
  },
} + skip_when_dependency_upgrade;

local push_chart_job(chart, script, when, autoTriggerOnCommitMsg) =
  push_chart_job_if(
    chart,
    script,
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

local push_chart_script(chart, repository_url, repository_username, repository_password) =

   (if chart == 'stackstate' || chart == 'suse-observability' then update_2nd_degree_chart_deps(chart) else [])
   + [
    'helm dependencies update ${CHART}',
    'helm cm-push --username ' + repository_username + ' --password ' + repository_password + ' ${CHART} ' + repository_url,
  ];

local push_stackstate_chart_releases =
{
 push_stackstate_release_to_internal: push_chart_job_if(
    'stackstate',
    push_chart_script(
      'stackstate',
      '${CHARTMUSEUM_INTERNAL_URL}',
      '${CHARTMUSEUM_INTERNAL_USERNAME}',
      '${CHARTMUSEUM_INTERNAL_PASSWORD}',
    ),
    variables.rules.tag.all_release_rules,
    ) {
    before_script: helm_fetch_dependencies,
    stage: 'push-charts-to-internal',
  },
  push_stackstate_release_to_public: push_chart_job_if(
    'stackstate',
    push_chart_script(
'stackstate',
      '${CHARTMUSEUM_URL}',
      '${CHARTMUSEUM_USERNAME}',
      '${CHARTMUSEUM_PASSWORD}',
    ),
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
  integration_test_stackstate_k8s: itest_chart_job('suse-observability'),
};

local push_charts_to_internal_jobs = {
  ['push_%s_to_internal' % chart]: (push_chart_job(
    chart,
    push_chart_script(
chart,
    '${CHARTMUSEUM_INTERNAL_URL}',
    '${CHARTMUSEUM_INTERNAL_USERNAME}',
    '${CHARTMUSEUM_INTERNAL_PASSWORD}',
    ),
'on_success',
 if chart == 'suse-observability-agent' then 'publish-suse-observability-agent' else if chart == 'stackstate-k8s-agent' then 'publish-k8s-agent' else 'publish-' + chart
) + {
    stage: 'push-charts-to-internal',
  } + (
  if chart == 'stackstate' then
  { before_script: helm_fetch_dependencies + ['.gitlab/bump_sts_chart_master_version.sh stackstate-internal ' + chart] }
  else if chart == 'suse-observability' then
  { before_script: helm_fetch_dependencies + [
    '.gitlab/configure_git.sh',
    // tags don't have CI_COMMIT_BRANCH, so I fetches the current branch(s) for current HEAD (HEAD points to a detached commit)
    // but there may be multiple branches so I iterate all of them and push a commit to each branch
    'export BRANCHES=${CI_COMMIT_BRANCH:-$(git for-each-ref --format="%(objectname) %(refname:short)" refs/remotes/origin | awk -v branch="$(git rev-parse HEAD)" \'$1==branch && $2!="origin" {print $2}\' | sed -E "s/^origin\\/(.*)$/\\1/")}',
    // It extracts version from tag, e.g. suse-observability/1.3.2 => 1.3.2
    '.gitlab/set_sts_chart_master_version.sh stable/' + chart + " $(echo $CI_COMMIT_TAG | sed -E 's/^" + chart + "\\/(.*)$/\\1/')",
  ] }
  { script+: ['.gitlab/bump_sts_chart_master_version_v2.sh stable/' + chart] }
  else {}
  ))
  for chart in (charts + public_charts)
};

local push_charts_to_public_jobs = {
  ['push_%s_to_public' % chart]: (push_chart_job(
    chart,
    push_chart_script(
chart,
      '${CHARTMUSEUM_URL}',
'${CHARTMUSEUM_USERNAME}',
'${CHARTMUSEUM_PASSWORD}',
    ),
'manual',
if chart == 'suse-observability-agent' then 'publish-suse-observability-agent' else if chart == 'stackstate-k8s-agent' then 'publish-k8s-agent' else 'publish-' + chart
) + {
    stage: 'push-charts-to-public',

    needs: ['push_%s_to_internal' % chart],
  } + (
    if chart == 'stackstate' || chart == 'suse-observability' then
  { before_script: helm_fetch_dependencies + ['.gitlab/bump_sts_chart_master_version.sh stackstate ' + chart] }
  else {}
  ))
  for chart in public_charts
  if chart != 'stackstate' && chart != 'suse-observability'
};

local push_suse_observability_to_rancher_registry = {
  'push_suse-observability-agent_to_rancher': (push_chart_job(
    'suse-observability-agent',
    [
      '.gitlab/publish-suse-agent-to-rancher.sh',
    ],
    'manual',
    'publish-suse-observability-agent',
) + {
    stage: 'push-charts-to-rancher',
    needs: ['push_suse-observability-agent_to_internal'],
  }),
  'push_suse-observability-values_to_rancher': (push_chart_job(
    'suse-observability-values',
    [
      '.gitlab/publish-suse-observability-values-to-rancher.sh',
    ],
    'manual',
    'publish-suse-observability-values',
) + {
    stage: 'push-charts-to-rancher',
    needs: ['push_suse-observability-values_to_internal'],
  }),
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
      '.gitlab/update_sg_version.sh stable/suse-observability "hbase."',
      '.gitlab/update_chart_version.sh stable/suse-observability hbase local:stable/hbase',
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
      '.gitlab/update_chart_version.sh stable/suse-observability anomaly-detection $UPDATE_AAD_CHART_VERSION',
      '.gitlab/commit_changes_and_push.sh anomaly-detection $UPDATE_AAD_CHART_VERSION',
    ],
  },
};

local update_docker_images = {
  local job(requiredEnvName, scripts) = {
    image: variables.images.stackstate_devops,
    stage: 'update',
    variables: {
      GIT_AUTHOR_EMAIL: 'sts-admin@stackstate.com',
      GIT_AUTHOR_NAME: 'stackstate-system-user',
      GIT_COMMITTER_EMAIL: 'sts-admin@stackstate.com',
      GIT_COMMITTER_NAME: 'stackstate-system-user',
    },
    before_script: [
      '.gitlab/configure_git.sh',
      // tags don't have CI_COMMIT_BRANCH, so fetch the current branch(s) for current HEAD (HEAD points to a detached commit)
      // but there may be multiple branches so iterate all of them and push a commit to each branch
      'export BRANCHES=${CI_COMMIT_BRANCH:-$(git for-each-ref --format="%(objectname) %(refname:short)" refs/remotes/origin | awk -v branch="$(git rev-parse HEAD)" \'$1==branch && $2!="origin" {print $2}\' | sed -E "s/^origin\\/(.*)$/\\1/")}',
    ],
    rules: [
      {
        @'if': '$' + requiredEnvName,
        when: 'always',
      },
      {
        when: 'never',
      },
    ],
    script: scripts,
  },

  update_stackstate_version_to_latest: job('UPDATE_STACKSTATE_DOCKER_VERSION', ['.gitlab/suse-observability/update_stackstate_version_to_latest.sh']),
  update_stackpacks_version_to_latest: job('UPDATE_STACKPACKS_DOCKER_VERSION', ['.gitlab/suse-observability/update_stackpacks_version_to_latest.sh']),
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
  stages: ['validate', 'test', 'update', 'build', 'push-charts-to-internal', 'push-charts-to-public', 'push-charts-to-rancher'],

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
+ update_docker_images
+ push_suse_observability_to_rancher_registry
