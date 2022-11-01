// Base imports
local grafana = import 'grafonnet-lib/grafonnet/grafana.libsonnet';
local functions = import 'jsonnet-libs/extras/grafana_dashboards_repo/functions.libsonnet';
local variables = import 'jsonnet-libs/extras/grafana_dashboards_repo/variables.libsonnet';

// Shortcuts
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local row = grafana.row;
local link = grafana.link;
local prometheus = grafana.prometheus;
local template = grafana.template;

// Local variables
local datasource = '$datasource';

// Main
local stackstate_sync_topo_total_counts = grafana.singlestat.new(
    'Sync - Topology Components & Relations',
    description='Total number of topology components and relations present at any given point in time.',
    datasource='$datasource',
    format='',
    gaugeShow=true,
    valueName='current',
    span=3,
    thresholds='100000000,100000000',
).addTarget(
        prometheus.target(
        expr='sum by (element_type, pod)(stackstate_sync_topo_total_counts{%s})' % variables.grafana.standard_selectors_string,
        legendFormat='{{element_type}}s',
    )
);

local stackstate_sync_changes_total = graphPanel.new(
  title='Sync - Changes Rate',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (type, pod)(rate(stackstate_sync_changes_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{type}}',
  )
);

local stackstate_sync_extTopo_changes_total = graphPanel.new(
  title='Sync - Changes Rate per SyncId',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (node_id, operation_type, pod)(rate(stackstate_sync_extTopo_changes_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{node_id}} - {{operation_type}}',
  )
);


local stackstate_sync_extTopo_conflated_changes_total = graphPanel.new(
  title='ExtTopo - conflated changes rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (integration_type, pod)(rate(stackstate_sync_extTopo_conflated_changes_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

local stackstate_sync_observed_conflate_changes_total = graphPanel.new(
  title='ExtTopo - observed changes to conflate rate',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (integration_type, pod)(rate(stackstate_sync_extTopo_observed_conflate_changes_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

// ExtTopo messages that bypass deduplication
local stackstate_sync_extTopo_changes_forwarded_total = graphPanel.new(
  title='ExtTopo forwarded messages',
  datasource=datasource,
).addTarget(
  prometheus.target(
      expr='sum by (integration_type, pod)(rate(stackstate_sync_extTopo_changes_forwarded_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
      legendFormat='{{pod}} - {{integration_type}}',
    )
);

// ExtTopo deduplicated messages
local stackstate_sync_extTopo_duplicated_changes_total = graphPanel.new(
  title='ExtTopo deduplicated messages',
  datasource=datasource,
).addTarget(
  prometheus.target(
      expr='sum by (integration_type, pod)(rate(stackstate_sync_extTopo_duplicated_changes_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
      legendFormat='{{pod}} - {{integration_type}}',
    )
);

local stackstate_sync_latency_row = row.new(
    title='Sync Latency'
);

// stackstate_sync_latency
local stackstate_sync_latency_seconds = graphPanel.new(
  title='Sync - Latency Seconds',
  description='Time taken from collecting the data (in agent for example) until persisting in topology',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_sync_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

// stackstate_sync_sts_latency
local stackstate_sync_sts_latency_seconds = graphPanel.new(
  title='Sync - STS (Receiver to Sync) Latency Seconds',
  description='Time taken from ingesting the data in STS (pushing to Kafka) until persisting in topology',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_sync_sts_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

local stackstate_extTopo_conflate_latency_row = row.new(
    title='ExtTopo Conflate Latency'
);

// stackstate_extTopo_conflate_latency
local stackstate_extTopo_conflate_latency_seconds = graphPanel.new(
  title='ExtTopo Conflate - Latency Seconds',
  description='Time taken from collecting the data (in agent for example) until passing through conflation',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_sync_conflate_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

// stackstate_extTopo_conflate_sts_latency
local stackstate_extTopo_conflate_sts_latency_seconds = graphPanel.new(
  title='Sync - STS (Receiver to Conflate) Latency Seconds',
  description='Time taken from ingesting the data in STS (pushing to Kafka) until passing through conflation',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_sync_conflate_sts_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

local stackstate_plugin_idExtractor_latency_row = row.new(
    title='Plugin IdExtractor Latency'
);

// stackstate_extTopo_conflate_latency
local stackstate_plugin_idExtractor_latency_seconds = graphPanel.new(
  title='Plugin IdExtractor - Latency Seconds',
  description='Time taken from collecting the data (in agent for example) until passing through the idExtractor stage',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_plugin_idExtractor_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

// stackstate_plugin_idExtractor_sts_latency
local stackstate_plugin_idExtractor_sts_latency_seconds = graphPanel.new(
  title='Sync - STS (Receiver to IdExtractor) Latency Seconds',
  description='Time taken from ingesting the data in STS (pushing to Kafka) until passing through the idExtractor stage',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_plugin_idExtractor_sts_latency_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

// Batch failures
local stackstate_sync_batch_errors_total = graphPanel.new(
  title='ExtTopo - Batch Errors Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (node_id, pod)(rate(stackstate_sync_extTopo_batch_errors_total{namespace=~"$namespace"}[1m]))',
    legendFormat='{{pod}} - {{node_id}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Sync',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'sync'],
  time_from='now-1h',
  uid='b7c79dd698255aa350c49903afc29c0d',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_sync_topo_total_counts,
      stackstate_sync_extTopo_changes_total,
      stackstate_sync_changes_total,
      stackstate_sync_extTopo_conflated_changes_total,
      stackstate_sync_observed_conflate_changes_total,
      stackstate_sync_batch_errors_total,
      stackstate_sync_extTopo_changes_forwarded_total,
      stackstate_sync_extTopo_duplicated_changes_total,
      stackstate_sync_latency_row,
      stackstate_sync_latency_seconds,
      stackstate_sync_sts_latency_seconds,
      stackstate_extTopo_conflate_latency_row,
      stackstate_extTopo_conflate_sts_latency_seconds,
      stackstate_extTopo_conflate_latency_seconds,
      stackstate_plugin_idExtractor_latency_row,
      stackstate_plugin_idExtractor_latency_seconds,
      stackstate_plugin_idExtractor_sts_latency_seconds,

    ],
  )
)
