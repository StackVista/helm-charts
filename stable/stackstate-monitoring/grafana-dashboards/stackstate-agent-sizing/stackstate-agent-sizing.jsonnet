// Base imports
local grafana = import 'grafonnet-lib/grafonnet/grafana.libsonnet';
local functions = import 'jsonnet-libs/extras/grafana_dashboards_repo/functions.libsonnet';
local variables = import 'jsonnet-libs/extras/grafana_dashboards_repo/variables.libsonnet';

// Shortcuts
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local link = grafana.link;
local prometheus = grafana.prometheus;
local template = grafana.template;

// Local variables
local datasource = '$datasource';

local stackstate_receiver_unique_agent = graphPanel.new(
  title='Receiver - Connected agent count',
  description='Number of currently connected agents',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_passed_count{%s, element_type="agent"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{element_type}}',
  )
);

local total_persistent_volume_usage = graphPanel.new(
  title='Total persistent volume usage',
  description='Aggregate of all persistent volumes',
  format='bytes',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='(sum without(instance, node, persistentvolumeclaim) (kubelet_volume_stats_capacity_bytes{cluster="$cluster", namespace="$namespace", job="kubelet", persistentvolumeclaim=~".*"}) - sum without(instance, node, persistentvolumeclaim) (kubelet_volume_stats_available_bytes{cluster="$cluster", namespace="$namespace", job="kubelet", persistentvolumeclaim=~".*"}))',
    legendFormat='{{namespace}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Agent sizing',
  editable=true,
  refresh='60s',
  tags=['stackstate', 'receiver'],
  time_from='now-1h',
  uid='d86633e624b6df5beace9cde97223928',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_receiver_unique_agent,
      total_persistent_volume_usage,
    ],
  )
)
