// Imports
local variables = import '.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet';

// Shortcuts
local repositories = variables.helm.repositories;
local public_charts = variables.helm.public_charts;  // map: parent -> [local deps]
local internal_charts = variables.helm.internal_charts;  // map: parent -> [local deps]

// All chart names that are independently published (keys of both maps).
local public_chart_names = std.objectFields(public_charts);
local internal_chart_names = std.objectFields(internal_charts);
local published_charts = public_chart_names + internal_chart_names;

// All local-only chart names (deduped across both maps).
local local_charts = std.set(std.flattenArrays(
  [public_charts[p] for p in public_chart_names] +
  [internal_charts[p] for p in internal_chart_names]
));

// Local deps for a published chart (empty when the chart has none listed).
local local_deps_of(chart) =
  if std.objectHas(public_charts, chart) then public_charts[chart]
  else if std.objectHas(internal_charts, chart) then internal_charts[chart]
  else [];

// CI change paths for a published chart: its own tree plus every local dep's tree.
local change_paths_for_published(chart) =
  ['stable/' + chart + '/**/*'] +
  ['local/' + d + '/**/*' for d in local_deps_of(chart)];

// Files that, when modified, should re-run every chart's build/validate jobs.
// A change to the generated CI config can change job definitions in non-obvious
// ways, so we want to revalidate everything to make sure nothing regresses.
local ci_config_change_paths = ['.gitlab-ci.yml'];

local go_cache = { key: { files: ['go.mod', 'go.sum'] }, paths: ['/go/pkg/mod/'] };

// Build 2nd-degree file:// chart deps. chartPath is e.g. stable/suse-observability or local/hbase;
// helm cannot resolve nested file:// deps automatically, so we walk them once.
local update_2nd_degree_chart_deps(chartPath) = [
  'yq e \'.dependencies[] | select (.repository == "file*").repository | sub("^file://","")\' ' +
  chartPath + '/Chart.yaml | xargs -I % helm dependencies build --skip-refresh ' + chartPath + '/%',
];

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
  }, {
    @'if': '$RUN_UPDATECLI',
    when: 'never',
  }, {
    @'if': '$RUN_UPDATECLI_STACKSTATE',
    when: 'never',
  }, {
    @'if': '$RUN_UPDATECLI_STACKPACKS',
    when: 'never',
  }] + super.rules,
};

// Build a published chart with all dependencies and then share it as an artifact with next jobs in
// the pipeline. This is the most expensive step so we want to execute it only once.
local build_chart_job(chart) = {
  image: variables.images.stackstate_helm_test,
  before_script: helm_fetch_dependencies,
  script:
    update_2nd_degree_chart_deps('stable/' + chart) + [
      'helm dependencies build stable/' + chart,
      'mkdir -p stable/' + chart + '/build; helm package --destination stable/' + chart + '/build stable/' + chart,
    ],
  stage: 'build',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: change_paths_for_published(chart) + ci_config_change_paths,
    },
  ],
  artifacts: {
    paths: [
      'stable/' + chart + '/charts/',
      'stable/' + chart + '/build/',
    ],
  },
};
local build_chart_jobs = {
  ['build_%s' % chart]: (build_chart_job(chart))
  for chart in published_charts
};

// Build a local chart's file:// deps so downstream validate/test jobs can resolve subchart
// templates. No packaging or publishing — only the charts/ directory is shared as an artifact.
local build_local_chart_job(chart) = {
  image: variables.images.stackstate_helm_test,
  before_script: helm_fetch_dependencies,
  script:
    update_2nd_degree_chart_deps('local/' + chart) + [
      'helm dependencies build local/' + chart,
    ],
  stage: 'build',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: ['local/' + chart + '/**/*'] + ci_config_change_paths,
    },
  ],
  artifacts: {
    paths: [
      'local/' + chart + '/charts/',
    ],
  },
};
local build_local_chart_jobs = {
  ['build_local_%s' % chart]: (build_local_chart_job(chart))
  for chart in local_charts
};

