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
local steps = ['dirty_elements', 'own_state', 'topological_sort'];

local quantile_targets(metric) = [
  prometheus.target(
    expr='%(metric)s' % { metric: metric, selectors: std.join(', ', variables.grafana.standard_selectors) },
    legendFormat='{{pod}} - {{quantile}}',
  ),
];

local step_targets(metric, chart_steps=steps, chart_quantiles=['0.98']) = [
  prometheus.target(
    expr='%(metric)s{%(selectors)s}' % { metric: metric, selectors: std.join(', ', variables.grafana.standard_selectors + ['quantile="%s"' % quantile, 'step="%s"' % step]) },
    legendFormat='{{pod}} - %s - %s' % [quantile, step],
  )
  for quantile in chart_quantiles
  for step in chart_steps
];

local rate_targets(metric, period='1m') = [
  prometheus.target(
    expr='rate(%(metric)s{%(selector)s}[%(period)s])' % { metric: metric, selector: std.join(', ', variables.grafana.standard_selectors), period: period },
    legendFormat='{{pod}}',
  ),
];

// Main

// stackstate_state_propagation_calculation_duration

local stackstate_state_propagation_calculation_duration_seconds = graphPanel.new(
  title='Propagation - Calculation Duration Seconds',
  datasource=datasource,
).addTargets(quantile_targets(metric='stackstate_state_propagation_calculation_duration_seconds{%(selectors)s}' % variables.grafana.standard_selectors_string));

local stackstate_state_propagation_calculation_duration_seconds_count = graphPanel.new(
  title='Propagation - Calculation batches per second',
  datasource=datasource,
).addTargets(rate_targets(metric='stackstate_state_propagation_calculation_duration_seconds_count'));

// stackstate_state_propagation_calculation_steps_duration

local stackstate_state_propagation_calculation_steps_duration_seconds = graphPanel.new(
  title='Propagation - Calculation Steps Duration Seconds',
  datasource=datasource,
).addTargets(step_targets(metric='stackstate_state_propagation_calculation_steps_duration_seconds'));

// stackstate_state_propagation_persistence_duration

local stackstate_state_propagation_persistence_duration_seconds = graphPanel.new(
  title='Propagation - Persistence Duration Seconds',
  datasource=datasource,
).addTargets(quantile_targets(metric='sum by (pod, quantile)(stackstate_state_propagation_persistence_duration_seconds{%(selectors)s})' % variables.grafana.standard_selectors_string));

// stackstate_state_propagation_enqueued / dequeued

local stackstate_state_propagation_enqueued_commands_total = graphPanel.new(
  title='Propagation - Enqueued Commands Total - Rate (per second)',
  datasource=datasource,
).addTargets(rate_targets(metric='stackstate_state_propagation_enqueued_commands_total'));

local stackstate_state_propagation_dequeued_commands_total = graphPanel.new(
  title='Propagation - Dequeued Commands Total - Rate (per second)',
  datasource=datasource,
).addTargets(rate_targets(metric='stackstate_state_propagation_dequeued_commands_total'));

local stackstate_state_propagation_queued_commands_difference = graphPanel.new(
  title='Propagation - Queue size (enqueued - dequeued)',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='%(metric_one)s{%(selectors)s} - %(metric_two)s{%(selectors)s}' % {
      metric_one: 'stackstate_state_propagation_enqueued_commands_total',
      metric_two: 'stackstate_state_propagation_dequeued_commands_total',
      selectors: variables.grafana.standard_selectors_string,
    },
    legendFormat='{{pod}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - State Propagation',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'state-propagation'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      // stackstate_state_propagation_calculation_duration
      stackstate_state_propagation_calculation_duration_seconds,
      stackstate_state_propagation_calculation_duration_seconds_count,

      // stackstate_state_propagation_calculation_steps_duration
      stackstate_state_propagation_calculation_steps_duration_seconds,

      // stackstate_state_propagation_persistence_duration
      stackstate_state_propagation_persistence_duration_seconds,

      // stackstate_state_propagation_enqueued / dequeued
      stackstate_state_propagation_dequeued_commands_total,
      stackstate_state_propagation_enqueued_commands_total,
      stackstate_state_propagation_queued_commands_difference,
    ],
  )
)
