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

// Main
local stackstate_running_query_count = graphPanel.new(
  title='Topology Queries - Running Count',
  description='Amount of queries actively running',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_websocket_view_stream_viewers_total{%s}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{request_query}}',
  )
);

local stackstate_queued_topology_queries = graphPanel.new(
  title='Topology Queries - Queued Executions',
  description='Amount of topology queries queued up for execution',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_topology_service_queued_count{%s}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{priority}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Topology Queries',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'topology'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_running_query_count,
      stackstate_queued_topology_queries,
    ],
  )
)