// Validates modified published charts: schema check, lint, kubeconform.
local validate_chart_job(chart) = {
  image: variables.images.chart_testing,
  before_script: ['.gitlab/validate_before_script.sh'],
  script: [
    'yamale --schema /etc/ct/chart_schema.yaml stable/' + chart + '/Chart.yaml',
    'yamllint --config-file /etc/ct/lintconf.yaml stable/' + chart + '/Chart.yaml',
    'yamllint --config-file /etc/ct/lintconf.yaml stable/' + chart + '/values.yaml',
    'if [ -f stable/' + chart + '/ci/default-values.yaml ]; then yamllint --config-file /etc/ct/lintconf.yaml stable/' + chart + '/ci/default-values.yaml; fi',
    'if [ -f stable/' + chart + '/ci/default-values.yaml ]; then helm lint stable/' + chart + ' --values stable/' + chart + '/ci/default-values.yaml; fi',
    '.gitlab/validate_local_dep_versions.sh stable/' + chart,
    '.gitlab/validate_kubeconform.sh stable/' + chart,
  ],
  stage: 'validate',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: change_paths_for_published(chart) + ci_config_change_paths,
    },
  ],
};
local validate_chart_jobs = {
  ['validate_%s' % chart]: (validate_chart_job(chart))
  for chart in published_charts
};

// Validates that every docker image referenced by a chart resolves in its
// registry. Image enumeration is delegated to the chart's
// installation/o11y-*-get-images.sh; only the charts that ship one are wired up.
local charts_with_image_validation = ['suse-observability', 'suse-observability-agent'];
local validate_chart_images_job(chart) = {
  image: variables.images.stackstate_devops,
  script: ['.gitlab/validate_chart_images.sh stable/' + chart],
  stage: 'validate',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: change_paths_for_published(chart) + ci_config_change_paths,
    },
  ],
};
local validate_chart_images_jobs = {
  ['validate_%s_images' % chart]: (validate_chart_images_job(chart))
  for chart in charts_with_image_validation
};

// Lightweight validation for local-only charts: schema + lint. No version check, no publish.
local validate_local_chart_job(chart) = {
  image: variables.images.chart_testing,
  before_script: ['.gitlab/validate_before_script.sh'],
  script: [
    'yamale --schema /etc/ct/chart_schema.yaml local/' + chart + '/Chart.yaml',
    'yamllint --config-file /etc/ct/lintconf.yaml local/' + chart + '/Chart.yaml',
    'if [ -f local/' + chart + '/values.yaml ]; then yamllint --config-file /etc/ct/lintconf.yaml local/' + chart + '/values.yaml; fi',
    'if [ -f local/' + chart + '/ci/default-values.yaml ]; then yamllint --config-file /etc/ct/lintconf.yaml local/' + chart + '/ci/default-values.yaml; fi',
    'if [ -f local/' + chart + '/ci/default-values.yaml ]; then helm lint local/' + chart + ' --values local/' + chart + '/ci/default-values.yaml; fi',
    '.gitlab/validate_local_dep_versions.sh local/' + chart,
    '.gitlab/validate_kubeconform.sh local/' + chart,
  ],
  stage: 'validate',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: ['local/' + chart + '/**/*'] + ci_config_change_paths,
    },
  ],
};
local validate_local_chart_jobs = {
  ['validate_local_%s' % chart]: (validate_local_chart_job(chart))
  for chart in local_charts
};

// Checks if a chart version has been updated, if not then return an error.
// Only runs for published charts — local charts are pinned to "*" and never need version bumps.
local check_chart_version_job(chart) = {
  image: variables.images.chart_testing,
  before_script: ['.gitlab/validate_before_script.sh'],
  script: [
    '.gitlab/verify_versions_bumped.sh ' + chart,
  ],
  stage: 'validate',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: ['stable/' + chart + '/**/*'],
    },
  ],
};
local check_chart_version_jobs = {
  ['check_%s_version' % chart]: (check_chart_version_job(chart))
  for chart in published_charts
  if chart != 'suse-observability'
};

// Runs unit tests on a published chart when it or any of its local deps change.
local test_chart_job(chart) = {
  image: variables.images.stackstate_helm_test,
  tags: ['sts-k8s-xl-runner'],
  script: [
    'go mod download',
    'go test ./stable/' + chart + '/test/...',
  ],
  stage: 'test',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: change_paths_for_published(chart),
      exists: ['stable/' + chart + '/test/*.go'],
    },
  ],
  variables: {
    CGO_ENABLED: 0,
    GOPATH: '/go',
  },
  cache: go_cache,
};
local test_chart_jobs = {
  ['test_%s' % chart]: (test_chart_job(chart))
  for chart in published_charts
};

