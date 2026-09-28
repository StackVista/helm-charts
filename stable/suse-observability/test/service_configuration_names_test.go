package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// Exercise invalid callers in an isolated chart copy. Existing component tests
// verify that real callers supply the names from their dedicated helpers.
func TestServiceConfigurationRequiresExplicitNames(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	deploymentPath := filepath.Join(chart, "templates", "api", "deployment-api.yaml")
	original, err := os.ReadFile(deploymentPath)
	require.NoError(t, err)
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, tc := range []struct {
		name, argument, helper, secretComponent, errorMessage string
		empty                                                 bool
	}{
		{
			name: "missing-config", argument: "configMapName", helper: "configmap",
			errorMessage: "stackstate.service.pod.volumes: configMapName must not be empty",
		},
		{
			name: "missing-log-config", argument: "logConfigMapName", helper: "log.configmap",
			errorMessage: "stackstate.service.pod.volumes: logConfigMapName must not be empty",
		},
		{
			name: "missing-secret-common-env", argument: "extraEnvSecretName", helper: "secret", secretComponent: "all",
			errorMessage: "stackstate.service.envvars: extraEnvSecretName must not be empty when secret environment variables are configured",
		},
		{
			name: "missing-secret-component-env", argument: "extraEnvSecretName", helper: "secret", secretComponent: "api",
			errorMessage: "stackstate.service.envvars: extraEnvSecretName must not be empty when secret environment variables are configured",
		},
		{name: "missing-secret-without-secret-env", argument: "extraEnvSecretName", helper: "secret"},
		{name: "empty-secret-without-secret-env", argument: "extraEnvSecretName", helper: "secret", empty: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			argument := ` "` + tc.argument + `" (include "stackstate.api.` + tc.helper + `.fullname" .)`
			require.Equal(t, 1, strings.Count(string(original), argument))
			replacement := ""
			if tc.empty {
				replacement = ` "` + tc.argument + `" ""`
			}
			content := strings.Replace(string(original), argument, replacement, 1)
			require.NoError(t, os.WriteFile(deploymentPath, []byte(content), 0600))

			values := map[string]string{
				"stackstate.features.server.split":                        "true",
				"stackstate.components.api.extraEnv.open.NON_SECRET_TEST": "plain-value",
			}
			if tc.secretComponent != "" {
				values["stackstate.components."+tc.secretComponent+".extraEnv.secret.SECRET_TEST"] = "secret-value"
			}
			options := apiResourceNameTestOptions(values)
			options.ValuesFiles = []string{valuesFile}
			output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
			if tc.errorMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				return
			}
			require.NoError(t, err)
			resources := helmtestutil.NewKubernetesResources(t, output)
			api := "nightly-suse-observability-api"
			require.Contains(t, resources.Deployments, api)
			pod := resources.Deployments[api].Spec.Template.Spec
			assertConfigurationVolumes(t, pod.Volumes, api, api+"-log")
			require.NotEmpty(t, pod.Containers)
			assert.Contains(t, pod.Containers[0].Env, corev1.EnvVar{Name: "NON_SECRET_TEST", Value: "plain-value"})
		})
	}
}
