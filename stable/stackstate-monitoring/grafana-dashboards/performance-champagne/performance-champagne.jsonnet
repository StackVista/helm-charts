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
local singlestatGuageHeight = 150;
local selectors = ['cluster="$cluster"', 'namespace="$namespace"'];
local selectors_string = std.join(', ', selectors);
local selectors_ds = ['cluster=""', 'job="kubelet"', 'metrics_path="/metrics"', 'namespace="$namespace"'];
local selectors_ds_string = std.join(', ', selectors_ds);

local stackstate_receiver_unique_agent = grafana.singlestat.new(
    'Receiver - Connected agent count',
    datasource='$datasource',
    format='',
    gaugeShow=true,
    valueName='current',
    height=singlestatGuageHeight,
    span=3,
    thresholds='100,1000',
).addTarget(
        prometheus.target(
        expr='stackstate_receiver_unique_element_passed_count{%s, element_type="agent"}' % selectors_string,
        legendFormat='{{element_type}}',
    )
);


local stackstate_receiver_unique_element_request_approx_count = graphPanel.new(
  title='Receiver - Current incoming element (approximate) count',
  description='Amount of data being offered to the receiver at this moment',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_request_approx_count{%s}' % selectors_string,
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_element_create_request_approx_count = graphPanel.new(
  title='Receiver - Incoming created elements in the last hour (approximate) count',
  description='Amount of created in the past hour',
  datasource=datasource,
  fill=0
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_request_approx_count{%s}' % selectors_string,
    legendFormat='{{element_type}}',
  )
).addTarget(
    prometheus.target(
        expr='stackstate_receiver_element_create_passed_max{%s}' % selectors_string,
        legendFormat='{{element_type}} budget',
      )
).addSeriesOverride(
 {
         alias: '/.* budget/',
         color: '#F2495C',
         dashes: true,
       }
);

local stackstate_receiver_unique_element_saturation = graphPanel.new(
  title='Receiver - Element budget saturation',
  description='Percentage of agent element processing budget filled up',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_unique_element_passed_count{%s} / stackstate_receiver_unique_element_passed_max{%s}' % [selectors_string, selectors_string],
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_created_element_saturation = graphPanel.new(
  title='Receiver - Created elements hourly budget saturation',
  description='Percentage of newly created elements budget filled up',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_passed_count{%s} / stackstate_receiver_element_create_passed_max{%s}' % [selectors_string, selectors_string],
    legendFormat='{{element_type}}',
  )
);

local stackstate_receiver_created_element_total = graphPanel.new(
  title='Receiver - Created elements hourly total',
  description='Total elements created',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='stackstate_receiver_element_create_passed_count{%s}' % selectors_string,
    legendFormat='{{element_type}}',
  )
);

local es_disk = graphPanel.new(
  title='ES - Used disk space',
  description='ES Used disk space',
  datasource=datasource,
  formatY1='gbytes',
).addTarget(
  prometheus.target(
    expr='sum without(instance, node) (kubelet_volume_stats_available_bytes{%s, persistentvolumeclaim="data-stackstate-performance-elasticsearch-master-0"})' % selectors_string,
    legendFormat='{{element_type}}',
  )
);

local stackstate_sync_ignored_changes_total = graphPanel.new(
  title='Ignored Changes Rate',
  datasource=datasource,
).addTarget(
  prometheus.target(
    expr='sum by (integration_type, pod)(rate(stackstate_sync_ignored_changes_total{%s}[1m]))' % selectors_string,
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
    expr='stackstate_sync_latency_seconds{%s, quantile="0.98"}' % selectors_string,
    legendFormat='{{pod}} - {{integration_type}}',
  )
);

local dn_disk = graphPanel.new(
  title='Datanode - Used disk space',
  description='Datanode Used disk space',
  datasource=datasource,
  formatY1='gbytes',
).addTarget(
  prometheus.target(
    expr='sum without(instance, node) (kubelet_volume_stats_available_bytes{%s, persistentvolumeclaim="data-stackstate-performance-hbase-hdfs-dn-0"})' % selectors_string,
    legendFormat='{{element_type}}',
  )
);

// Init dashboard
dashboard.new(
  'StackState - Amplifier Performance',
  editable=true,
  refresh='10s',
  tags=['stackstate', 'performance-champagne'],
  time_from='now-1h',
)
.addTemplates(variables.grafana.common_dashboard_variables)

.addPanels(
  functions.grafana.grid_positioning(
    [
      stackstate_receiver_unique_agent,
      stackstate_receiver_unique_element_saturation,
      stackstate_sync_ignored_changes_total,
      stackstate_sync_latency_seconds,
      stackstate_receiver_unique_element_request_approx_count,
      stackstate_receiver_element_create_request_approx_count,
      stackstate_receiver_created_element_saturation,
      stackstate_receiver_created_element_total,
      es_disk,
      dn_disk,
    ],
  )
)
