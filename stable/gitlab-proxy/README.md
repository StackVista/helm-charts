# gitlab-proxy

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.29.0](https://img.shields.io/badge/AppVersion-1.29.0-informational?style=flat-square)

A Helm chart for GitLab Proxy - Nginx-based caching proxy for GitLab packages registry

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchLabels | object | `{}` |  |
| affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey | string | `"kubernetes.io/hostname"` |  |
| affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight | int | `1` |  |
| auth.htpasswd | string | `"test:$apr1$lpJpERMg$rjNCT9/GOewTido8UmQYD.\n"` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"docker.io/bitnami/nginx"` |  |
| image.tag | string | `"1.29.0-debian-12-r2"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | string | `nil` |  |
| ingress.className | string | `"ingressClass"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"gitlab.proxy"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
| ingress.tls[0].hosts[0] | string | `"gitlab.proxy"` |  |
| ingress.tls[0].secretName | string | `"secret-tls"` |  |
| initContainer.resources.limits.cpu | string | `"150m"` |  |
| initContainer.resources.limits.ephemeral-storage | string | `"2Gi"` |  |
| initContainer.resources.limits.memory | string | `"192Mi"` |  |
| initContainer.resources.requests.cpu | string | `"100m"` |  |
| initContainer.resources.requests.ephemeral-storage | string | `"50Mi"` |  |
| initContainer.resources.requests.memory | string | `"128Mi"` |  |
| initContainer.securityContext.allowPrivilegeEscalation | bool | `false` |  |
| initContainer.securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| initContainer.securityContext.privileged | bool | `false` |  |
| initContainer.securityContext.readOnlyRootFilesystem | bool | `true` |  |
| initContainer.securityContext.runAsGroup | int | `1001` |  |
| initContainer.securityContext.runAsNonRoot | bool | `true` |  |
| initContainer.securityContext.runAsUser | int | `1001` |  |
| initContainer.securityContext.seLinuxOptions | object | `{}` |  |
| initContainer.securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| livenessProbe.failureThreshold | int | `6` |  |
| livenessProbe.initialDelaySeconds | int | `30` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.successThreshold | int | `1` |  |
| livenessProbe.tcpSocket.port | string | `"http"` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| nameOverride | string | `""` |  |
| nginx.serverBlock | string | `"\n\nresolver 10.0.224.1 valid=900s ipv6=off;\nresolver_timeout 5s;\n\nproxy_cache_path /opt/bitnami/nginx/tmp/nginx-cache levels=1:2 keys_zone=maven_cache:100m max_size=1g inactive=60d use_temp_path=off;\nproxy_cache_path /opt/bitnami/nginx/tmp/nginx-cache-metadata levels=1:2 keys_zone=maven_metadata_cache:10m max_size=100m inactive=1h use_temp_path=off;\n\nlog_format cache_debug '$remote_addr - $remote_user [$time_local] '\n                  '\"$request\" $status $body_bytes_sent '\n                  '\"$http_referer\" \"$http_user_agent\" \"$http_x_forwarded_for\" '\n                  '\"$upstream_addr\" \"$upstream_response_time\" \"$upstream_status\" '\n                  'cache_status:$upstream_cache_status '\n                  'upstream_cache_control:\"$upstream_http_cache_control\" '\n                  'upstream_expires:\"$upstream_http_expires\"';\n\nserver {\n  listen 0.0.0.0:8080;\n\n# Set cache and cache key once\n proxy_cache maven_cache;\n proxy_cache_key \"gitlab.com$request_uri\";\n\n  access_log /opt/bitnami/nginx/logs/cache_debug.log cache_debug;\n# Default cache settings for all content\n proxy_cache_valid 200 206 1y;\n proxy_cache_valid 404 1m;\n proxy_cache_valid any 0;\n\n# Cache optimization settings\n proxy_cache_use_stale error timeout invalid_header updating http_500 http_502 http_503 http_504;\n proxy_cache_background_update on;\n proxy_cache_lock on;\n proxy_cache_lock_timeout 5s;\n\n  # Force caching regardless of upstream headers\n  proxy_ignore_headers Cache-Control Expires Set-Cookie;\n  proxy_hide_header Set-Cookie;\n\n  # Cache bypass for no-cache requests\n  set $no_cache 0;\n  if ($http_cache_control ~* \"no-cache\") {\n    set $no_cache 1;\n  }\n  proxy_cache_bypass $no_cache;\n  proxy_no_cache $no_cache;\n\n  # Debug headers - ADD THESE!\n  add_header X-Cache-Status $upstream_cache_status always;\n  add_header X-Cache-Key \"gitlab.com$request_uri\" always;\n\n  # Override cache settings for metadata files only\n  location ~* /maven-metadata\\.xml$ {\n\n    if ($request_method ~ ^(PUT|POST|DELETE|PATCH)$) {\n        return 405;  # Method Not Allowed\n    }\n\n\n      auth_basic \"GitLab Packages Registry Proxy\";\n      auth_basic_user_file /etc/nginx/auth/.htpasswd;\n      proxy_pass https://gitlab.com;\n      proxy_set_header Host \"gitlab.com\";\n      proxy_set_header X-Real-IP $remote_addr;\n      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n      proxy_set_header X-Forwarded-Proto $scheme;\n      proxy_set_header PRIVATE-TOKEN \"GITLAB PERSONAL TOKEN WITH API SCOPE\";\n      proxy_set_header Authorization \"\";\n\n    proxy_http_version 1.1;\n    proxy_set_header Connection \"\";\n\n     proxy_ssl_verify off;\n     proxy_ssl_server_name on;\n     proxy_ssl_name gitlab.com;  # SNI hostname\n     proxy_ssl_protocols TLSv1.2 TLSv1.3;\n\n    # Override cache settings for metadata\n    proxy_cache_valid 200 206 1h;\n    proxy_cache_valid 404 1m;\n    expires 1h;\n  }\n\n  location / {\n\n    if ($request_method ~ ^(PUT|POST|DELETE|PATCH)$) {\n        return 405;  # Method Not Allowed\n    }\n      auth_basic \"GitLab Packages Registry Proxy\";\n      auth_basic_user_file /etc/nginx/auth/.htpasswd;\n      proxy_pass https://gitlab.com;\n      proxy_set_header Host \"gitlab.com\";\n      proxy_set_header X-Real-IP $remote_addr;\n      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n      proxy_set_header X-Forwarded-Proto $scheme;\n      proxy_set_header PRIVATE-TOKEN \"GITLAB PERSONAL TOKEN WITH API SCOPE\";\n      proxy_set_header Authorization \"\";\n\n\n    # Artifacts cache settings\n    proxy_cache_valid 200 206 1y;\n    proxy_cache_valid 404 1m;\n    expires 1y;\n    add_header Cache-Control \"public, immutable\";\n\n    proxy_http_version 1.1;\n    proxy_set_header Connection \"\";\n\n     proxy_ssl_verify off;\n     proxy_ssl_server_name on;\n     proxy_ssl_name gitlab.com;  # SNI hostname\n     proxy_ssl_protocols TLSv1.2 TLSv1.3;\n  }\n\n  # Health check endpoint\n  location /health {\n      access_log off;\n      return 200 \"healthy\\n\";\n      add_header Content-Type text/plain;\n  }\n}\n"` |  |
| nodeSelector | object | `{}` |  |
| persistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| persistence.size | string | `"200Gi"` |  |
| persistence.storageClass | string | `nil` |  |
| podAnnotations | object | `{}` |  |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.fsGroup | int | `1001` |  |
| podSecurityContext.fsGroupChangePolicy | string | `"Always"` |  |
| podSecurityContext.supplementalGroups | list | `[]` |  |
| podSecurityContext.sysctls | list | `[]` |  |
| readinessProbe.failureThreshold | int | `3` |  |
| readinessProbe.httpGet.path | string | `"/health"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.periodSeconds | int | `5` |  |
| readinessProbe.successThreshold | int | `1` |  |
| readinessProbe.timeoutSeconds | int | `3` |  |
| replicaCount | int | `2` |  |
| resources.limits.ephemeral-storage | string | `"2Gi"` |  |
| resources.limits.memory | string | `"256Mi"` |  |
| resources.requests.cpu | string | `"500m"` |  |
| resources.requests.ephemeral-storage | string | `"50Mi"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| revisionHistoryLimit | int | `10` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.privileged | bool | `false` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsGroup | int | `1001` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1001` |  |
| securityContext.seLinuxOptions | object | `{}` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `false` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |

