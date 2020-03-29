// Imports
local variables = import '.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet';

// Shortcuts
local repositories = variables.helm.repositories;

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

{
  image: 'quay.io/helmpack/chart-testing:v3.0.0-beta.2',
  stages: ['test', 'build'],
  push_stable_charts: push_charts_template { only: { refs: ['master'] } },
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
  variables: {
    HELM_VERSION: 'v3.1.2',
    KUBEVAL_SCHEMA_LOCATION: 'https://raw.githubusercontent.com/instrumenta/kubernetes-json-schema/master',
  },
}
