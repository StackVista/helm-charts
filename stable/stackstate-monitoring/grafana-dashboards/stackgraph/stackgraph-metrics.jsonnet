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

local stackgraph_readcache_retention_time = graphPanel.new(
  title='StackGraph - Retention time in minutes',
  description='Numbers of seconds that last entry evicted due to size limits stay in the cache',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_readcache_retention_time{%s} / 60' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}}',
  )
);

// stackgraph index lookups per transaction 98
local stackgraph_index_lookup_duration_seconds_98 = graphPanel.new(
  title='StackGraph - Transaction Index Lookups Seconds 98 percentile',
  description='Time spent on index lookups on a transaction 98 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_index_lookup_duration_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ name }}',
  )
);

// stackgraph index lookups per transaction 95
local stackgraph_index_lookup_duration_seconds_95 = graphPanel.new(
  title='StackGraph -  Transaction Index Lookups Seconds 95 percentile',
  description='Time spent on index lookups on a transaction 95 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_index_lookup_duration_seconds{%s, quantile="0.95"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ name }}',
  )
);

// stackgraph index lookups per transaction 75
local stackgraph_index_lookup_duration_seconds_75 = graphPanel.new(
  title='StackGraph - Transaction Index Lookups Seconds 75 percentile',
  description='Time spent on index lookups on a transaction 75 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_index_lookup_duration_seconds{%s, quantile="0.75"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ name }}',
  )
);

// stackgraph index get -> scan rewrite duration 98
local stackgraph_scan_index_lookup_duration_ms_98 = graphPanel.new(
  title='StackGraph - Scan Lookups Seconds 98 percentile',
  description='Time spent on index lookups on a fallback from GET to SCAN 98 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_scan_index_lookup_duration_ms{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ rowId }}',
  )
);

// stackgraph index get -> scan rewrite duration 95
local stackgraph_scan_index_lookup_duration_ms_95 = graphPanel.new(
  title='StackGraph - Scan Lookups Seconds 95 percentile',
  description='Time spent on index lookups on a fallback from GET to SCAN 95 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_scan_index_lookup_duration_ms{%s, quantile="0.95"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ rowId }}',
  )
);

// stackgraph index get -> scan rewrite duration 75
local stackgraph_scan_index_lookup_duration_ms_75 = graphPanel.new(
  title='StackGraph - Scan Lookups Seconds 75 percentile',
  description='Time spent on index lookups on a fallback from GET to SCAN 75 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_scan_index_lookup_duration_ms{%s, quantile="0.75"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ rowId }}',
  )
);

// Execution time per query
local stackstate_topology_query_execute_duration_seconds_98 = graphPanel.new(
  title='Topology Query - Query Executiom time 98 percentile',
  description='Execution time per query 98 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_topology_query_execute_duration_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ query }}',
  )
);

// Execution time per query
local stackstate_topology_query_execute_duration_seconds_95 = graphPanel.new(
  title='Topology Query - Query Executiom time 95 percentile',
  description='Execution time per query 95 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_topology_query_execute_duration_seconds{%s, quantile="0.95"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ query }}',
  )
);

// Execution time per query
local stackstate_topology_query_execute_duration_seconds_75 = graphPanel.new(
  title='Topology Query - Query Executiom time 75 percentile',
  description='Execution time per query 75 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_topology_query_execute_duration_seconds{%s, quantile="0.75"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ query }}',
  )
);

// Timeouts per query
local stackstate_topology_query_timeout_total = graphPanel.new(
  title='Topology Query - Timeouts per query',
  description='Execution time per query 98 percentile',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='rate(stackstate_topology_query_timeout_total{%s}[1m])' % variables.grafana.standard_selectors_string,
    legendFormat='{{ query }}',
  )
);

// Init dashboard
dashboard.new(
  'StackGraph - ReadCache and Transaction Metrics',
  editable=true,
  refresh='10s',
  tags=['stackgraph', 'readcache', 'transactions'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackgraph_readcache_retention_time,
      stackgraph_index_lookup_duration_seconds_98,
      stackgraph_index_lookup_duration_seconds_95,
      stackgraph_index_lookup_duration_seconds_75,
      stackgraph_scan_index_lookup_duration_ms_98,
      stackgraph_scan_index_lookup_duration_ms_95,
      stackgraph_scan_index_lookup_duration_ms_75,
      stackstate_topology_query_execute_duration_seconds_98,
      stackstate_topology_query_execute_duration_seconds_95,
      stackstate_topology_query_execute_duration_seconds_75,
      stackstate_topology_query_timeout_total,
    ],
  )
)
