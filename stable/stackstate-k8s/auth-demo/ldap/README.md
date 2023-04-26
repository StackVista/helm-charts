
Install apacheds with `kubectl apply -f ldap-server-deployment.yaml`. See also https://gitlab.com/stackvista/stackstate-installation/-/tree/master/plugin-products/apacheds for details on the dockeri image used.

Configure stackstate helm chart with `ldap-auth.yaml` values file (next to the standard generated values you already have). Assumption is they will be in the same namespace
