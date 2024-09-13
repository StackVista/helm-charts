To test with keycloak:
* Install keycloak, it will make an ingress route for https://auth-demo-keycloak.test.stackstate.io:
  ```
  helm repo add codecentric https://codecentric.github.io/helm-charts
  helm repo update
  helm upgrade --install keycloak codecentric/keycloak --values keycloak-values.yaml
  ```
* Import the keycloak-realm.json. That should create a StackState realm. To login to keycloak use username `admin` and password `password` (see also the `keycloak-values.yaml`)
* Now create a new test user in that realm and assign the user the `stackstate-admin` role
* Check in Keycloak that the StackState client secret in the StackState realm matches the one in the `keycloak-values.yaml`. If not update the yaml file.
* Install StackState using your normal `values.yaml`, the `ingress.yaml` in the `auth` directory and the `keycloak-auth.yaml`. StackState will be installed with an ingress of https://demo-auth.test.stackstate.io and expects to find keycloak at https://auth-demo-keycloak.test.stackstate.io
* Now open the stackstate url in a browser and you should be redirected to Keycloak where you can login with the user you just created.
