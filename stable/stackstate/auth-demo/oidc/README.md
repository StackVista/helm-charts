Testing OIDC is possible with any OIDC provider, though configuring Gitlab for this is one of the easiest ways to test this. Another example is google, but that is somewhat more complicated to setup (and google doesn't have any groups or roles by default).

For this demo we assume the `ingress.yaml` in the `auth` directory will be used making StackState accessible at https://auth-demo.test.stackstate.io.
To setup OIDC go to your Gitlab account settings -> Applications.  Add an application here:
* Name: Anything you like
* Redirect URI: https://auth-demo.test.stackstate.io/loginCallback
* Scopes: openid, email

Save it and copy/paste the client id and secret into the `gitlab-oidc.yaml` file.

Install StackState as usual but include the extra yaml files to setup ingress and configure OIDC authentication: `--values gitlab-oidc.yaml --values ../ingress.yaml`. You can of course change the URL in the `ingress.yaml` file, but make sure that the Redirect URI in your Gitlab Application config uses that same hostname.

After deployment go to StackState (if you followed this example https://auth-demo.test.stackstate.io). You should be redirected to Gitlab and need to give permission to StackState there, after which you should be logged in. The OIDC configuration is such that if you're a member of the `stackvista` group on Gitlab you have admin access to StackState.