// Runs unit tests on a local chart in isolation.
local test_local_chart_job(chart) = {
  image: variables.images.stackstate_helm_test,
  tags: ['sts-k8s-xl-runner'],
  script: [
    'go mod download',
    'go test ./local/' + chart + '/test/...',
  ],
  stage: 'test',
  rules: [
    {
      @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
      changes: ['local/' + chart + '/**/*'],
      exists: ['local/' + chart + '/test/*.go'],
    },
  ],
  variables: {
    CGO_ENABLED: 0,
    GOPATH: '/go',
  },
  cache: go_cache,
};
local test_local_chart_jobs = {
  ['test_local_%s' % chart]: (test_local_chart_job(chart))
  for chart in local_charts
};

local resource_usage = {
  resource_usage: {
    image: variables.images.stackstate_helm_test,
    tags: ['sts-k8s-xl-runner'],
    before_script: helm_fetch_dependencies,
    script: update_2nd_degree_chart_deps('stable/suse-observability') + [
      'helm dependencies build stable/suse-observability',
      'helm dependencies build stable/suse-observability-values',
      'go mod download',
      'go test ./test/...',
    ],
    stage: 'test',
    rules: [
      {
        @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
        changes: [
          'test/*.go',
          'test/**/*',
          'stable/suse-observability-values/**/*',
        ] + change_paths_for_published('suse-observability'),
      },
    ],
    variables: {
      CGO_ENABLED: 0,
      GOPATH: '/go',
    },
    artifacts: {
      paths: [
        'test/resource_usage.txt',
      ],
    },
    cache: go_cache,
  },
};

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
        @'if': '$CI_COMMIT_BRANCH == "master" && $CI_COMMIT_AUTHOR == "stackstate-system-user <suse-observability-ops@stackstate.com>"  && $CI_COMMIT_MESSAGE =~ /\\[' + autoTriggerOnCommitMsg + ']/',
        changes: change_paths_for_published(chart),
        when: 'on_success',
      },
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        changes: change_paths_for_published(chart),
        when: when,
      },
      {
        @'if': '$CI_COMMIT_TAG =~ /^' + chart + '\\/.*/',
        when: 'on_success',
      },
    ]
  );

local push_chart_script(chart, repository_url, repository_username, repository_password) =
  (if chart == 'suse-observability' then update_2nd_degree_chart_deps('stable/' + chart) else [])
  + [
    'helm dependencies update --skip-refresh ${CHART}',
    'helm cm-push --username ' + repository_username + ' --password ' + repository_password + ' ${CHART} ' + repository_url,
  ];

local push_prerelease_chart_script(chart, repository_url, repository_username, repository_password) =
  (if chart == 'suse-observability' then update_2nd_degree_chart_deps('stable/' + chart) else [])
  + [
    'helm dependencies update --skip-refresh ${CHART}',
    '.gitlab/modify_chart_to_prerelease_version.sh ${CHART}',
    'helm cm-push --username ' + repository_username + ' --password ' + repository_password + ' ${CHART} ' + repository_url,
  ];

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
                                      'publish-' + chart
                                    ) + {
                                      stage: 'push-charts-to-internal',
                                    } + (
if chart == 'suse-observability' then
                                           { before_script: helm_fetch_dependencies + [
                                             '.gitlab/configure_git.sh',
                                             // tags don't have CI_COMMIT_BRANCH, so I fetches the current branch(s) for current HEAD (HEAD points to a detached commit)
                                             // but there may be multiple branches so I iterate all of them and push a commit to each branch
                                             'export BRANCHES=${CI_COMMIT_BRANCH:-$(git for-each-ref --format="%(objectname) %(refname:short)" refs/remotes/origin | awk -v branch="$(git rev-parse HEAD)" \'$1==branch && $2!="origin" {print $2}\' | sed -E "s/^origin\\/(.*)$/\\1/")}',
                                             // It extracts version from tag, e.g. suse-observability/1.3.2 => 1.3.2. This is required to make sure the tagged version gets released to stackstate-internal
                                             // during the consecutive cm-push executed in the `script:` section.
                                             '.gitlab/set_chart_release_version.sh stable/' + chart + " $(echo $CI_COMMIT_TAG | sed -E 's/^" + chart + "\\/(.*)$/\\1/')",
                                           ] }
                                           { script+: [
                                             '.gitlab/tag_sts_chart_pre_release.sh ' + chart,
                                             '.gitlab/bump_suse_chart_pre_master_version.sh stable/' + chart,
                                           ] }
                                         else { before_script: helm_fetch_dependencies }
                                       ))
  for chart in published_charts
};

