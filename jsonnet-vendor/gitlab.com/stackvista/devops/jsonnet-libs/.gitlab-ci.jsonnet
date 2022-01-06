local gitlab = import 'gitlab/init.libsonnet';

{
  image: {
    entrypoint: [''],
    name: 'docker.io/sparkprime/jsonnet:latest',
  },
  jsonnet_lint: {
    stage: 'verify',
    script: '.gitlab/jsonnet_lint.sh',
  },
  semver_release:
    gitlab.semantic_release_job(debug=true) +
    { needs: ['jsonnet_lint'] },
  stages: [
    'verify',
    'tag',
  ],
}
