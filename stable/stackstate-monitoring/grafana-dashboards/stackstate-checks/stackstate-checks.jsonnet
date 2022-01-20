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
local stackstate_check_change_rate = graphPanel.new(
  title='Check - Changes Rate',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (check_name, pod)(rate(stackstate_check_change_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{check_name}}',
  )
);

local stackstate_check_error_rate = graphPanel.new(
  title='Check - Error Rate',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (check_name, pod)(rate(stackstate_check_error_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{check_name}}',
  )
);

local stackstate_check_run_rate = graphPanel.new(
  title='Check - Run Rate',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (check_name, pod)(rate(stackstate_check_run_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{check_name}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Checks',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'checks'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_check_change_rate,
      stackstate_check_error_rate,
      stackstate_check_run_rate,
    ],
  )
)