local push_prerelease_charts_to_public_jobs = {
  ['push_prerelease_%s_to_public' % chart]: (push_chart_job(
                                    chart,
                                    push_prerelease_chart_script(
                                      chart,
                                      '${CHARTMUSEUM_URL}',
                                      '${CHARTMUSEUM_USERNAME}',
                                      '${CHARTMUSEUM_PASSWORD}',
                                    ),
                                    'manual',
                                    'publish-' + chart
                                  ) + {
                                    stage: 'push-charts-to-public',

                                    needs: ['push_%s_to_internal' % chart],

                                    before_script: helm_fetch_dependencies,
                                  })
  for chart in public_chart_names
  if chart != 'suse-observability'
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
                                    'publish-' + chart
                                  ) + {
                                    stage: 'push-charts-to-public',

                                    needs: ['push_prerelease_%s_to_public' % chart],

                                    before_script: helm_fetch_dependencies,
                                  })
  for chart in public_chart_names
  if chart != 'suse-observability'
};

local push_suse_observability_to_rancher_registry = {
  'push_suse-observability-agent_prerelease_to_rancher': (push_chart_job(
                                                      'suse-observability-agent',
                                                      [
                                                        '.gitlab/publish-suse-agent-to-rancher.sh',
                                                      ],
                                                      'manual',
                                                      'publish-suse-observability-agent',
                                                    ) + {
                                                      stage: 'push-charts-to-rancher',
                                                      needs: [
                                                      'push_suse-observability-agent_to_internal',
                                                      ],
                                                    }),
  'push_suse-observability-agent_to_rancher': (push_chart_job(
                                                 'suse-observability-agent',
                                                 [
                                                   '.gitlab/publish-suse-agent-to-rancher.sh release',
                                                 ],
                                                 'manual',
                                                 'publish-suse-observability-agent',
                                               ) + {
                                                 stage: 'push-charts-to-rancher',
                                                 needs: ['push_suse-observability-agent_prerelease_to_rancher'],
                                               }),
};

// These variables are actually explicitly unset in a default pipeline. The reason for this is that we want to be very specific in a triggered job what
// we want as source and target, to aovid accidentally pushing/manipulating master when doing work/testing on these pipelines
local updatecli_variables = {
  TARGET_BRANCH: 'unset',
  STACKSTATE_SOURCE_BRANCH: 'unset',
  STACKPACKS_SOURCE_BRANCH: 'unset',
};

// updatecli pipelines that have a CI job (apply, plus optional open-MR).
// Adding an entry here also extends validate_updatecli_config /
// validate_updatecli_diff to cover the new pipeline.
//
// directPush: how to land updatecli's changes after the (uniform) apply job:
//   true  -> .gitlab/updatecli_squash_push.sh squash-merges UPDATECLI_BRANCH
//            into TARGET_BRANCH as one signed commit, then deletes the
//            working branch. No MR; `title` is unused, `commitMessage`
//            is required.
//   false -> .gitlab/open_updatecli_mr.sh opens an MR from UPDATECLI_BRANCH
//            into TARGET_BRANCH using `title`. `commitMessage` is unused.
local updatecli_pipelines = [
  {
    // id is used in the GitLab job name (underscored); pipelineid must match
    // the `pipelineid:` field in the corresponding updatecli yaml because
    // that's what updatecli uses to build its working branch name.
    id: 'docker_images',
    pipelineid: 'docker-images',
    dir: 'update-docker-images',
    triggerVar: 'RUN_UPDATECLI',
    directPush: false,
    title: 'Bump helm chart docker images',
  },
  {
    id: 'stackstate',
    pipelineid: 'stackstate',
    dir: 'update-stackstate',
    triggerVar: 'RUN_UPDATECLI_STACKSTATE',
    directPush: true,
    commitMessage: 'Bump StackState and StackGraph versions',
  },
  {
    id: 'stackpacks',
    pipelineid: 'stackpacks',
    dir: 'update-stackpacks',
    triggerVar: 'RUN_UPDATECLI_STACKPACKS',
    directPush: true,
    commitMessage: 'Bump StackPacks version',
  },
];

