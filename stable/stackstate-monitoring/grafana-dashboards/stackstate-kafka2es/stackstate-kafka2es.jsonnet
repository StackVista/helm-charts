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
local stackstate_kafka2es_data_latency_seconds = graphPanel.new(
  title='Kafka2Es - Data Latency Seconds',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='stackstate_kafka2es_data_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local stackstate_kafka2es_data_latency_seconds_count = graphPanel.new(
  title='Kafka2Es - Kafka Messages consumed (rate msg/sec) per 1m',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by(data_type, service)(rate(stackstate_kafka2es_data_latency_seconds_count{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local capacity(data_type) =
graphPanel.new(
  title='Kafka2Es - %s' % data_type,
  datasource=datasource,
  format='bytes',
).addTarget(
  prometheus.target(
    expr='max by (index_name)(stackstate_kafka2es_index_capacity_quota_bytes{data_type="%s", %s})' % [data_type, variables.grafana.standard_selectors_string],
    legendFormat='{{index_name}} - Quotum',
  )
).addTarget(
  prometheus.target(
    expr='max by (index_name)(stackstate_kafka2es_index_capacity_usage_bytes{data_type="%s", %s})' % [data_type, variables.grafana.standard_selectors_string],
    legendFormat='{{index_name}} - Used',
  )
);


// Init dashboard
dashboard.new(
  'StackState - Kafka2Es',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'kafka2es'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_kafka2es_data_latency_seconds,
      stackstate_kafka2es_data_latency_seconds_count,
      capacity('GenericEvent'),
      capacity('TopologyEvent'),
      capacity('StsEvent'),
      capacity('MultiMetric'),
      capacity('Trace'),
    ],
  )
)
