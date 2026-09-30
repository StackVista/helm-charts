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

func TestAPIKeySecretLookupPreservesCredentials(t *testing.T) {
	const storedKey = "stored-fixture-api-key"
	chart := secretLookupTestChart(t, "templates/global/secret-api-key.yaml")
	renamedChart := secretLookupTestChart(t, "templates/global/secret-api-key.yaml")
	replaceAPIKeySecretName(t, renamedChart)
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, tc := range []struct {
		name, release, namespace, secretName, provided, expected string
		values                                                   map[string]string
		renamed, missing, emptyStored, external, global          bool
	}{
		{name: "default", release: "suse-observability", secretName: "suse-observability-api-key", expected: storedKey},
		{name: "custom-release-and-namespace", namespace: "customer-space", expected: storedKey},
		{name: "fullname-override", secretName: "custom-api-key", values: map[string]string{"fullnameOverride": "CUSTOM"}, expected: storedKey},
		{name: "prefix-suffix", secretName: "global-pre-custom-post-end-api-key", values: map[string]string{
			"fullnameOverride": "custom", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end",
		}, expected: storedKey},
		{name: "truncated", secretName: strings.Repeat("a", 54) + "-api-key", values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}, expected: storedKey},
		{name: "dedicated-helper", secretName: "explicit-api-key-secret", renamed: true, expected: storedKey},
		{name: "same-key", provided: storedKey, expected: storedKey},
		{name: "replacement", provided: "replacement-key", expected: "replacement-key"},
		{name: "fresh-with-key", missing: true, provided: "new-key", expected: "new-key"},
		{name: "missing-secret-and-key", missing: true},
		{name: "empty-stored-key", emptyStored: true},
		{name: "external", external: true, renamed: true, provided: "ignored-key"},
		{name: "global-external", global: true, external: true, renamed: true, provided: "ignored-key"},
		{name: "global-omitted", global: true},
		{name: "global-provided", global: true, provided: "global-key", expected: "global-key"},
		{name: "global-precedence", global: true, provided: "global-key", expected: "global-key", values: map[string]string{
			"stackstate.apiKey.key": "legacy-key", "global.receiverApiKey": "older-key", "stackstate.receiver.apiKey": "oldest-key",
		}},
		{name: "global-legacy-fallback", global: true, expected: "legacy-key", values: map[string]string{
			"stackstate.apiKey.key": "legacy-key", "global.receiverApiKey": "older-key", "stackstate.receiver.apiKey": "oldest-key",
		}},
		{name: "legacy-value-precedence", provided: "older-key", expected: "legacy-key", values: map[string]string{
			"stackstate.apiKey.key": "legacy-key", "stackstate.receiver.apiKey": "oldest-key",
		}},
		{name: "oldest-value-fallback", expected: "oldest-key", values: map[string]string{"stackstate.receiver.apiKey": "oldest-key"}},
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
					secretName = "nightly-suse-observability-api-key"
				}
				secretPath := "/api/v1/namespaces/" + namespace + "/secrets/" + secretName
				var existing *corev1.Secret
				if !tc.missing {
					key := storedKey
					if tc.emptyStored {
						key = ""
					}
					existing = &corev1.Secret{
						TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
						ObjectMeta: metav1.ObjectMeta{
							Name: secretName, Namespace: namespace,
							Labels: map[string]string{"app.kubernetes.io/managed-by": "Helm"},
							Annotations: map[string]string{
								"meta.helm.sh/release-name": release, "meta.helm.sh/release-namespace": namespace,
							},
						},
						Data: map[string][]byte{"API_KEY": []byte(key)},
					}
				}
				kubeconfig, requests := secretLookupTestAPI(t, secretPath, existing)
				values := map[string]string{
					"global.receiverApiKey":                   tc.provided,
					"global.suseObservability.receiverApiKey": "",
					"stackstate.apiKey.key":                   "",
					"stackstate.receiver.apiKey":              "",
				}
				if tc.global {
					values["global.suseObservability.sizing.profile"] = "trial"
					values["global.suseObservability.receiverApiKey"] = tc.provided
					values["global.receiverApiKey"] = ""
				}
				maps.Copy(values, tc.values)
				if tc.external {
					values["stackstate.apiKey.fromExternalSecret"] = "customer-api-key"
				}
				options := apiResourceNameTestOptions(values)
				options.KubectlOptions.Namespace = namespace
				options.ValuesFiles = []string{valuesFile}
				selectedChart := chart
				if tc.renamed {
					selectedChart = renamedChart
				}
				args := []string{"--dry-run=server", "--disable-openapi-validation", "--kubeconfig", kubeconfig, "--kube-context", "fixture"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output, err := helm.RenderTemplateE(t, options, selectedChart, release, nil, args...)
				require.NoError(t, err)
				if tc.external || (tc.global && tc.expected == "") {
					// No manifest is produced, so any Secret read would be a
					// template lookup, not Helm's resource ownership check.
					for _, path := range requests() {
						assert.NotContains(t, path, "/secrets/")
					}
				} else if !tc.global {
					require.Contains(t, requests(), secretPath, "Legacy lookup must follow the declaration's name and namespace")
				}
				resources := helmtestutil.NewKubernetesResources(t, output)
				if tc.expected == "" {
					assert.Empty(t, resources.Secrets)
				} else {
					require.Len(t, resources.Secrets, 1)
					require.Contains(t, resources.Secrets, secretName)
					assert.Equal(t, map[string][]byte{"API_KEY": []byte(tc.expected)}, resources.Secrets[secretName].Data)
				}
			})
		}
	}
}