// Common skeleton for both apply and follow-up jobs in the updatecli flow.
local updatecli_job_base(p) = {
  image: variables.images.stackstate_devops,
  stage: 'update',
  variables: updatecli_variables {
    // Updatecli's gitlab SCM names its working branch
    //   updatecli-<branch>-<pipelineid>
    // (the dash separator comes from workingbranchseparator in the pipeline
    // yaml). The follow-up job needs this name to open the MR or to
    // squash-merge it into TARGET_BRANCH.
    UPDATECLI_BRANCH: 'updatecli-${TARGET_BRANCH}-' + p.pipelineid,
  },
  before_script: [
    '.gitlab/configure_git.sh',
    'export GITLAB_TOKEN="$HELM_CHARTS_PAT"',
  ],
  rules: [
    {
      @'if': '$' + p.triggerVar,
      when: 'always',
    },
    {
      when: 'never',
    },
  ],
};

local updatecli_job_pair(p) = {
  // Uniform apply job for every pipeline: updatecli runs normally and pushes
  // its per-target commits to UPDATECLI_BRANCH (working branch correctly
  // based on TARGET_BRANCH because workingbranch defaults to true).
  ['update_helm_chart_' + p.id]: updatecli_job_base(p) + {
    variables+: {
      UPDATE_CLI_EMAIL: '$STACKSTATE_SYSTEM_USER_EMAIL',
      UPDATE_CLI_USER: '$STACKSTATE_SYSTEM_USER_NAME',
    },
    before_script+: [
      'export UPDATE_CLI_PGP_KEY="$(cat $STACKSTATE_SYSTEM_USER_PGP_KEY)"',
      'export UPDATE_CLI_PGP_PASSPHRASE="$STACKSTATE_SYSTEM_USER_PGP_PASS_PHRASE"',
    ],
    script: [
      'updatecli pipeline apply -c updatecli/updatecli.d/' + p.dir + '/ -v updatecli/values.d/values.yaml',
    ],
  },
} + (
  if p.directPush then {
    // Follow-up: squash-merge UPDATECLI_BRANCH into TARGET_BRANCH as a single
    // signed commit, then delete the working branch.
    ['squash_push_updatecli_' + p.id]: updatecli_job_base(p) + {
      needs: ['update_helm_chart_' + p.id],
      script: [
        '.gitlab/updatecli_squash_push.sh "$UPDATECLI_BRANCH" "$TARGET_BRANCH" "' + p.commitMessage + '"',
      ],
    },
  } else {
    // Follow-up: open an MR from UPDATECLI_BRANCH into TARGET_BRANCH.
    ['open_updatecli_' + p.id + '_mr']: updatecli_job_base(p) + {
      needs: ['update_helm_chart_' + p.id],
      script: [
        '.gitlab/open_updatecli_mr.sh "$UPDATECLI_BRANCH" "$TARGET_BRANCH" "' + p.title + '"',
      ],
    },
  }
);

local updatecli_job = std.foldl(
  function(acc, p) acc + updatecli_job_pair(p),
  updatecli_pipelines,
  {},
);

local validate_updatecli_config = {
  validate_updatecli_config: {
    image: variables.images.chart_testing,
    stage: 'validate',
    variables: updatecli_variables,
    script: [
      'echo "Updatecli config changes detected — validating file structure"',
      'test -f updatecli/values.d/values.yaml',
      'yamllint -d "{rules: {document-start: disable, line-length: disable}}" updatecli/values.d/values.yaml',
    ] + [
      'test -f updatecli/updatecli.d/' + p.dir + '/update.yaml'
      for p in updatecli_pipelines
    ],
    rules: [
      {
        @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
        changes: ['updatecli/**/*'],
      },
    ],
  },
  validate_updatecli_diff: {
    image: variables.images.stackstate_devops,
    stage: 'validate',
    // diff needs real branch values so updatecli can clone and run the
    // sources. TARGET_BRANCH follows the MR's actual target so the diff is
    // computed against the helm-charts branch that would receive the change;
    // source branches stay on master since that's where the upstream tags
    // we validate against live.
    variables: {
      TARGET_BRANCH: '$CI_MERGE_REQUEST_TARGET_BRANCH_NAME',
      STACKSTATE_SOURCE_BRANCH: 'master',
      STACKPACKS_SOURCE_BRANCH: 'master',
    },
    script: [
      'export GITLAB_TOKEN="$HELM_CHARTS_PAT"',
    ] + [
      'updatecli pipeline diff --config updatecli/updatecli.d/' + p.dir + '/ --values updatecli/values.d/values.yaml'
      for p in updatecli_pipelines
    ],
    rules: [
      {
        @'if': '$CI_PIPELINE_SOURCE == "merge_request_event"',
        changes: ['updatecli/**/*'],
      },
    ],
  },
};

