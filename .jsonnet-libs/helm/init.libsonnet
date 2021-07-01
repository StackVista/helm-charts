local common_repos = {
  'banzaicloud-stable': 'https://kubernetes-charts.banzaicloud.com',
  'kube-eagle': 'https://raw.githubusercontent.com/cloudworkz/kube-eagle-helm-chart/master',
  'vmware-tanzu': 'https://vmware-tanzu.github.io/helm-charts',
  bitnami: 'https://charts.bitnami.com',
  eks: 'https://aws.github.io/eks-charts',
  elastic: 'https://helm.elastic.co',
  gitlab: 'https://charts.gitlab.io',
  gremlin: 'https://helm.gremlin.com',
  incubator: 'https://kubernetes-charts-incubator.storage.googleapis.com',
  jetstack: 'https://charts.jetstack.io',
  jfrog: 'https://charts.jfrog.io',
  stable: 'https://kubernetes-charts.storage.googleapis.com',
  stackstate: 'https://helm.stackstate.io',
  prometheus: 'https://prometheus-community.github.io/helm-charts',
};

{
  local Helm = self,

  delete(name, debug=false, dry_run=false, extra_opts=[], purge=true):: (
    local debug_opt = if debug then '--debug';
    local dry_run_opt = if dry_run then '--dry-run';
    local purge_opt = if purge then '--purge';

    // Combine all option flags into string
    local combined_opts = std.join(' ', [debug_opt, dry_run_opt, purge_opt] + extra_opts);

    'helm delete %s %s' % [combined_opts, name]
  ),

  init(debug=false, dry_run=false, extra_opts=[]):: (
    local debug_opt = if debug then '--debug';
    local dry_run_opt = if dry_run then '--dry-run';

    // Combine all option flags into string
    local combined_opts = std.join(' ', [debug_opt, dry_run_opt] + extra_opts);

    'helm init %s' % combined_opts
  ),

  install(repo, name='', atomic=false, debug=false, dry_run=false, force=true, extra_opts=[], namespace='', recreate_pods=false, set={}, set_file={}, set_string={}, timeout=600, upgrade=true, values=[], version='', wait=false):: (
    local debug_opt = if debug then '--debug';
    local dry_run_opt = if dry_run then '--dry-run';
    local force_opt = if force then '--force';
    local install_opt = if upgrade then '--install';
    local namespace_opt = if std.length(namespace) != 0 then '--namespace %s' % namespace;
    local recreate_pods_opt = if recreate_pods then '--recreate-pods';
    local set_file_opt = std.join(' ', ["--set-file '%s'=\"%s\"" % [key, set_file[key]] for key in std.objectFields(set_file)]);
    local set_opt = std.join(' ', ["--set '%s'=\"%s\"" % [key, set[key]] for key in std.objectFields(set)]);
    local set_string_opt = std.join(' ', ["--set-string '%s'=\"%s\"" % [key, set_string[key]] for key in std.objectFields(set_string)]);
    local timeout_opt = '--timeout %s' % timeout;
    local values_opt = std.join(' ', ['--values %s' % value_file for value_file in values]);
    local version_opt = if std.length(version) != 0 then '--version %s' % version;
    local wait_opt = if wait then '--wait';

    // Determine if chart name is needed
    local chart_name = if (upgrade == true && std.length(name) == 0) then (error 'For an upgrade, the "name" keyword argument must be set.') else if (upgrade == false) then '' else name;

    // Combine all option flags into string
    local combined_opts = std.join(' ', [debug_opt, dry_run_opt, force_opt, install_opt, namespace_opt, recreate_pods_opt, set_file_opt, set_opt, set_string_opt, timeout_opt, values_opt, version_opt, wait_opt] + extra_opts);

    // This 'install' function is used for *both* a straight install, or an idempotent upgrade/install.
    // The variable below determines if the 'upgrade' or 'install' command will be called.
    // By default, an idempotent upgrade/install is preferred.
    local upgrade_install_command = if upgrade then 'upgrade' else 'install';
    std.prune('helm %s %s %s %s' % [upgrade_install_command, combined_opts, chart_name, repo])
  ),

  repo: {
    add(name, url, debug=false, username=false, password=false, extra_opts=[]):: (
      local debug_opt = if debug then '--debug';
      local password_opt = if password != false then "--password='%s'" % password;
      local username_opt = if username != false then "--username='%s'" % username;

      // Combine all option flags into string
      local combined_opts = std.join(' ', [debug_opt, password_opt, username_opt] + extra_opts);

      'helm repo add %s %s %s' % [combined_opts, name, url]
    ),

    add_common():: [
      Helm.repo.add(name=repo, url=common_repos[repo], debug=true)
      for repo in std.objectFields(common_repos)
    ],
  },
}
