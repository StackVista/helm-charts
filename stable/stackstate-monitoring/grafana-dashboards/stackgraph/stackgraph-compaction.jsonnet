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

local stackgraph_slicing_duration_seconds_98 = graphPanel.new(
  title='StackGraph - Slicing durations Seconds 98 percentile',
  description='Time spent on slicing certain table 98 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_slicing_duration_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ table_type }}',
  )
);

local stackgraph_slicing_sliced_rows_count_total = graphPanel.new(
  title='Sync - Sliced rows Rate',
  description='Rate at which we slice rows per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table_type, pod)(rate(stackgraph_slicing_sliced_rows_count_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table_type}}',
  )
);

local stackgraph_slicing_errors_count_total = graphPanel.new(
  title='Sync - Slicing errors Rate',
  description='Rate of errors while slicing per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table_type, pod)(rate(stackgraph_slicing_errors_count_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table_type}}',
  )
);

local stackgraph_slicing_compaction_duration_seconds_98 = graphPanel.new(
  title='StackGraph - Slicing compaction durations Seconds 98 percentile',
  description='Time spent on compaction after slicing certain table 98 percentile',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackgraph_slicing_compaction_duration_seconds{%s, quantile="0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{ table_type }}',
  )
);

local stackgraph_slicing_compaction_errors_count_total = graphPanel.new(
  title='Sync - Compaction errors Rate',
  description='Rate of errors while compaction per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table_type, pod)(rate(stackgraph_slicing_compaction_errors_count_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table_type}}',
  )
);

local hbase_compactionscompletedcount = graphPanel.new(
  title='Sync - Compaction completed Rate',
  description='Rate of errors while compaction per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table, pod)(rate(hbase_compactionscompletedcount{%s, name="RegionServer", sub="Regions"}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table}}',
  )
);

local hbase_majorcompactedcellscount = graphPanel.new(
  title='Sync - Major compacted cells Rate',
  description='Rate of cells compacted by major compaction per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table, pod)(rate(hbase_majorcompactedcellscount{%s, name="RegionServer", sub="Server",}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table}}',
  )
);

local hbase_compactedcellscount = graphPanel.new(
  title='Sync - Major compacted cells Rate',
  description='Rate of cells compacted by minor compaction per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table, pod)(rate(hbase_compactedcellscount{%s, name="RegionServer", sub="Server",}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table}}',
  )
);

local hbase_compactionqueuelength = graphPanel.new(
  title='Sync - Compaction queue length',
  description='Compaction queue length',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='hbase_compactionqueuelength{%s, name="RegionServer", sub="Server",}' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}}',
  )
);

local hbase_numfilescompactedcount = graphPanel.new(
  title='Sync - Num files compacted Rate',
  description='Rate of num files compacted per second of the 1 minute average',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (table, pod)(rate(hbase_numfilescompactedcount{%s, name="RegionServer", sub="Regions",}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{pod}} - {{table}}',
  )
);

// Init dashboard
dashboard.new(
  'StackGraph - Compaction',
  editable=true,
  refresh='10s',
  tags=['stackgraph', 'compaction', 'slicing', 'major compaction'],
  time_from='now-1h',
  uid='c4787962e219da7026ba634e4b2d8f5f'
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
        stackgraph_slicing_duration_seconds_98,
        stackgraph_slicing_sliced_rows_count_total,
        stackgraph_slicing_errors_count_total,
        stackgraph_slicing_compaction_duration_seconds_98,
        stackgraph_slicing_compaction_errors_count_total,
        hbase_compactionscompletedcount,
        hbase_majorcompactedcellscount,
        hbase_compactedcellscount,
        hbase_compactionqueuelength,
        hbase_numfilescompactedcount,
    ],
  )
)
