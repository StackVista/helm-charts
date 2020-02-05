// Imports
local variables = import '../.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet';

// Shortcut
local repositories = variables.helm.repositories;

{
  'chart-dirs': ['stable'],
  'chart-repos': [
    std.join('=', ['%s' % std.strReplace(name, '_', '-'), repositories[name]])
    for name in std.objectFields(repositories)
  ],
  'excluded-charts': ['gitlab-runner'],
  'helm-extra-args': std.join(' ', [
    '--debug',
    '--kubeconfig ${KUBECONFIG}',
    '--timeout 1200',
  ]),
  'target-branch': 'master',
  remote: 'helm',
}
