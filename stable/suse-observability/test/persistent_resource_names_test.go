package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// Deliberately rename only the extracted helpers in a copied chart. Every
// producer and consumer must follow independently, without changing claim specs,
// workload identities, mount paths, init containers or unrelated resources.
// This is a wiring test, not a supported PVC migration procedure.
func TestPersistentResourceReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	helpers := map[string]string{
		"api.txlog.persistentvolumeclaim":               "api-txlog",
		"authorizationSync.txlog.persistentvolumeclaim": "authorization-sync-txlog",
		"checks.txlog.persistentvolumeclaim":            "checks-txlog",
		"healthSync.txlog.persistentvolumeclaim":        "health-sync-txlog",
		"notification.txlog.persistentvolumeclaim":      "notification-txlog",
		"state.txlog.persistentvolumeclaim":             "state-txlog",
		"sync.txlog.persistentvolumeclaim":              "sync-txlog",
		"checks.tmp.persistentvolumeclaim":              "checks-tmp",
		"healthSync.tmp.persistentvolumeclaim":          "health-sync-tmp",
		"state.tmp.persistentvolumeclaim":               "state-tmp",
		"sync.tmp.persistentvolumeclaim":                "sync-tmp",
		"stackpacks.persistentvolumeclaim":              "stackpacks",
		"stackpacks.local.persistentvolumeclaim":        "stackpacks-local",
		"stackpacks.scripts.configmap":                  "stackpacks-scripts",
	}
	renames := map[string]string{}
	for helper, suffix := range helpers {
		name := "stackstate." + helper + ".fullname"
		explicit := "explicit-" + suffix
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(name) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAll(data, -1), 1, name)
		data = definition.ReplaceAll(data, []byte(`{{- define "`+name+`" -}}`+explicit+`{{- end -}}`))
		renames["nightly-suse-observability-"+suffix] = explicit
	}
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, split := range []bool{false, true} {
		for _, mode := range []string{"Mono", "Distributed"} {
			for _, enabled := range []bool{false, true} {
				t.Run(fmt.Sprintf("split=%t/%s/enabled=%t", split, mode, enabled), func(t *testing.T) {
					values := map[string]string{
						"stackstate.features.server.split":                           fmt.Sprint(split),
						"stackstate.features.storeTransactionLogsToPVC.enabled":      fmt.Sprint(enabled),
						"stackstate.features.storeTransactionLogsToPVC.volumeSize":   "7Gi",
						"stackstate.features.storeTransactionLogsToPVC.storageClass": "txlog-storage",
						"hbase.deployment.mode":                                      mode,
						"stackstate.stackpacks.pvc.size":                             "3Gi",
						"stackstate.stackpacks.pvc.storageClass":                     "images-storage",
						"stackstate.stackpacks.localpvc.size":                        "4Gi",
						"stackstate.stackpacks.localpvc.storageClass":                "local-storage",
					}
					for _, component := range []string{"checks", "healthSync", "state", "sync"} {
						key := "stackstate.components." + component + ".tmpToPVC"
						if enabled {
							values[key+".volumeSize"] = "12Gi"
							values[key+".storageClass"] = "tmp-storage"
						} else {
							values[key] = ""
						}
					}
					if !enabled {
						values["stackstate.stackpacks.source"] = "s3-bucket"
						values["stackstate.stackpacks.s3.bucket"] = "customer-stackpacks"
					}
					options := apiResourceNameTestOptions(values)
					options.ValuesFiles = []string{fullValues}
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)

					// Preserve the complete PVC, including annotations, class,
					// access modes, capacity and any explicit volume binding.
					assert.Len(t, after.PersistentVolumeClaims, len(before.PersistentVolumeClaims))
					for name, pvc := range before.PersistentVolumeClaims {
						expected := pvc.DeepCopy()
						if renamed, ok := renames[name]; ok {
							expected.Name = renamed
							assert.NotContains(t, after.PersistentVolumeClaims, name)
						}
						require.Contains(t, after.PersistentVolumeClaims, expected.Name)
						assert.Equal(t, *expected, after.PersistentVolumeClaims[expected.Name])
					}
					for helper, suffix := range helpers {
						if !strings.HasSuffix(helper, ".persistentvolumeclaim") {
							continue
						}
						exists := enabled && split
						if suffix == "stackpacks" {
							exists = enabled
						} else if suffix == "stackpacks-local" {
							exists = mode == "Mono"
						}
						_, legacyExists := before.PersistentVolumeClaims["nightly-suse-observability-"+suffix]
						_, explicitExists := after.PersistentVolumeClaims["explicit-"+suffix]
						assert.Equal(t, exists, legacyExists, suffix)
						assert.Equal(t, exists, explicitExists, suffix)
					}

					assert.Len(t, after.ConfigMaps, len(before.ConfigMaps))
					for name, config := range before.ConfigMaps {
						expected := config.DeepCopy()
						if name == "nightly-suse-observability-stackpacks-scripts" {
							expected.Name = renames[name]
							assert.NotContains(t, after.ConfigMaps, name)
						}
						for key, value := range expected.Data {
							expected.Data[key] = strings.ReplaceAll(value,
								"pvc: nightly-suse-observability-stackpacks-local",
								"pvc: explicit-stackpacks-local")
						}
						require.Contains(t, after.ConfigMaps, expected.Name)
						assert.Equal(t, *expected, after.ConfigMaps[expected.Name])
					}
					if mode == "Mono" {
						assert.Contains(t, after.ConfigMaps["suse-observability-backup-config"].Data["config"],
							"pvc: explicit-stackpacks-local")
					}

					references := map[string]int{}
					assert.Len(t, after.Deployments, len(before.Deployments))
					for name, deployment := range before.Deployments {
						require.Contains(t, after.Deployments, name)
						expected := deployment.DeepCopy()
						actual := after.Deployments[name]
						for i := range expected.Spec.Template.Spec.Volumes {
							volume := &expected.Spec.Template.Spec.Volumes[i]
							if volume.PersistentVolumeClaim != nil {
								if renamed, ok := renames[volume.PersistentVolumeClaim.ClaimName]; ok {
									references[renamed]++
									require.Contains(t, after.PersistentVolumeClaims, renamed)
									volume.PersistentVolumeClaim.ClaimName = renamed
								}
							}
							if volume.ConfigMap != nil {
								if renamed, ok := renames[volume.ConfigMap.Name]; ok {
									references[renamed]++
									require.Contains(t, after.ConfigMaps, renamed)
									volume.ConfigMap.Name = renamed
								}
							}
						}
						// Only the deliberate test rename changes this checksum.
						const checksum = "checksum/stackpack-scripts-configmap"
						if previous, ok := expected.Spec.Template.Annotations[checksum]; ok {
							require.NotEmpty(t, actual.Spec.Template.Annotations[checksum])
							assert.NotEqual(t, previous, actual.Spec.Template.Annotations[checksum])
							expected.Spec.Template.Annotations[checksum] = actual.Spec.Template.Annotations[checksum]
						}
						assert.Equal(t, *expected, actual)
					}
					for _, renamed := range renames {
						_, pvcExists := after.PersistentVolumeClaims[renamed]
						_, configExists := after.ConfigMaps[renamed]
						if pvcExists || configExists {
							assert.Positive(t, references[renamed], renamed+" must be mounted")
						} else {
							assert.Zero(t, references[renamed], renamed+" must not be referenced")
						}
					}
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
					assert.Equal(t, before.CronJobs, after.CronJobs)
					assert.Equal(t, before.Secrets, after.Secrets)
					afterJobs := commonSecretJobsByStableName(t, after.Jobs)
					assert.Len(t, afterJobs, len(before.Jobs))
					for name, job := range commonSecretJobsByStableName(t, before.Jobs) {
						require.Contains(t, afterJobs, name)
						expected := job.DeepCopy()
						expected.Name = afterJobs[name].Name // Existing timestamp only.
						assert.Equal(t, *expected, afterJobs[name])
					}
				})
			}
		}
	}
}
