package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// Pre-change source: c0971c0ee6756ccf7a4f23cc8655577fa3079e19.
// Normal tests restore only the six old naming expressions. The optional test
// compares against an independently exported Git revision with local dependencies.
func TestPeripheralConfigMapNamingCompatibility(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	restorePeripheralConfigMapNames(t, chart)
	checkPeripheralConfigMapNamingCompatibility(t, chart)
}

func TestPeripheralConfigMapNamingGitBaselineCompatibility(t *testing.T) {
	chart := os.Getenv("PERIPHERAL_CONFIGMAP_BASELINE_CHART")
	if chart == "" {
		t.Skip("set PERIPHERAL_CONFIGMAP_BASELINE_CHART to the pre-change chart with built local dependencies")
	}
	checkPeripheralConfigMapNamingCompatibility(t, chart)
}

func restorePeripheralConfigMapNames(t *testing.T, chart string) {
	t.Helper()
	path := filepath.Join(chart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for helper, body := range map[string]string{
		"stackstate.router.configmap.fullname":                 `{{ template "common.fullname.short" . }}-router`,
		"stackstate.stackpacks.scripts.configmap.fullname":     `{{ template "common.fullname.short" . }}-stackpacks-scripts`,
		"kubernetes-rbac-agent.url.configmap.fullname":         `{{ include "kubernetes-rbac-agent.app.name" . }}-url`,
		"kubernetes-rbac-agent.clusterName.configmap.fullname": `{{ include "kubernetes-rbac-agent.app.name" . }}-cluster-name`,
		"kafkaup-operator.configmap.fullname":                  `{{ template "common.fullname.short" . }}-kafkaup-config`,
		"stackstate.backup.hbase.configmap.fullname":           `{{ template "common.fullname.global" (merge (dict "Base" "backup") .) }}-sts-hbase-backup`,
	} {
		header := `{{- define "` + helper + `" -}}`
		pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
		require.Len(t, pattern.FindAllString(content, -1), 1, helper)
		content = pattern.ReplaceAllString(content, header+body+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func checkPeripheralConfigMapNamingCompatibility(t *testing.T, beforeChart string) {
	t.Helper()
	for _, tc := range []struct {
		name, release, prefix, operatorPrefix, backupPrefix, namespace, sizing string
		values                                                                 map[string]string
	}{
		{name: "default", release: "suse-observability"},
		{name: "nightly", release: "nightly"},
		{name: "namespace", release: "nightly", namespace: "tenant-b"},
		{name: "fullname", release: "nightly", prefix: "customer", values: map[string]string{"fullnameOverride": "customer"}},
		{name: "affixes", release: "nightly", prefix: "g-pre-customer-post-end",
			operatorPrefix: "g-nightly-kafkaup-operator-end", backupPrefix: "g-nightly-backup-end", values: map[string]string{
				"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
				"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
			}},
		{name: "global-override", release: "nightly", backupPrefix: "global-service", values: map[string]string{"global.fullnameOverride": "global-service"}},
		{name: "operator-override", release: "nightly", operatorPrefix: "customer-operator", values: map[string]string{"kafkaup-operator.fullnameOverride": "customer-operator"}},
		{name: "long-name", release: "nightly", prefix: strings.Repeat("a", 54), values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "monolithic", release: "nightly", values: map[string]string{"stackstate.features.server.split": "false", "hbase.deployment.mode": "Mono"}},
		{name: "nonha", release: "nightly", sizing: "global_sizing_50_nonha.yaml"},
		{name: "ha", release: "nightly", sizing: "global_sizing_150_ha.yaml"},
		{name: "maintenance", release: "nightly", values: map[string]string{"stackstate.components.router.mode.status": "maintenance"}},
		{name: "automatic", release: "nightly", values: map[string]string{"stackstate.components.router.mode.status": "automatic"}},
		{name: "argo", release: "nightly", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "stackpacks-s3", release: "nightly", values: map[string]string{
			"stackstate.stackpacks.source": "s3-bucket", "stackstate.stackpacks.s3.bucket": "customer-stackpacks",
		}},
		{name: "backups-disabled", release: "nightly", values: map[string]string{"global.backup.enabled": "false"}},
		{name: "agent-disabled", release: "nightly", values: map[string]string{"stackstate.k8sAuthorization.enabled": "false"}},
		{name: "operator-disabled", release: "nightly", values: map[string]string{"kafkaup-operator.enabled": "false"}},
		{name: "external-secrets", release: "nightly", values: map[string]string{
			"global.url.fromSecret": "customer-url", "global.clusterName.fromSecret": "customer-cluster",
			"global.s3proxy.credentials.fromExternalSecret": "customer-storage",
			"stackstate.apiKey.fromExternalSecret":          "customer-api-key", "stackstate.authentication.fromExternalSecret": "customer-auth",
		}},
		{name: "metadata", release: "nightly", values: map[string]string{"configmap.annotations.customer": "preserved"}},
	} {
		if tc.prefix == "" {
			tc.prefix = tc.release + "-suse-observability"
			if tc.release == "suse-observability" {
				tc.prefix = tc.release
			}
		}
		if tc.operatorPrefix == "" {
			tc.operatorPrefix = tc.release + "-kafkaup-operator"
		}
		if tc.backupPrefix == "" {
			tc.backupPrefix = tc.release + "-backup"
		}
		inventory := map[string]string{
			tc.prefix + "-router":                   "suse-observability-router",
			tc.prefix + "-stackpacks-scripts":       "suse-observability-stackpacks-scripts",
			tc.release + "-rbac-agent-url":          "suse-observability-rbac-agent-url",
			tc.release + "-rbac-agent-cluster-name": "suse-observability-rbac-agent-cluster-name",
			tc.operatorPrefix + "-kafkaup-config":   "suse-observability-kafkaup-config",
			tc.backupPrefix + "-sts-hbase-backup":   "suse-observability-hbase-backup",
		}
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				for key, value := range tc.values {
					values[key] = value
				}
				options := apiResourceNameTestOptions(values)
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.sizing != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.sizing))
				}
				beforeOutput := renderRBACNamingChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderRBACNamingChart(t, "..", tc.release, options, upgrade)
				before := helmtestutil.NewKubernetesResources(t, beforeOutput)
				after := helmtestutil.NewKubernetesResources(t, afterOutput)
				renames := map[string]string{}
				for old, target := range inventory {
					config, found := before.ConfigMaps[old]
					if !found {
						assert.NotContains(t, after.ConfigMaps, target, "Creation conditions must stay unchanged")
						continue
					}
					require.Contains(t, after.ConfigMaps, target)
					config.Name = target
					assert.Equal(t, config, after.ConfigMaps[target], "ConfigMap content and metadata must stay unchanged")
					if old != target {
						renames["ConfigMap/"+old] = target
						assert.NotContains(t, after.ConfigMaps, old)
					}
				}
				seen := map[string]bool{}
				expected := backupNamingDocuments(t, beforeOutput, renames, seen)
				actual := backupNamingDocuments(t, afterOutput, nil, nil)
				for name, deployment := range before.Deployments {
					require.Contains(t, after.Deployments, name)
					target := after.Deployments[name]
					allowed := deployment.DeepCopy()
					rewritePeripheralConfigMapReferences(&allowed.Spec.Template.Spec, renames)
					for _, volume := range deployment.Spec.Template.Spec.Volumes {
						if volume.ConfigMap == nil || renames["ConfigMap/"+volume.ConfigMap.Name] == "" {
							continue
						}
						key := map[string]string{
							tc.prefix + "-router":                 "checksum/router-configmap",
							tc.prefix + "-stackpacks-scripts":     "checksum/stackpack-scripts-configmap",
							tc.operatorPrefix + "-kafkaup-config": "checksum/configmap",
						}[volume.ConfigMap.Name]
						require.NotEmpty(t, key, "Every changed volume must have its existing checksum")
						oldHash, newHash := deployment.Spec.Template.Annotations[key], target.Spec.Template.Annotations[key]
						require.Regexp(t, "^[0-9a-f]{64}$", oldHash)
						require.Regexp(t, "^[0-9a-f]{64}$", newHash)
						assert.NotEqual(t, oldHash, newHash)
						allowed.Spec.Template.Annotations[key] = newHash
						doc := expected["Deployment/"+deployment.Namespace+"/"+name].([]interface{})[0].(map[string]interface{})
						template := doc["spec"].(map[string]interface{})["template"].(map[string]interface{})
						template["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[key] = newHash
					}
					assert.Equal(t, *allowed, target, "Only reviewed references and existing checksums may change")
					assertApplicationConfigMapUpgradePatches(t, deployment, target)
				}
				for name, statefulset := range before.Statefulsets {
					require.Contains(t, after.Statefulsets, name)
					target := after.Statefulsets[name]
					allowed := statefulset.DeepCopy()
					rewritePeripheralConfigMapReferences(&allowed.Spec.Template.Spec, renames)
					assert.Equal(t, *allowed, target, "Only the HBase ConfigMap reference may change; storage and immutable fields stay unchanged")
					assertConfigMapNamingUpgradePatches(t, &statefulset, &target)
				}
				assert.Equal(t, expected, actual, "Only the six reviewed ConfigMaps, references and checksums may change")
				for old := range renames {
					assert.True(t, seen[old], old)
				}
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
				assert.Equal(t, before.Secrets, after.Secrets)
				assert.Equal(t, before.Services, after.Services)
				assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
				original := backupNamingDocuments(t, beforeOutput, nil, nil)
				for key, document := range actual {
					if strings.HasPrefix(key, "Job/") {
						assert.Equal(t, original[key], document, "Immutable Jobs must not gain references to replacement ConfigMaps")
					}
				}
			})
		}
	}
}

func rewritePeripheralConfigMapReferences(pod *corev1.PodSpec, renames map[string]string) {
	for i := range pod.Volumes {
		if ref := pod.Volumes[i].ConfigMap; ref != nil {
			if target := renames["ConfigMap/"+ref.Name]; target != "" {
				ref.Name = target
			}
		}
	}
	for _, containers := range [][]corev1.Container{pod.Containers, pod.InitContainers} {
		for i := range containers {
			for j := range containers[i].EnvFrom {
				if ref := containers[i].EnvFrom[j].ConfigMapRef; ref != nil {
					if target := renames["ConfigMap/"+ref.Name]; target != "" {
						ref.Name = target
					}
				}
			}
		}
	}
}
