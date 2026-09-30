package test

import (
	"fmt"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAuthSecretLookupPreservesCredentials(t *testing.T) {
	const storedPassword = "$2b$10$N9qo8uLOickgx2ZMRZoMye.IUIrCxGvz9/6pN.XMlqL9hzTgDaPGy"
	const replacementPassword = "098f6bcd4621d373cade4e832627b4f6"
	renamedChart := authSecretLookupTestChart(t, true)
	chart := authSecretLookupTestChart(t, false)
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, tc := range []struct {
		name, release, namespace, secretName, provided string
		values                                         map[string]string
		renamed, missing, external                     bool
	}{
		{name: "default", release: "suse-observability", secretName: "suse-observability-auth"},
		{name: "custom-release-and-namespace", release: "nightly", namespace: "customer-space", secretName: "nightly-suse-observability-auth"},
		{name: "fullname-override", secretName: "custom-auth", values: map[string]string{"fullnameOverride": "custom"}},
		{name: "prefix-suffix", secretName: "global-pre-custom-post-end-auth", values: map[string]string{
			"fullnameOverride": "custom", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end",
		}},
		{name: "truncated", secretName: strings.Repeat("a", 54) + "-auth", values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "dedicated-helper", secretName: "explicit-auth-secret", renamed: true},
		{name: "same-password", provided: storedPassword},
		{name: "explicit-replacement", provided: replacementPassword},
		{name: "fresh-with-password", missing: true, provided: replacementPassword},
		{name: "missing-secret-and-password", missing: true},
		{name: "external-secret", external: true, renamed: true},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				release, namespace, secretName := tc.release, tc.namespace, tc.secretName
				if release == "" {
					release = "nightly"
				}
				if namespace == "" {
					namespace = "observability"
				}
				if secretName == "" {
					secretName = "nightly-suse-observability-auth"
				}
				secretPath := "/api/v1/namespaces/" + namespace + "/secrets/" + secretName
				var existing *corev1.Secret
				if !tc.missing {
					existing = &corev1.Secret{
						TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
						ObjectMeta: metav1.ObjectMeta{
							Name: secretName, Namespace: namespace,
							Labels: map[string]string{"app.kubernetes.io/managed-by": "Helm"},
							Annotations: map[string]string{
								"meta.helm.sh/release-name": release, "meta.helm.sh/release-namespace": namespace,
							},
						},
						Data: map[string][]byte{"default_password": []byte(storedPassword)},
					}
				}
				kubeconfig, requests := secretLookupTestAPI(t, secretPath, existing)
				values := map[string]string{"stackstate.authentication.adminPassword": tc.provided}
				maps.Copy(values, tc.values)
				if tc.external {
					values["stackstate.authentication.fromExternalSecret"] = "customer-auth"
				}
				options := apiResourceNameTestOptions(values)
				options.KubectlOptions.Namespace = namespace
				options.ValuesFiles = []string{valuesFile}
				selectedChart := chart
				if tc.renamed {
					selectedChart = renamedChart
				}
				// Always override kubeconfig explicitly: this test only talks to
				// its read-only mock API, never the developer's current cluster.
				args := []string{"--dry-run=server", "--disable-openapi-validation", "--kubeconfig", kubeconfig, "--kube-context", "fixture"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output, err := helm.RenderTemplateE(t, options, selectedChart, release, nil, args...)
				if tc.external {
					require.NoError(t, err)
					for _, path := range requests() {
						assert.NotContains(t, path, "/secrets/"+secretName)
						assert.NotContains(t, path, "/secrets/explicit-auth-secret")
						assert.NotContains(t, path, "/secrets/customer-auth")
					}
					resources := helmtestutil.NewKubernetesResources(t, output)
					assert.NotContains(t, resources.Secrets, secretName)
					assert.NotContains(t, resources.Secrets, "explicit-auth-secret")
					assert.NotContains(t, resources.Secrets, "customer-auth")
					return
				}
				require.Contains(t, requests(), secretPath, "Lookup must use the same name and namespace as the Secret declaration")
				if tc.missing && tc.provided == "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), "Admin password is required for new installations")
					return
				}
				require.NoError(t, err)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Secrets, secretName)
				expectedPassword := storedPassword
				if tc.provided != "" {
					expectedPassword = tc.provided
				}
				assert.Equal(t, expectedPassword, string(resources.Secrets[secretName].Data["default_password"]))
			})
		}
	}
}

func authSecretLookupTestChart(t *testing.T, rename bool) string {
	t.Helper()
	chart := secretLookupTestChart(t, "templates/global/secret-auth.yaml")
	if rename {
		replaceAuthSecretName(t, chart)
	}
	return chart
}