local beest_triggers = {
  beest_agent_trigger: {
    image: variables.images.stackstate_devops,
    stage: 'update',
    script: [
      'AGENT_HELM_CHART_VERSION=$(yq e ".version" stable/suse-observability-agent/Chart.yaml)',
      'echo "Triggering beest pipeline with AGENT_HELM_CHART_VERSION=${AGENT_HELM_CHART_VERSION}"',
      'curl --verbose --fail-with-body --request POST --form "token=$STS_BEEST_TRIGGER_TOKEN" --form ref=main --form "variables[AGENT_HELM_CHART_VERSION]=${AGENT_HELM_CHART_VERSION}" --form "variables[TRIGGER_AGENT_X86_TESTS]=true" "https://gitlab.com/api/v4/projects/$BEEST_PROJECT_ID/trigger/pipeline"',
    ],
    rules: [
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        changes: change_paths_for_published('suse-observability-agent'),
        when: 'manual',
      },
      {
        @'if': '$CI_COMMIT_BRANCH',
        changes: change_paths_for_published('suse-observability-agent'),
        when: 'manual',
      },
    ],
    allow_failure: true,
  },

  beest_stackstate_trigger: {
    image: variables.images.stackstate_devops,
    stage: 'update',
    script: [
      'STACKSTATE_HELM_CHART_VERSION=$(yq e ".version" stable/suse-observability/Chart.yaml)',
      'echo "Triggering beest pipeline with STACKSTATE_HELM_CHART_VERSION=${STACKSTATE_HELM_CHART_VERSION}"',
      'curl --verbose --fail-with-body --request POST --form "token=$STS_BEEST_TRIGGER_TOKEN" --form ref=main --form "variables[STACKSTATE_HELM_CHART_VERSION]=${STACKSTATE_HELM_CHART_VERSION}" --form "variables[TRIGGER_AGENT_X86_TESTS]=true" "https://gitlab.com/api/v4/projects/$BEEST_PROJECT_ID/trigger/pipeline"',
    ],
    rules: [
      {
        @'if': '$CI_COMMIT_BRANCH == "master"',
        changes: change_paths_for_published('suse-observability'),
        when: 'manual',
      },
      {
        @'if': '$CI_COMMIT_BRANCH',
        changes: change_paths_for_published('suse-observability'),
        when: 'manual',
      },
    ],
    allow_failure: true,
  },
};

// Main
{
  // Only run for merge requests, tags, the default (master) branch, or explicitly requested update pipelines
  workflow: {
    rules: [
      { @'if': '$CI_MERGE_REQUEST_IID' },
      { @'if': '$CI_COMMIT_TAG' },
      { @'if': '$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH' },
      { @'if': '$RUN_UPDATECLI' },
      { @'if': '$RUN_UPDATECLI_STACKSTATE' },
      { @'if': '$RUN_UPDATECLI_STACKPACKS' },
    ],
  },
  image: variables.images.chart_testing,
  stages: ['build', 'push-charts-to-test', 'validate', 'test', 'update', 'push-charts-to-internal', 'push-charts-to-public', 'push-charts-to-rancher'],

  variables: {
    HELM_VERSION: 'v3.1.2',
    PROMOTION_DRY_RUN: 'no',
  },
}
+ build_chart_jobs
+ build_local_chart_jobs
+ validate_chart_jobs
+ validate_chart_images_jobs
+ validate_local_chart_jobs
+ validate_updatecli_config
+ check_chart_version_jobs
+ test_chart_jobs
+ test_local_chart_jobs
+ resource_usage

+ push_charts_to_internal_jobs
+ push_prerelease_charts_to_public_jobs
+ push_charts_to_public_jobs
+ updatecli_job
+ push_suse_observability_to_rancher_registry
+ beest_triggers
