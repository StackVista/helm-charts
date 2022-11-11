// Base imports
local grafana = import 'grafonnet-lib/grafonnet/grafana.libsonnet';

// Shortcuts
local template = grafana.template;

{
  local Variable = self,

  grafana: {
    common_dashboard_variables: [
      template.datasource(
        name='datasource',
        query='prometheus',
        current='Prometheus',
      ),
      template.new(
        name='cluster',
        datasource='$datasource',
        query='label_values(kube_pod_info, cluster)',
        label='cluster',
        refresh='time',
        sort=1,
      ),
      template.new(
        name='namespace',
        datasource='$datasource',
        query='label_values(kube_pod_info{cluster="$cluster"}, namespace)',
        label='namespace',
        refresh='time',
        sort=1,
      ),
      template.new(
        name='service',
        datasource='$datasource',
        query='label_values(kube_service_info{cluster="$cluster", namespace=~"$namespace"}, service)',
        label='service',
        refresh='time',
        includeAll=true,
      ),
      template.new(
        name='pod',
        datasource='$datasource',
        query='label_values(kube_pod_info{cluster="$cluster", namespace=~"$namespace"}, pod)',
        label='pod',
        refresh='time',
        sort=1,
        includeAll=true,
      ),
    ],
    standard_selectors: ['cluster="$cluster"', 'namespace="$namespace"', 'pod=~"$pod"', 'service=~"$service"'],
    standard_selectors_string: std.join(', ', Variable.grafana.standard_selectors),
    namespace_dashboard_variables: [
      template.datasource(
        name='datasource',
        query='prometheus',
        current='Prometheus',
      ),
      template.new(
        name='cluster',
        datasource='$datasource',
        query='label_values(kube_pod_info, cluster)',
        label='cluster',
        refresh='time',
        sort=1,
      ),
      template.new(
        name='namespace',
        datasource='$datasource',
        query='label_values(kube_pod_info{cluster="$cluster"}, namespace)',
        label='namespace',
        refresh='time',
        sort=1,
      ),
    ],
    namespace_selectors: ['cluster="$cluster"', 'namespace="$namespace"'],
    namespace_selectors_string: std.join(', ', Variable.grafana.namespace_selectors),
  },
}
