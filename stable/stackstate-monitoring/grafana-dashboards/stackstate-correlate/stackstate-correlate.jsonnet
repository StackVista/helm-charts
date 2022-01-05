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

local stackstate_correlate_unique_active_element_approx_count = graphPanel.new(
  title='Correlate - Approximate count of elements produced',
  description='Amount of elements produced in the correlator',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_correlate_unique_active_element_approx_count{%s}' % variables.grafana.standard_selectors_string,
    legendFormat='{{element_type}}',
  )
);

local stackstate_correlate_element_create_approx_count = graphPanel.new(
  title='Correlate - Approximate count of created elements in the last hour',
  description='Approximate count of new elements create in the past hour',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_correlate_element_create_approx_count{%s}' % variables.grafana.standard_selectors_string,
    legendFormat='{{element_type}}',
  )
);

// Correlation buffer occupancy
local stackstate_correlate_buffer_occupancy = graphPanel.new(
  title='Correlator buffers occupancy',
  datasource=datasource,
).addTarget(
  prometheus.target(
      expr='stackstate_correlate_buffer_occupancy{%s}' % variables.grafana.standard_selectors_string,
      legendFormat='{{correlation_type}}',
    )
);

// Init dashboard
dashboard.new(
  'StackState - Correlation workers',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'correlate'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_correlate_unique_active_element_approx_count,
      stackstate_correlate_element_create_approx_count,
      stackstate_correlate_buffer_occupancy,
    ],
  )
)
