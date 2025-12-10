{{- define "stackstate.router.configmap.data" -}}
data:
  listeners.yaml: |-
    resources:
    - "@type": type.googleapis.com/envoy.config.listener.v3.Listener
      name: listener_0
      address:
        socket_address:
          protocol: TCP
          address: 0.0.0.0
          port_value: 8080
      filter_chains:
      - filters:
        - name: envoy.filters.network.http_connection_manager
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
            {{- if or .Values.stackstate.components.router.errorlog.enabled .Values.stackstate.components.router.accesslog.enabled }}
            access_log:
            - name: envoy.access_loggers.file
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                path: /dev/stdout
            {{- if not .Values.stackstate.components.router.accesslog.enabled }}
              filter:
                status_code_filter:
                  comparison:
                    op: GE
                    value:
                      default_value: 400
                      runtime_key: log_ge_response_status_code
            {{- end }}
            {{- end }}
            codec_type: AUTO
            stat_prefix: ingress_http
            upgrade_configs:
            - upgrade_type: websocket
            route_config:
              name: local_route
              virtual_hosts:
              - name: backend
                domains:
                - "*"
                routes:
                - match:
                    prefix: "/api"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                - match:
                    prefix: "/loginCallback"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                - match:
                    prefix: "/loginInfo"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                - match:
                    prefix: "/logout"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                - match:
                    prefix: "/prometheus/api/v1/"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                    prefix_rewrite: "/api/metrics/"
                {{- if eq (include "stackstate.receiver.split.enabled" .) "true" }}
                - match:
                    prefix: "/stsAgent/api/v1/connections"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-process-agent"
                - match:
                    prefix: "/receiver/stsAgent/api/v1/connections"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-process-agent"
                    prefix_rewrite: "/stsAgent/api/v1/connections"
                - match:
                    prefix: "/stsAgent/api/v1/collector"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-process-agent"
                - match:
                    prefix: "/receiver/stsAgent/api/v1/collector"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-process-agent"
                    prefix_rewrite: "/stsAgent/api/v1/collector"
                - match:
                    prefix: "/stsAgent/logs/k8s"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-logs"
                - match:
                    prefix: "/receiver/stsAgent/logs/k8s"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-logs"
                    prefix_rewrite: "/stsAgent/logs/k8s"
                - match:
                    prefix: "/stsAgent/logs/openshift/loki/api/v1/push"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-logs"
                - match:
                    prefix: "/receiver/stsAgent/logs/openshift/loki/api/v1/push"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-logs"
                    prefix_rewrite: "/stsAgent/logs/openshift/loki/api/v1/push"
                - match:
                    prefix: "/stsAgent"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-base"
                - match:
                    prefix: "/receiver/"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver-base"
                    prefix_rewrite: "/"
                {{- else }}
                - match:
                    prefix: "/stsAgent"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver"
                - match:
                    prefix: "/receiver/"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-receiver"
                    prefix_rewrite: "/"
                {{- end }}
                - match:
                    prefix: "/"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-ui"
                  response_headers_to_add:
                    - header:
                        key: "X-Frame-Options"
                        value: "DENY"
                      append: true
            http_filters:
            - name: envoy.filters.http.router
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters.yaml: |-
    resources:
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      http_protocol_options: {}
      connect_timeout: 2s
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
        endpoints:
        {{- if eq .RouterState "active" }}
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-headless"
                  port_value: 7070
        {{- end }}
      {{- if eq (include "stackstate.receiver.split.enabled" .) "true" }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver-logs"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver-logs"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ template "common.fullname.short" . }}-receiver-logs"
                  port_value: 7077
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver-process-agent"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver-process-agent"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ template "common.fullname.short" . }}-receiver-process-agent"
                  port_value: 7077
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver-base"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver-base"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ template "common.fullname.short" . }}-receiver-base"
                  port_value: 7077
      {{- else }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ template "common.fullname.short" . }}-receiver"
                  port_value: 7077
      {{- end }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-ui"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      http_protocol_options: {}
      connect_timeout: 2s
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-ui"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ template "common.fullname.short" . }}-ui"
                  port_value: 8080
{{- end -}}
