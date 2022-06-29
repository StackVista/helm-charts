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

local gauge_target(metric, legend='{{pod}}') =
  local selector = std.join(', ', variables.grafana.standard_selectors);
  prometheus.target(
    expr='sum(%(metric)s{%(selector)s}) by (service)' % { metric: metric, selector: selector },
    legendFormat=legend,
  );

local counter_target(metric, period='1m', legend='{{pod}}', intervalFactor=5) =
  local selector = std.join(', ', variables.grafana.standard_selectors);
  prometheus.target(
    expr='sum(increase(%(metric)s{%(selector)s}[%(period)s])) by (service)' % { metric: metric, period: period, selector: selector },
    legendFormat=legend,
    intervalFactor=intervalFactor,
  );

local histogram_target(metric, quantile, legend, extra_selectors=[]) =
  local selectors = variables.grafana.standard_selectors + extra_selectors;
  local selector = std.join(', ', selectors);
  prometheus.target(
    expr='histogram_quantile(%(quantile)s, sum(rate(%(metric)s{%(selector)s}[$__rate_interval])) by (le))' % { metric: metric, quantile: quantile, selector: selector },
    legendFormat=legend,
  );

local step_target(step, quantile, legend) =
  histogram_target('spotlight_step_timings_seconds_bucket', quantile, legend, extra_selectors=['step="%(step)s"' % { step: step }]);

// Main

local anomalies_detected = graphPanel.new(
  title='Anomalies detected (per minute)',
  datasource=datasource,
  aliasColors={ LOW: '#FADE2A', MEDIUM: '#FF9830', HIGH: '#F2495C' },
).addTargets([
  counter_target(metric='low_anomalies_found_total', legend='LOW'),
  counter_target(metric='medium_anomalies_found_total', legend='MEDIUM'),
  counter_target(metric='high_anomalies_found_total', legend='HIGH'),
]);

local errors = graphPanel.new(
  title='Errors (per minute)',
  datasource=datasource,
).addTargets([
  counter_target(metric='processing_error_count_total', legend='Processing'),
  counter_target(metric='invalid_data_count_total', legend='Invalid Data'),
  counter_target(metric='dead_on_arrival_count_total', legend='Training'),
]);

local job_runs = graphPanel.new(
  title='Job runs (per minute)',
  datasource=datasource,
).addTargets([
  counter_target(metric='train_pipeline_runs_total', legend='train'),
  counter_target(metric='detect_pipeline_runs_total', legend='detect'),
]);

local aad_scale = graphPanel.new(
  title='Should the AAD scale? (number of streams)',
  datasource=datasource,
).addTargets([
  gauge_target(metric='spotlight_streams_with_anomaly_check', legend='Anomaly Health Checks'),
  gauge_target(metric='spotlight_streams_checked', legend='Checked Streams'),
]);

local histogram_panel(title, metric) = graphPanel.new(
  title=title,
  datasource=datasource,
  format='s',
).addTargets([
  histogram_target(metric, 0.9, '90%'),
  histogram_target(metric, 0.75, '75%'),
  histogram_target(metric, 0.5, '50%'),
]);

local step_panel(title, step) = graphPanel.new(
  title=title,
  datasource=datasource,
  format='s',
).addTargets([
  step_target(step, 0.9, '90%'),
  step_target(step, 0.75, '75%'),
  step_target(step, 0.5, '50%'),
]);

// Init dashboard
dashboard.new(
  'StackState - Spotlight',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'spotlight'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      anomalies_detected,
      errors,
      job_runs,
      aad_scale,
      histogram_panel('Training runs (quantiles)', 'train_timings_seconds_bucket'),
      histogram_panel('Detect runs (quantiles)', 'detect_timings_seconds_bucket'),
      step_panel('Telemetry query (quantiles)', 'source'),
      step_panel('Store annotations (quantiles)', 'save_annotations'),
      step_panel('Report topology event (quantiles)', 'log_topology_event'),
      step_panel('Topology query (quantiles)', 'topology_query'),
    ],
  )
)
