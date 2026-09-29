package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHBaseBackupEnvNamesPreserveLegacyIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, release, expected string
		values                  map[string]string
	}{
		{"default", "olly", "olly-backup-sts-hbase-backup", nil},
		{"release-matches-base", "backup", "backup-sts-hbase-backup", nil},
		{"local-overrides", "olly", "olly-backup-sts-hbase-backup", map[string]string{"fullnameOverride": "custom-hbase", "fullnamePrefix": "local-", "fullnameSuffix": "-local"}},
		{"global-prefix-suffix", "nightly", "global-nightly-backup-end-sts-hbase-backup", map[string]string{"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end"}},
		{"global-override", "nightly", "existing-backup-sts-hbase-backup", map[string]string{"global.fullnameOverride": "Existing-Backup"}},
		{"truncation", "nightly", strings.Repeat("a", 53) + "-sts-hbase-backup", map[string]string{"global.fullnameOverride": strings.Repeat("a", 53) + "-tail"}},
	} {
		for _, mode := range []string{"Mono", "Distributed"} {
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				values := map[string]string{"deployment.mode": mode}
				for key, value := range tc.values {
					values[key] = value
				}
				options := &helm.Options{ValuesFiles: []string{"values/full.yaml"}, SetValues: values, Logger: logger.Discard, KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"}}
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, tc.release, options)
				resources := helmtestutil.NewKubernetesResources(t, output)
				refs := 0
				for _, sts := range resources.Statefulsets {
					for _, container := range sts.Spec.Template.Spec.Containers {
						if container.Name != "regionserver" && container.Name != "stackgraph" {
							continue
						}
						require.Len(t, container.EnvFrom, 2)
						require.NotNil(t, container.EnvFrom[0].ConfigMapRef)
						require.NotNil(t, container.EnvFrom[1].SecretRef)
						assert.Equal(t, tc.expected, container.EnvFrom[0].ConfigMapRef.Name)
						assert.Equal(t, tc.expected, container.EnvFrom[1].SecretRef.Name)
						refs++
					}
				}
				assert.Equal(t, 1, refs, "the storage container must retain both required backup environment references")
			})
		}
	}
}

func TestHBaseBackupEnvReferencesFollowSeparateHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_backup-name-overrides.tpl"), []byte(`
{{- define "stackstate.backup.hbase.configmap.fullname" -}}independent-backup-config{{- end -}}
{{- define "stackstate.backup.hbase.secret.fullname" -}}independent-backup-credentials{{- end -}}
`), 0600))
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, mode := range []string{"Mono", "Distributed"} {
		t.Run(mode, func(t *testing.T) {
			options := &helm.Options{ValuesFiles: []string{values}, SetValues: map[string]string{"deployment.mode": mode}, Logger: logger.Discard}
			before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
			output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
			require.NoError(t, err)
			after := helmtestutil.NewKubernetesResources(t, output)
			refs := 0
			for name, sts := range before.Statefulsets {
				expected := sts.DeepCopy()
				for i := range expected.Spec.Template.Spec.Containers {
					for j := range expected.Spec.Template.Spec.Containers[i].EnvFrom {
						ref := &expected.Spec.Template.Spec.Containers[i].EnvFrom[j]
						if ref.ConfigMapRef != nil && ref.ConfigMapRef.Name == "nightly-backup-sts-hbase-backup" {
							ref.ConfigMapRef.Name = "independent-backup-config"
							refs++
						}
						if ref.SecretRef != nil && ref.SecretRef.Name == "nightly-backup-sts-hbase-backup" {
							ref.SecretRef.Name = "independent-backup-credentials"
							refs++
						}
					}
				}
				require.Contains(t, after.Statefulsets, name)
				assert.Equal(t, *expected, after.Statefulsets[name], fmt.Sprintf("%s storage identity and pod template", name))
			}
			assert.Equal(t, 2, refs)
			assert.Len(t, after.Statefulsets, len(before.Statefulsets))
			assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
			assert.Equal(t, before.Services, after.Services)
		})
	}
}
