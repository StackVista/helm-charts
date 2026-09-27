package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestConfigurationTemplatesUseHelper(t *testing.T) {
	helmtestutil.RequireTemplateValueAccessViaHelper(t, "../templates", "_configuration.tpl", `\.Values\.backup\.(overrideS3Endpoint|awsSecrets)\b`, "victoria-metrics.backup.connection")
}

func TestBackupConnectionStandalone(t *testing.T) {
	for _, endpoint := range []string{"https://storage.example", "https://{{ .Release.Name }}.example"} {
		t.Run(endpoint, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "connection.yaml")
			require.NoError(t, os.WriteFile(path, []byte(fmt.Sprintf("backup:\n  overrideS3Endpoint: %q\n  awsSecrets: '{{ .Release.Name }}-credentials'\n", endpoint)), 0600))
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "orders", &helm.Options{
				ValuesFiles: []string{"values/default.yaml", path},
				SetValues: map[string]string{
					"backup.s3Prefix":                      "orders",
					"backup.enabled":                       "true",
					"global.backup.enabled":                "true",
					"global.s3proxy.credentials.accessKey": "test",
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Len(t, resources.Statefulsets, 1)
			for _, sts := range resources.Statefulsets {
				expected := "https://storage.example"
				if endpoint != expected {
					expected = "https://orders.example"
				}
				foundEndpoint, credentials := false, 0
				for _, c := range sts.Spec.Template.Spec.InitContainers {
					for _, env := range c.Env {
						if env.Name == "S3_ENDPOINT" {
							assert.Equal(t, expected, env.Value)
							foundEndpoint = true
						}
					}
				}
				for _, c := range sts.Spec.Template.Spec.Containers {
					for _, env := range c.Env {
						if env.Name == "AWS_ACCESS_KEY_ID" || env.Name == "AWS_SECRET_ACCESS_KEY" {
							require.NotNil(t, env.ValueFrom)
							require.NotNil(t, env.ValueFrom.SecretKeyRef)
							assert.Equal(t, "orders-credentials", env.ValueFrom.SecretKeyRef.Name)
							credentials++
						}
					}
				}
				assert.True(t, foundEndpoint)
				assert.Equal(t, 2, credentials)
			}
		})
	}
}
