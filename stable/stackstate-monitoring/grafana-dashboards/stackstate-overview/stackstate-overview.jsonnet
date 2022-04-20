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
  title='Kafka2Es - Data Latency',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='stackstate_kafka2es_data_latency_seconds{%s, quantile="0.95"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local stackstate_kafka2es_data_latency_seconds_count = graphPanel.new(
  title='Kafka2Es - Rate of consumed messages',
  datasource=datasource,
  format='cps',
).addTarget(
  prometheus.target(
    expr='sum by(data_type, service)(rate(stackstate_kafka2es_data_latency_seconds_count{%s}[$__rate_interval]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{data_type}} - {{service}}',
  )
);

local stackstate_kafka2es_time_to_catchup = graphPanel.new(
  title='Kafka2Es - Time to catch up',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='stackstate_kafka2es_data_latency_seconds{%(selectors)s, quantile="0.95"} / sum by(data_type, service)(rate(stackstate_kafka2es_data_latency_seconds_count{%(selectors)s}[$__rate_interval]))' % ({ selectors: variables.grafana.standard_selectors_string }),
    legendFormat='{{data_type}} - {{service}}',
  )
);

local kafka_lag = graphPanel.new(
  title='Kafka - Topic partition lag',
  datasource=datasource,
  format='short',
).addTarget(
  prometheus.target(
    expr='kafka_consumer_consumer_fetch_manager_metrics_records_lag_max{namespace="$namespace"}' % ({ selectors: variables.grafana.standard_selectors_string }),
    legendFormat='{{pod}} - {{topic}}',
  )
);

local kafka_lag_time_to_catchup = graphPanel.new(
  title='Kafka - Topics time to catch up',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='kafka_consumer_consumer_fetch_manager_metrics_records_lag_max{namespace="$namespace"} / kafka_consumer_consumer_fetch_manager_metrics_records_consumed_rate{namespace="$namespace"}' % ({ selectors: variables.grafana.standard_selectors_string }),
    legendFormat='{{pod}} - {{topic}}',
  )
);

local stackstate_receiver_unique_element_saturation = graphPanel.new(
  title='Receiver - Unique elements seen budget saturation',
  description='Percentage of agent element processing budget used, fully saturated when at 1 (i.e. dropping data)',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_passed_count{%(standard_selectors_string)s} / stackstate_receiver_unique_element_passed_max{%(standard_selectors_string)s}' % variables.grafana,
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_created_element_saturation = graphPanel.new(
  title='Receiver - Created elements hourly budget saturation',
  description='Percentage of newly created elements budget used, fully saturated when at 1 (i.e. dropping data)',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_passed_count{%(standard_selectors_string)s} / stackstate_receiver_element_create_passed_max{%(standard_selectors_string)s}' % variables.grafana,
    legendFormat='{{element_type}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Overview',
  editable=true,
  refresh='1m',
  tags=['stackstate', 'overview'],
  time_from='now-1h',
  description='Look here for a high level overview of StackState load and performance',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_kafka2es_data_latency_seconds,
      stackstate_kafka2es_data_latency_seconds_count,
      stackstate_kafka2es_time_to_catchup,
      kafka_lag,
      kafka_lag_time_to_catchup,
      stackstate_receiver_unique_element_saturation,
      stackstate_receiver_created_element_saturation,
    ],
  )
)
