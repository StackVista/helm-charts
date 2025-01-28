# Installation

The chart uses k8s operators which replaces the former dependent charts. Each of dependency requires some changes, bellow is the list of required modifications:

## ClickHouse

You should modify the previous chart to point to new ClickHouse installation
```yaml
stackstate:
  components:
    all:
      clickHouse:
        hostnames: clickhouse-suse-observability
        username: observability
        password: observability # The same password as configured in this helm chart for `observability` user
opentelemetry-collector:
  config:
    exporters:
      clickhousests:
        endpoint: tcp://clickhouse-suse-observability:9000?dial_timeout=10s&compress=lz4
        username: observability
        password: observability # The same password as configured in this helm chart for `observability` user
```
