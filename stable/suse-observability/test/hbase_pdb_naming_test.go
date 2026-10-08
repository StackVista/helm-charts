package test

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHBasePDBNamingPreservesPlatformResources(t *testing.T) {
	beforeChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(beforeChart, os.DirFS("..")))
	// Only the six old PDB expressions are restored; no runtime consumers
	// are renamed. Frozen subchart snapshots independently protect this delta.
	definitions := ""
	for helper, suffix := range map[string]string{
		"hbase.master": "hbase-master", "hbase.regionserver": "hbase-rs",
		"hdfs.namenode": "hdfs-nn", "hdfs.secondarynamenode": "hdfs-snn",
		"hdfs.datanode": "hdfs-dn", "tephra": "tephra",
	} {
		definitions += `{{- define "hbase.` + helper + `.poddisruptionbudget.fullname" -}}{{ template "common.fullname.short" . }}-` + suffix + `{{- end -}}` + "\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(beforeChart, "templates", "_hbase-pdb-baseline.tpl"), []byte(definitions), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	afterChart, err := filepath.Abs("..")
	require.NoError(t, err)
	for _, tc := range []struct {
		name, release, namespace, base, sizing string
		count                                  int
		values                                 map[string]string
	}{
		{name: "product-release", release: "suse-observability", base: "suse-observability-hbase", count: 6},
		{name: "custom-release", release: "nightly", base: "nightly-hbase", count: 6},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", base: "suse-observability-hbase", count: 6},
		{name: "root-override", release: "nightly", base: "nightly-hbase", count: 6, values: map[string]string{"fullnameOverride": "customer-platform"}},
		{name: "subchart-affixes", release: "nightly", base: "g-pre-customer-post-end", count: 6, values: map[string]string{
			"hbase.fullnameOverride": "customer", "hbase.fullnamePrefix": "pre-", "hbase.fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "long-name", release: "nightly", base: strings.Repeat("a", 54), count: 6, values: map[string]string{"hbase.fullnameOverride": strings.Repeat("a", 70)}},
		{name: "mono", release: "nightly", base: "nightly-hbase", count: 1, values: map[string]string{"hbase.deployment.mode": "Mono"}},
		{name: "ha-overrides-legacy-mono", release: "nightly", base: "nightly-hbase", count: 6, sizing: "global_sizing_150_ha.yaml", values: map[string]string{"hbase.deployment.mode": "Mono"}},
		{name: "nonha-overrides-legacy-distributed", release: "nightly", base: "nightly-hbase", count: 1, sizing: "global_sizing_50_nonha.yaml"},
		{name: "argo", release: "nightly", base: "nightly-hbase", count: 6, values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "secondary-disabled", release: "nightly", base: "nightly-hbase", count: 5, values: map[string]string{"hbase.hdfs.secondarynamenode.enabled": "false"}},
		{name: "shared-monitor-namespace", release: "nightly", base: "nightly-hbase", count: 6, values: map[string]string{"hbase.servicemonitor.namespace": "shared-monitoring"}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := map[string]string{
					"hbase.deployment.mode": "Distributed", "hbase.hdfs.secondarynamenode.enabled": "true",
					"hbase.all.metrics.servicemonitor.enabled":     "true",
					"hbase.hdfs.datanode.extraEnv.secret.TEST_KEY": "fixture-value",
				}
				maps.Copy(values, tc.values)
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{valuesFile}
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.sizing != "" {
					sizing, err := filepath.Abs(filepath.Join("values", tc.sizing))
					require.NoError(t, err)
					options.ValuesFiles = append(options.ValuesFiles, sizing)
				}
				args := []string{"--api-versions", "policy/v1/PodDisruptionBudget", "--api-versions", "monitoring.coreos.com/v1"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				beforeOutput, err := helm.RenderTemplateE(t, options, beforeChart, tc.release, nil, args...)
				require.NoError(t, err)
				afterOutput, err := helm.RenderTemplateE(t, options, afterChart, tc.release, nil, args...)
				require.NoError(t, err)
				beforeResources := helmtestutil.NewKubernetesResources(t, beforeOutput)
				afterResources := helmtestutil.NewKubernetesResources(t, afterOutput)
				renames := map[string]string{}
				for _, component := range []string{"hbase-master", "hbase-rs", "hdfs-nn", "hdfs-snn", "hdfs-dn", "tephra"} {
					old := tc.base + "-" + component
					if _, found := beforeResources.Pdbs[old]; found {
						renames["PodDisruptionBudget/"+old] = "suse-observability-" + component
					}
				}
				require.Len(t, renames, tc.count, "mode-dependent PDB inventory must stay unchanged")
				seen := map[string]bool{}
				before := backupNamingDocuments(t, beforeOutput, renames, seen)
				after := backupNamingDocuments(t, afterOutput, nil, nil)
				assert.Equal(t, before, after, "Only HBase PDB metadata names may change")
				beforeContract := resourceNamingUpgradeContract(t, beforeOutput)
				for old, target := range renames {
					require.True(t, seen[old], "budget not exercised: %s", old)
					require.Contains(t, beforeContract, old)
					fields := beforeContract[old]
					delete(beforeContract, old)
					require.NotContains(t, beforeContract, "PodDisruptionBudget/"+target)
					beforeContract["PodDisruptionBudget/"+target] = fields
				}
				assert.Equal(t, beforeContract, resourceNamingUpgradeContract(t, afterOutput))
				assert.Equal(t, beforeResources.Statefulsets, afterResources.Statefulsets)
				assert.Equal(t, beforeResources.Services, afterResources.Services)
				assert.Equal(t, beforeResources.ServiceMonitors, afterResources.ServiceMonitors)
				assert.Equal(t, beforeResources.PersistentVolumeClaims, afterResources.PersistentVolumeClaims)
			})
		}
	}
}
