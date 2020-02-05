{
  local Gitlab = self,

  semantic_release_job(
    debug=false,
    extra_opts=[],
    extra_vars={},
    gitlab_token='${gitlab_api_scope_token}',
    image='docker.io/stackstate/stackstate-ci-images:stackstate-semver-release',
    only={ refs: ['master'] },
    stage='tag',
  )::
    local debug_opt = if debug == true then '--debug';
    local opts = std.join(' ', [debug_opt] + extra_opts);
    {
      image: image,
      only: only,
      script: [
        'semantic-release %s' % [opts],
      ],
      stage: stage,
      variables: {
        GITLAB_TOKEN: gitlab_token,
        GIT_AUTHOR_EMAIL: 'sts-admin@stackstate.com',
        GIT_AUTHOR_NAME: 'stackstate-system-user',
        GIT_COMMITTER_EMAIL: 'sts-admin@stackstate.com',
        GIT_COMMITTER_NAME: 'stackstate-system-user',
      } + extra_vars,
    },
}
