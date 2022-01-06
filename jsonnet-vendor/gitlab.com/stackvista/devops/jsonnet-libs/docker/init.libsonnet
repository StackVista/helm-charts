{
  local Docker = self,

  build(image, build_args={}, cache=true, compress=true, context='.', dockerfile='Dockerfile', extra_opts=[], extra_tags=[], squash=true, target=false):: [
    local buildargs_opt = std.join(' ', ['--build-arg %s=%s' % [key, build_args[key]] for key in std.objectFields(build_args)]);
    local cache_opt = if cache == false then '--no-cache' else if std.type(cache) == 'boolean' then '--no-cache' else '--cache-from=%s' % cache;
    local compress_opt = if compress == true then '--compress';
    local dockerfile_opt = std.format('--file %s', dockerfile);
    local squash_opt = if squash == true then '--squash';
    local tags_opt = std.join(' ', ['--tag %s:%s' % [image, tag] for tag in extra_tags]);
    local target_opt = if std.type(target) == 'string' then '--target %s' % target;

    local opts = std.join(' ', [buildargs_opt, cache_opt, squash_opt, dockerfile_opt, target_opt, compress_opt, tags_opt] + extra_opts);
    'docker build %s --tag %s %s' % [opts, image, context],
  ],

  build_and_push(image, build_args={}, cache=true, compress=true, context='.', dockerfile='Dockerfile', extra_opts=[], extra_tags=[], squash=true, target=false):: (
    Docker.build(image, build_args, cache, compress, context, dockerfile, extra_opts, extra_tags, squash, target) +
    Docker.push_all(['%s:%s' % [image, tag] for tag in extra_tags])
  ),

  build_file(context='.', dockerfile='Dockerfile'):: [
    'docker build --file %s %s' % [dockerfile, context],
  ],

  cp(image, src, dest):: [
    'docker create %s | xargs -I{} docker cp {}:%s %s' % [image, src, dest],
  ],

  login_dockerhub(server='docker.io', user='${docker_user}', password='${docker_password}'):: (
    Docker.login(server, user, password)
  ),

  login_quay(server='quay.io', user='${quay_user}', password='${quay_password}'):: (
    Docker.login(server, user, password)
  ),

  login(server, user, password):: [
    'echo "%s" | docker login --username=%s --password-stdin %s' % [password, user, server],
  ],

  push(image):: [
    'docker push %s' % image,
  ],

  push_all(images=[]):: (
    ['docker push %s' % image for image in images]
  ),

  run(image, cmd, opts=[]):: [
    local optstr = std.join(' ', opts);
    'docker run %s %s %s' % [optstr, image, cmd],
  ],
}
