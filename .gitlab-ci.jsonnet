// Imports
local variables = import '.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet';

// Shortcuts
local repositories = variables.helm.repositories;
local charts = variables.helm.charts;

local push_charts_template = {
  before_script:
  ['.gitlab/push_before_script.sh'] +
  [
    'helm repo add %s %s' % ['%s' % std.strReplace(name, '_', '-'), repositories[name]]
    for name in std.objectFields(repositories)
  ] +
  ['helm repo update'],
  script: ['sh test/sync-repo.sh'],
  stage: 'build',
};

local push_chart_job(chart, repository_url, repository_username, repository_password) = {
  script: [
    'helm plugin install https://github.com/chartmuseum/helm-push.git',
    'helm push --username ' + repository_username + ' --password ' + repository_password + ' ${CHART} ' + repository_url,
  ],
  only: { changes: [chart + '/**/*'] },
  variables: {
    CHART: 'stable/' + chart,
  },
};

local push_charts_to_internal_jobs = {
  ['push_%s_to_internal' % chart]: (push_chart_job(chart,
      '${CHARTMUSEUM_INTERNAL_URL}',
'${CHARTMUSEUM_INTERNAL_USERNAME}',
'${CHARTMUSEUM_INTERNAL_PASSWORD}') {
    stage: 'push-charts-to-internal',
    only+: { refs: ['master'] },
  })
  for chart in charts
};

local push_charts_to_public_jobs = {
  ['push_%s_to_public' % chart]: (push_chart_job(chart,
      '${CHARTMUSEUM_URL}',
'${CHARTMUSEUM_USERNAME}',
'${CHARTMUSEUM_PASSWORD}') {
    stage: 'push-charts-to-public',
    only+: { refs: ['master'] },
    needs: ['push_%s_to_internal' % chart],
    when: 'manual',
  })
  for chart in charts
};

local validate_and_push_jobs = {
  push_stable_charts: push_charts_template {
    only: { refs: ['master'] },
    when: 'manual',
  },
  push_test_charts: push_charts_template {
    except: { refs: ['master'] },
    needs: ['validate_charts'],
    only: { refs: ['branches'] },
    variables: {
      AWS_BUCKET: 's3://helm-test.stackstate.io',
      REPO_URL: 'https://helm-test.stackstate.io/',
    },
  },
  validate_charts: {
    before_script: ['.gitlab/validate_before_script.sh'],
    environment: 'stseuw1-sandbox-main-eks-sandbox/${CI_COMMIT_REF_NAME}',
    except: { refs: ['master'] },
    only: { refs: ['branches'] },
    script: [
      'ct list-changed --config test/ct.yaml',
      'ct lint --debug --config test/ct.yaml',
      '.gitlab/validate_kubeval.sh',
    ],
    stage: 'test',
  },
};

// Main
{
  image: 'quay.io/helmpack/chart-testing:v3.0.0-beta.2',
  stages: ['test', 'build', 'push-charts-to-internal', 'push-charts-to-public'],

  variables: {
    HELM_VERSION: 'v3.1.2',
    KUBEVAL_SCHEMA_LOCATION: 'https://raw.githubusercontent.com/instrumenta/kubernetes-json-schema/master',
  },
}
+ push_charts_to_internal_jobs
+ push_charts_to_public_jobs
+ validate_and_push_jobs
