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
local stackstate_sync_duration_seconds = graphPanel.new(
  title='Sync - Average processing time for a synchronization message',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='(sum by (integration_type, pod)(rate(stackstate_sync_duration_seconds_sum{%(standard_selectors_string)s}[1m])) / (sum by (integration_type, pod)(rate(stackstate_sync_changes_total{%(standard_selectors_string)s}[1m]))))' % variables.grafana,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

local stackstate_sync_changes_total = graphPanel.new(
  title='Sync - Changes Rate',
  description='Rate per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (integration_type, type, pod)(rate(stackstate_sync_changes_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{integration_type}} - {{type}}',
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

// Batch failures
local stackstate_sync_batch_errors_total = graphPanel.new(
  title='Sync - Batch Errors Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (integration_type, pod)(rate(stackstate_sync_batch_errors_total{namespace=~"$namespace"}[1m]))',
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Sync',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'sync'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_sync_duration_seconds,
      stackstate_sync_changes_total,
      stackstate_sync_extTopo_conflated_changes_total,
      stackstate_sync_observed_conflate_changes_total,
      stackstate_sync_latency_seconds,
      stackstate_sync_sts_latency_seconds,
      stackstate_sync_batch_errors_total,
      stackstate_sync_extTopo_changes_forwarded_total,
      stackstate_sync_extTopo_duplicated_changes_total,
    ],
  )
)
