local master = import 'init.libsonnet';
local docker_tags = ['some-tag', 'some-other-tag'];

{
  some_key: master.docker.build_and_push(image='quay.io/stackstate/stackgraph', extra_tags=docker_tags),
  semver_release: master.gitlab.semantic_release_job(
    debug=true,
    extra_opts=['--dry-run'],
    extra_vars={ SOME_VAR: 'some_value' },
    gitlab_token='asdja7sg8eg',
    only={ refs: ['another_branch'] },
    stage='another_stage',
  ),
  helm_add_common_repos: master.helm.repo.add_common(),
  helm_add_custom_repo: master.helm.repo.add(name='custom-repo', url='https://custom-repo.helm.io', username='user', password='!QAZxsw2@'),
  helm_delete: master.helm.delete(name='some-other-chart'),
  helm_init: master.helm.init(extra_opts=['--client-only']),
  helm_install: master.helm.install(repo='stackstate/hbase', debug=true, version='1.2.3', upgrade=false, namespace='test', extra_opts=['--verify']),
  helm_upgrade: master.helm.install(repo='stackstate/stackstate', name='some-test-chart', recreate_pods=true, wait=true),
}
