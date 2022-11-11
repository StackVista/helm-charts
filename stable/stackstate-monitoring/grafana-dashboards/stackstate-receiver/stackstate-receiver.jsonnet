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

local stackstate_receiver_unique_element_request_approx_count = graphPanel.new(
  title='Receiver - Current incoming element (approximate) count',
  description='Amount of data being offered to the receiver at this moment',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_request_approx_count{%s}' % variables.grafana.namespace_selectors_string,
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_element_create_request_approx_count = graphPanel.new(
  title='Receiver - Incoming created elements in the last hour (approximate) count',
  description='Amount of created in the past hour',
  datasource=datasource,
  fill=0
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_request_approx_count{%s}' % variables.grafana.namespace_selectors_string,
    legendFormat='{{element_type}}',
  )
).addTarget(
    prometheus.target(
        expr='stackstate_receiver_element_create_passed_max{%s}' % variables.grafana.namespace_selectors_string,
        legendFormat='{{element_type}} budget',
      )
).addSeriesOverride(
 {
         alias: '/.* budget/',
         color: '#F2495C',
         dashes: true,
       }
);

local stackstate_receiver_unique_element_saturation = graphPanel.new(
  title='Receiver - Element budget saturation',
  description='Percentage of agent element processing budget filled up',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_passed_count{%(namespace_selectors_string)s} / stackstate_receiver_unique_element_passed_max{%(namespace_selectors_string)s}' % variables.grafana,
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_created_element_saturation = graphPanel.new(
  title='Receiver - Created elements hourly budget saturation',
  description='Percentage of newly created elements budget filled up',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_passed_count{%(namespace_selectors_string)s} / stackstate_receiver_element_create_passed_max{%(namespace_selectors_string)s}' % variables.grafana,
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_unique_agent = graphPanel.new(
  title='Receiver - Connected agent count',
  description='Number of currently connected agents',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_passed_count{%s, element_type="agent"}' % variables.grafana.namespace_selectors_string,
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_created_element_total = graphPanel.new(
  title='Receiver - Created elements hourly total',
  description='Total elements created',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_passed_count{%s}' % variables.grafana.namespace_selectors_string,
    legendFormat='{{element_type}}',
  )
);

// akka_http_responses_duration
local akka_http_responses_duration_seconds = graphPanel.new(
  title='HTTP - Response duration',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='akka_http_responses_duration_seconds{%s,quantile=~"0.98", app_component="receiver"}' % variables.grafana.namespace_selectors_string,
    legendFormat='{{product}} - {{path}} - {{status}}',
  )
);

local akka_http_responses_total = graphPanel.new(
  title='HTTP - Counts Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (product, path, status)(rate(akka_http_responses_total{%s, app_component="receiver"}[1m]))' % variables.grafana.namespace_selectors_string,
    legendFormat='{{product}} - {{path}} - {{status}}',

  )
);

local akka_http_responses_errors_total = graphPanel.new(
  title='HTTP - Error Counts Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (product, path, status)(rate(akka_http_responses_errors_total{%s, app_component="receiver"}[1m]))' % variables.grafana.namespace_selectors_string,
    legendFormat='{{product}} - {{path}} - {{status}}',
  )
);

local akka_http_requests_active = graphPanel.new(
  title='HTTP - Active Requests',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum(akka_http_requests_active{%s, app_component="receiver"})' % variables.grafana.namespace_selectors_string,
    legendFormat='Active HTTP requests',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Receiver',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'receiver'],
  time_from='now-1h',
  uid='49e75d350918156196744427031022e8',
)
.addTemplates(variables.grafana.namespace_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_receiver_unique_element_request_approx_count,
      stackstate_receiver_element_create_request_approx_count,
      stackstate_receiver_unique_element_saturation,
      stackstate_receiver_created_element_saturation,
      stackstate_receiver_unique_agent,
      stackstate_receiver_created_element_total,
      akka_http_responses_duration_seconds,
      akka_http_responses_total,
      akka_http_responses_errors_total,
      akka_http_requests_active,
    ],
  )
)
