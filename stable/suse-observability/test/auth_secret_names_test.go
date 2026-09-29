package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestAuthSecretReferencesFollowDedicatedHelper(t *testing.T) {
	chart := authSecretNameTestChart(t)
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, split := range []bool{false, true} {
		for _, tc := range []struct {
			name, file, external string
		}{
			{name: "admin"},
			{name: "ldap", file: "ldap_authentication.yaml"},
			{name: "oidc", file: "oidc_authentication.yaml"},
			{name: "rancher", file: "rancher_authentication.yaml"},
			{name: "keycloak", file: "keycloak_authentication.yaml"},
			{name: "file", file: "file_authentication.yaml"},
			{name: "bootstrap", file: "bootstrap_token.yaml"},
			{name: "external", external: "customer-auth"},
		} {
			t.Run(fmt.Sprintf("%s/split=%t", tc.name, split), func(t *testing.T) {
				options := apiResourceNameTestOptions(map[string]string{
					"stackstate.features.server.split":             fmt.Sprint(split),
					"stackstate.authentication.fromExternalSecret": tc.external,
				})
				options.ValuesFiles = []string{valuesFile}
				if tc.file != "" {
					path, err := filepath.Abs(filepath.Join("values", tc.file))
					require.NoError(t, err)
					options.ValuesFiles = append(options.ValuesFiles, path)
				}
				before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
				require.NoError(t, err)
				after := helmtestutil.NewKubernetesResources(t, output)
				legacy, explicit := "nightly-suse-observability-auth", "explicit-auth-secret"
				selected := explicit
				if tc.external == "" {
					require.Contains(t, before.Secrets, legacy)
					require.Contains(t, after.Secrets, explicit)
					expected := before.Secrets[legacy]
					expected.Name = explicit
					assert.Equal(t, expected, after.Secrets[explicit], "Authentication data and bootstrap tokens must stay unchanged")
					delete(before.Secrets, legacy)
					delete(after.Secrets, explicit)
				} else {
					selected = tc.external
					assert.NotContains(t, before.Secrets, legacy)
					assert.NotContains(t, after.Secrets, explicit)
					assert.NotContains(t, after.Secrets, tc.external)
				}
				assert.NotContains(t, after.Secrets, legacy)
				assert.Equal(t, before.Secrets, after.Secrets)
				references := 0
				for name, deployment := range before.Deployments {
					require.Contains(t, after.Deployments, name)
					expected := deployment.DeepCopy()
					if tc.external == "" {
						for i := range expected.Spec.Template.Spec.Containers {
							for j := range expected.Spec.Template.Spec.Containers[i].EnvFrom {
								ref := expected.Spec.Template.Spec.Containers[i].EnvFrom[j].SecretRef
								if ref != nil && ref.Name == legacy {
									ref.Name = explicit
								}
							}
						}
						const checksum = "checksum/auth-env"
						if old, ok := expected.Spec.Template.Annotations[checksum]; ok {
							actual := after.Deployments[name].Spec.Template.Annotations
							require.Contains(t, actual, checksum)
							assert.NotEqual(t, old, actual[checksum], "Only deliberate test renames change the Secret's manifest checksum")
							expected.Spec.Template.Annotations[checksum] = actual[checksum]
						}
					}
					assert.Equal(t, *expected, after.Deployments[name])
					for _, c := range after.Deployments[name].Spec.Template.Spec.Containers {
						for _, source := range c.EnvFrom {
							if source.SecretRef != nil && source.SecretRef.Name == selected {
								references++
							}
						}
					}
				}
				assert.Equal(t, 1, references, "API or monolithic server must consume the selected Secret")
				component := "server"
				if split {
					component = "api"
				}
				require.Contains(t, after.Deployments, "nightly-suse-observability-"+component)
				assert.Contains(t, after.Deployments["nightly-suse-observability-"+component].Spec.Template.Spec.Containers[0].EnvFrom,
					corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: selected}}})
				assert.Len(t, after.Deployments, len(before.Deployments))
				assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
				assert.Equal(t, before.Services, after.Services)
				assert.Equal(t, before.Roles, after.Roles)
				assert.Equal(t, before.RoleBindings, after.RoleBindings)
				assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
				assert.Equal(t, before.Statefulsets, after.Statefulsets)
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
			})
		}
	}
}

func authSecretNameTestChart(t *testing.T) string {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	replaceAuthSecretName(t, chart)
	return chart
}

func replaceAuthSecretName(t *testing.T, chart string) {
	t.Helper()
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definition := regexp.MustCompile(`(?s)\{\{- define "stackstate.auth.secret.fullname" -\}\}.*?\{\{- end -\}\}`)
	require.Len(t, definition.FindAll(data, -1), 1)
	data = definition.ReplaceAll(data, []byte(`{{- define "stackstate.auth.secret.fullname" -}}explicit-auth-secret{{- end -}}`))
	require.NoError(t, os.WriteFile(path, data, 0600))
}
