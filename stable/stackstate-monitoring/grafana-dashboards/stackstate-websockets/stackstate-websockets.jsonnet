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

// stackstate_websockets_stream_errors
local stackstate_websockets_stream_errors_totals = graphPanel.new(
  title='Streams - Errors Total',
  description='Errors are reported over a still active stream to the client, while the stream remains active',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, stream_type)(stackstate_websockets_stream_errors_totals{%s})' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

local stackstate_websockets_stream_errors_totals_rate = graphPanel.new(
  title='Streams - Errors Total Rate',
  description='Errors are reported over a still active stream to the client, while the stream remains active',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, stream_type)(rate(stackstate_websockets_stream_errors_totals{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

// stackstate_websockets_stream_failures
local stackstate_websockets_stream_failures_total = graphPanel.new(
  title='Streams - Failures Total',
  description='Failures stop a stream completely, the client will need to restart the stream',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, stream_type)(stackstate_websockets_stream_failures_total{%s})' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

local stackstate_websockets_stream_failures_total_rate = graphPanel.new(
  title='Streams - Failures Total Rate',
  description='Failures stop a stream completely, the client will need to restart the stream',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, stream_type)(rate(stackstate_websockets_stream_failures_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

// stackstate_websockets_stream_time_to_first_message_duration
local stackstate_websockets_stream_time_to_first_message_duration_seconds = graphPanel.new(
  title='Streams - Time to first message Duration Seconds (98th percentile)',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_websockets_stream_time_to_first_message_duration_seconds{%s,quantile=~"0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

local stackstate_websockets_stream_time_to_first_message_duration_seconds_count = graphPanel.new(
  title='Streams - Rate of stream starts per 1m',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, stream_type)(rate(stackstate_websockets_stream_time_to_first_message_duration_seconds_count{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

// stackstate_websockets_view_stream_start_duration
local stackstate_websockets_view_stream_start_duration_seconds = graphPanel.new(
  title='Streams - View Stream Start Duration Seconds (98th percentile)',
  description='The view stream sends 2 initial messages, this shows the timing for both message separately',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_websockets_view_stream_start_duration_seconds{%s, quantile=~"0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{stream_type}}',
  )
);

// stackstate_websockets_telemetry_time_to_first_message_duration
local stackstate_telemetry_stream_time_to_first_data_seconds = graphPanel.new(
  title='Telemetry - Time to first data Duration Seconds (98th percentile)',
  format='s',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_telemetry_stream_time_to_first_data_seconds{%s,quantile=~"0.98"}' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{datasource}}',
  )
);

local stackstate_telemetry_stream_time_to_first_data_seconds_count = graphPanel.new(
  title='Telemetry - Rate of telemetry stream starts per 1m',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, datasource)(rate(stackstate_telemetry_stream_time_to_first_data_seconds_count{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{datasource}}',
  )
);

local stackstate_telemetry_stream_errors_total = graphPanel.new(
  title='Telemetry - Errors Total',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, datasource)(stackstate_telemetry_stream_errors_total{%s})' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{datasource}}',
  )
);

local stackstate_telemetry_stream_errors_total_rate = graphPanel.new(
  title='Telemetry - Errors Total - 1m Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (service, datasource)(rate(stackstate_telemetry_stream_errors_total{%s}[1m]))' % variables.grafana.standard_selectors_string,
    legendFormat='{{service}} - {{datasource}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Websockets',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'websockets'],
  time_from='now-1h',
  uid='14f74ae1dc50cffc105f95502a851388',
)
.addTemplates(variables.grafana.common_dashboard_variables)
.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_websockets_stream_errors_totals,
      stackstate_websockets_stream_errors_totals_rate,
      stackstate_websockets_stream_failures_total,
      stackstate_websockets_stream_failures_total_rate,
      stackstate_websockets_stream_time_to_first_message_duration_seconds,
      stackstate_websockets_stream_time_to_first_message_duration_seconds_count,
      stackstate_websockets_view_stream_start_duration_seconds,
      stackstate_telemetry_stream_time_to_first_data_seconds,
      stackstate_telemetry_stream_time_to_first_data_seconds_count,
      stackstate_telemetry_stream_errors_total,
      stackstate_telemetry_stream_errors_total_rate,
    ],
  )
)
