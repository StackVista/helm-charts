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
local stackstate_kafka2es_source_data_latency_seconds = graphPanel.new(
  title='Kafka2Es - Data Latency (from source to being persisted)',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='stackstate_kafka2es_data_latency_seconds{%s, quantile="0.95"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local stackstate_kafka2es_data_latency_seconds_count = graphPanel.new(
  title='Kafka2Es - Kafka Messages consumed',
  datasource=datasource,
  format='cps',
).addTarget(
  prometheus.target(
    expr='sum by(data_type, service)(rate(stackstate_kafka2es_data_latency_seconds_count{%s}[$__rate_interval]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local stackstate_kafka2es_received_data_latency_seconds = graphPanel.new(
  title='Kafka2Es - Data Latency (from received to being persisted)',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='stackstate_kafka2es_received_data_latency_seconds{%s, quantile="0.95"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local stackstate_kafka2es_time_to_catchup = graphPanel.new(
  title='Kafka2Es - Time to catch up',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='sum by(data_type, service)(stackstate_kafka2es_received_data_latency_seconds{%(selectors)s, quantile="0.95"}) / sum by(data_type, service)(rate(stackstate_kafka2es_received_data_latency_seconds_count{%(selectors)s}[$__rate_interval]) > 0.05)' % ({ selectors: variables.grafana.standard_selectors_string }),
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
  uid='4e750e563e3452be01af051af1c440fe',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_kafka2es_source_data_latency_seconds,
      stackstate_kafka2es_data_latency_seconds_count,
      stackstate_kafka2es_received_data_latency_seconds,
      stackstate_kafka2es_time_to_catchup,
      capacity('GenericEvent'),
      capacity('TopologyEvent'),
      capacity('StsEvent'),
      capacity('MultiMetric'),
      capacity('Trace'),
    ],
  )
)
