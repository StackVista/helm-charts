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
	helmtestutil.RequireTemplateValueAccessViaHelper(t, "../templates", "_configuration.tpl", `\.Values\.backup\.s3\b`, "clickhouse.backup.connection")
}

func TestBackupConnectionStandalone(t *testing.T) {
	for _, endpoint := range []string{"storage.example:9000", "{{ .Release.Name }}.example:9000"} {
		t.Run(endpoint, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "connection.yaml")
			require.NoError(t, os.WriteFile(path, []byte(fmt.Sprintf("backup:\n  s3:\n    endpoint: %q\n    secretName: '{{ .Release.Name }}-credentials'\n", endpoint)), 0600))
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "orders", &helm.Options{
				ValuesFiles: []string{"values/default.yaml", path},
				SetValues: map[string]string{
					"global.backup.enabled":                "true",
					"global.s3proxy.credentials.accessKey": "test",
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			expected := "storage.example:9000"
			if endpoint != expected {
				expected = "orders.example:9000"
			}
			configs := 0
			for _, cm := range resources.ConfigMaps {
				if config, ok := cm.Data["config.yaml"]; ok {
					assert.Contains(t, config, `endpoint: "http://`+expected+`"`)
					configs++
				}
			}
			assert.Equal(t, 1, configs)
			require.Len(t, resources.Statefulsets, 1)
		})
	}
}
