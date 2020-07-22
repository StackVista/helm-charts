{{/*
Shared settings in configmap for server and api
*/}}
{{- define "stackstate.configmap.server-and-api" }}
{{- $authTypes := list -}}
stackstate.api.authentication.authServer.k8sServiceAccountAuthServer {}
{{- if .Values.caspr.enabled }}
stackstate {
  tenantAware = true
  tenant {
    identifier = {{ .Values.caspr.subscription.tenant }}
    name = "{{ .Values.caspr.subscription.tenantobj.name }}"
    subscription {
{{- if .Values.caspr.subscription.expirationDate }}
      expirationDate = {{ .Values.caspr.subscription.expirationDate }}
{{- end }}
      plan = {{ .Values.caspr.subscription.planobj.name }}
    }
  }
{{- if .Values.caspr.keycloak }}
  {{- $authTypes = append $authTypes "keycloakAuthServer" -}}
  api {
    authentication {
      authServer {
        keycloakAuthServer {
          keycloakBaseUri = "{{ .Values.caspr.keycloak.url }}"
          realm = "{{ .Values.caspr.keycloak.realm }}"
          clientId = "{{ .Values.caspr.keycloak.client }}"
          authenticationMethod = "client_secret_basic"
          secret = "{{ .Values.caspr.keycloak.secret }}"
          redirectUri = "https://{{ .Values.caspr.applicationInstance.host }}/loginCallback"
          jwsAlgorithm = "RS256"
        }
      }
    }
  }
{{- end }}
}
{{- else }}
{{- if or (hasKey .Values.stackstate.authentication.ldap "bind") (hasKey .Values.stackstate.authentication.ldap "ssl") }}
{{ $authTypes = append $authTypes "ldapAuthServer" -}}
stackstate {
  api {
    authentication {
      authServer {
        ldapAuthServer {
          connection {
{{- if hasKey .Values.stackstate.authentication.ldap "bind" }}
            bindCredentials {
              dn = ${LDAP_BIND_DN}
              password = ${LDAP_BIND_PASSWORD}
            }
{{- end }}
{{- if hasKey .Values.stackstate.authentication.ldap "ssl" }}
            ssl {
              sslType = {{ required "stackstate.authentication.ldap.ssl.type is required when configuring LDAP SSL" .Values.stackstate.authentication.ldap.ssl.type }}
          {{- if .Values.stackstate.authentication.ldap.ssl.trustCertificates }}
              trustCertificatesPath = "/opt/docker/secrets/ldap-certificates.pem"
          {{- end }}
          {{- if .Values.stackstate.authentication.ldap.ssl.trustStore }}
              trustStorePath = "/opt/docker/secrets/ldap-cacerts"
          {{- end }}
            }
{{- end }}
          }
        }
      }
    }
  }
}
{{- else }}
{{- $authTypes = append $authTypes "stackstateAuthServer" -}}
{{- end }}
{{- $authTypes = append $authTypes "k8sServiceAccountAuthServer" }}
stackstate.api.authentication.authServer.authServerType = [ {{- $authTypes | compact | join ", " -}} ]
{{- end }}
{{- with index .Values "cluster-agent" -}}
{{- if .enabled }}
stackstate.stackPacks {
  installOnStartUp += "kubernetes"

  installOnStartUpConfig {
    kubernetes.kubernetes_cluster_name = {{ .stackstate.cluster.name | quote }}
  }
}
{{- end -}}
{{- end }}
{{- end -}}
