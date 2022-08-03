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

// akka_http_responses_duration
local akka_http_responses_duration_seconds = graphPanel.new(
  title='HTTP - Response duration',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='akka_http_responses_duration_seconds{%s,quantile=~"0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{path}} - {{status}}',
  )
);

local akka_http_responses_total = graphPanel.new(
  title='HTTP - Counts Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, path, status)(rate(akka_http_responses_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{path}} - {{status}}',

  )
);

local akka_http_responses_errors_total = graphPanel.new(
  title='HTTP - Error Counts Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, path, status)(rate(akka_http_responses_errors_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{path}} - {{status}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - HTTP Endpoints',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'HTTP Endpoints'],
  time_from='now-1h',
  uid='ea9989a65cac81cb34d64eb6cf5d2c91',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      akka_http_responses_duration_seconds,
      akka_http_responses_total,
      akka_http_responses_errors_total,
    ],
  )
)
