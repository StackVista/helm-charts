{{- define "stackstate.router.configmap.data" -}}
data:
  listeners.yaml: |-
    resources:
    - "@type": type.googleapis.com/envoy.config.listener.v3.Listener
      name: listener_0
      {{- with include "stackstate.router.listener.buffers" . }}{{ . | nindent 6 }}{{ end }}
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
            - name: envoy.access_loggers.stdout
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
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
                    prefix: "/ssoLogin"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                - match:
                    prefix: "/prometheus/api/v1/"
                  route:
                    timeout: 0s
                    cluster: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
                    prefix_rewrite: "/api/metrics/"
                - match:
                    prefix: "/stsAgent/otel/"
                  route:
                    timeout: 0s
                    cluster: "{{ include "stackstate.otelCollector.fullname" . }}"
                    prefix_rewrite: "/"
                - match:
                    prefix: "/receiver/stsAgent/otel/"
                  route:
                    timeout: 0s
                    cluster: "{{ include "stackstate.otelCollector.fullname" . }}"
                    prefix_rewrite: "/"
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
                {{- if eq (include "stackstate.mcp.enabled" .) "true" }}
                - match:
                    prefix: "/mcp"
                  route:
                    timeout: 0s
                    cluster: "{{ template "stackstate.mcp.fullname" . }}"
                {{- end }}
                - match:
                    prefix: "/"
                  route:
                    timeout: 0s
                    cluster: "{{ template "stackstate.ui.fullname" . }}"
                  response_headers_to_add:
                    - header:
                        key: "X-Frame-Options"
                        value: "DENY"
                      append_action: APPEND_IF_EXISTS_OR_ADD
            http_filters:
            {{- if .Values.stackstate.components.router.secureCookies.enabled }}
            - name: envoy.filters.http.lua
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
                default_source_code:
                  inline_string: |
                    -- Append the Secure attribute to Set-Cookie response headers when the
                    -- request reached the edge over HTTPS (X-Forwarded-Proto: https), so
                    -- session cookies are never exposed over plaintext. Plain-HTTP setups
                    -- (no/!https X-Forwarded-Proto) are left untouched. The inbound header
                    -- is set by the TLS-terminating proxy in front of the router.
                    function envoy_on_request(handle)
                      local proto = handle:headers():get("x-forwarded-proto")
                      handle:streamInfo():dynamicMetadata():set(
                        "com.suse.observability.router", "forwarded_proto", proto or "")
                    end
                    function envoy_on_response(handle)
                      local meta = handle:streamInfo():dynamicMetadata():get(
                        "com.suse.observability.router")
                      if not meta or meta["forwarded_proto"] ~= "https" then
                        return
                      end
                      local headers = handle:headers()
                      local cookies = {}
                      for key, value in pairs(headers) do
                        if string.lower(key) == "set-cookie" then
                          table.insert(cookies, value)
                        end
                      end
                      if #cookies == 0 then
                        return
                      end
                      headers:remove("set-cookie")
                      for _, cookie in ipairs(cookies) do
                        if not string.find(string.lower(cookie), ";%s*secure") then
                          cookie = cookie .. "; Secure"
                        end
                        headers:add("set-cookie", cookie)
                      end
                    end
            {{- end }}
            - name: envoy.filters.http.router
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters.yaml: |-
    resources:
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http_protocol_options: {}
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-main"
        endpoints:
        {{- if eq .RouterState "active" }}
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.api.target.service.fullname" . }}"
                  port_value: 7070
        {{- end }}
      {{- if eq (include "stackstate.receiver.split.enabled" .) "true" }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver-logs"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver-logs"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.receiver.logs.service.fullname" . }}"
                  port_value: 7077
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver-process-agent"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver-process-agent"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.receiver.processAgent.service.fullname" . }}"
                  port_value: 7077
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver-base"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver-base"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.receiver.base.service.fullname" . }}"
                  port_value: 7077
      {{- else }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "common.fullname.short" . }}-receiver"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-receiver"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.receiver.service.fullname" . }}"
                  port_value: 7077
      {{- end }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ include "stackstate.otelCollector.fullname" . }}"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ include "stackstate.otelCollector.fullname" . }}"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.otelCollector.service.fullname" . }}"
                  port_value: 4318
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "stackstate.ui.fullname" . }}"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http_protocol_options: {}
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "common.fullname.short" . }}-{{ template "stackstate.router.api.name" . }}-ui"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.ui.service.fullname" . }}"
                  port_value: 8080
    {{- if eq (include "stackstate.mcp.enabled" .) "true" }}
    - "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: "{{ template "stackstate.mcp.fullname" . }}"
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http_protocol_options: {}
      connect_timeout: 2s
      {{- with include "stackstate.router.cluster.buffers" $ }}{{ . | nindent 6 }}{{ end }}
      load_assignment:
        cluster_name: "{{ template "stackstate.mcp.fullname" . }}"
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{{ include "stackstate.mcp.service.fullname" . }}"
                  port_value: {{ include "stackstate.mcp.port" . }}
    {{- end }}
{{- end -}}

{{/*
Kernel socket buffers are charged to the router's memory cgroup but not to Envoy's own accounting. Autotuned send
buffers towards slow receivers, and receive buffers for concurrent uploads, grow faster than the overload manager reacts.
*/}}
{{- define "stackstate.router.listener.buffers" -}}
per_connection_buffer_limit_bytes: 65536
socket_options:
- description: SO_RCVBUF, inherited by accepted connections
  level: 1
  name: 8
  int_value: 65536
  state: STATE_PREBIND
{{- end -}}

{{- define "stackstate.router.cluster.buffers" -}}
per_connection_buffer_limit_bytes: 65536
upstream_bind_config:
  socket_options:
  - description: SO_SNDBUF
    level: 1
    name: 7
    int_value: 65536
    state: STATE_PREBIND
{{- end -}}

{{/*
Driven by the container's cgroup memory, which includes the kernel socket buffers that Envoy's heap accounting misses.
Rejecting connections closes them at once; stop_accepting_connections leaves clients hanging in the kernel backlog.
*/}}
{{- define "stackstate.router.overloadManager" -}}
overload_manager:
  refresh_interval: 0.25s
  resource_monitors:
  - name: envoy.resource_monitors.cgroup_memory
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.resource_monitors.cgroup_memory.v3.CgroupMemoryConfig
  - name: envoy.resource_monitors.global_downstream_max_connections
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.resource_monitors.downstream_connections.v3.DownstreamConnectionsConfig
      max_active_downstream_connections: 50000
  actions:
  - name: envoy.overload_actions.shrink_heap
    triggers:
    - name: envoy.resource_monitors.cgroup_memory
      threshold:
        value: 0.5
  - name: envoy.overload_actions.disable_http_keepalive
    triggers:
    - name: envoy.resource_monitors.cgroup_memory
      threshold:
        value: 0.8
  - name: envoy.overload_actions.stop_accepting_requests
    triggers:
    - name: envoy.resource_monitors.cgroup_memory
      threshold:
        value: 0.8
  - name: envoy.overload_actions.reject_incoming_connections
    triggers:
    - name: envoy.resource_monitors.cgroup_memory
      threshold:
        value: 0.9
{{- end -}}
