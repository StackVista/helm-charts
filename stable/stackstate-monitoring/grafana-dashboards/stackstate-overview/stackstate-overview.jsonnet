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
local row = grafana.row;

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
    expr='sum by(data_type, service)(stackstate_kafka2es_received_data_latency_seconds{%(selectors)s, quantile="0.95"}) / sum by(data_type, service)(rate(stackstate_kafka2es_received_data_latency_seconds_count{%(selectors)s}[$__rate_interval]))' % ({ selectors: variables.grafana.standard_selectors_string }),
    legendFormat='{{data_type}} - {{service}}',
  )
);

local kafka_lag = graphPanel.new(
  title='Kafka - Topic record consumption - lag',
  datasource=datasource,
  format='short',
).addTarget(
  prometheus.target(
    expr='sum(max_over_time(kafka_consumer_consumer_fetch_manager_metrics_records_lag_max{%(selectors)s, topic!=""}[10m])) by (topic, service, client_id)' % ({ selectors: variables.grafana.standard_selectors_string }),
    legendFormat='{{topic}} - {{service}} - {{client_id}}',
  )
);

local kafka_lag_time_to_catchup = graphPanel.new(
  title='Kafka - Topic record consumption - time to catch up',
  datasource=datasource,
  format='s',
).addTarget(
  prometheus.target(
    expr='sum(max_over_time(kafka_consumer_consumer_fetch_manager_metrics_records_lag_max{%(selectors)s, topic!=""}[10m])) by (topic, service, client_id) / sum(max_over_time(kafka_consumer_consumer_fetch_manager_metrics_records_consumed_rate{%(selectors)s, topic!=""}[10m])) by (topic, service, client_id)' % ({ selectors: variables.grafana.standard_selectors_string }),
    legendFormat='{{topic}} - {{service}} - {{client_id}}',
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
  uid='bc77a880dabe35cc1e7d3175095235ff',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      row.new('Elasticserach Metric / Event / Trace ingestion latency'),
      stackstate_kafka2es_received_data_latency_seconds,
      stackstate_kafka2es_time_to_catchup,
      row.new('Kafka topic record consumption'),
      kafka_lag,
      kafka_lag_time_to_catchup,
      row.new('StackState Receiver limits'),
      stackstate_receiver_unique_element_saturation,
      stackstate_receiver_created_element_saturation,
    ],
  )
)
