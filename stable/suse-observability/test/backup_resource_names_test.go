package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

type backupNameFixture struct {
	helper, kind, legacy string
}

// Names are changed only in a disposable chart. Compare entire resources and
// embedded manual manifests, allowing only the expected identities/references.
// This does not imply that renaming the real backup PVCs is an upgrade strategy.
func TestBackupResourceReferencesFollowDedicatedHelpers(t *testing.T) {
	const prefix = "nightly-suse-observability-"
	fixtures := []backupNameFixture{
		{"log.configmap.fullname", "ConfigMap", prefix + "backup-log"},
		{"restore.scripts.configmap.fullname", "ConfigMap", prefix + "backup-restore-scripts"},
		{"configuration.configmap.fullname", "ConfigMap", prefix + "sts-backup-conf"},
		{"config.configmap.fullname", "ConfigMap", "suse-observability-backup-config"},
		{"config.secret.fullname", "Secret", "suse-observability-backup-config"},
		{"stackpacks.service.fullname", "Service", prefix + "backup-stackpacks"},
		{"stackgraph.tmp.persistentvolumeclaim.fullname", "PersistentVolumeClaim", prefix + "backup-stackgraph-tmp-data"},
		{"stackgraph.v2.tmp.persistentvolumeclaim.fullname", "PersistentVolumeClaim", prefix + "backup-stackgraph-v2-tmp-data"},
		{"configuration.persistentvolumeclaim.fullname", "PersistentVolumeClaim", prefix + "settings-backup-data"},
		{"stackgraph.cronjob.fullname", "CronJob", "suse-observability-backup-sg"},
		{"stackgraph.v2.cronjob.fullname", "CronJob", prefix + "backup-sg-v2"},
		{"configuration.cronjob.fullname", "CronJob", prefix + "backup-conf"},
		{"init.cronjob.fullname", "CronJob", prefix + "backup-init"},
		{"init.job.fullname", "Job", prefix + "backup-init-<timestamp>"},
		{"configuration.init.job.fullname", "Job", prefix + "init-pvc-<timestamp>"},
		{"clickhouse.cleanup.job.fullname", "Job", prefix + "ch-clean<timestamp>"},
		{"init.job.generateName", "Job", "backup-init-"},
		{"configuration.init.job.generateName", "Job", "init-pvc-"},
		{"clickhouse.cleanup.job.generateName", "Job", "ch-clean"},
		{"elasticsearch.list.job.fullname", "Job", "elasticsearch-list-snapshots"},
		{"elasticsearch.restore.job.fullname", "Job", "elasticsearch-restore-snapshot"},
		{"stackgraph.list.job.fullname", "Job", "stackgraph-list-backups"},
		{"stackgraph.restore.job.fullname", "Job", "stackgraph-restore-backup"},
		{"configuration.list.job.fullname", "Job", "configuration-list-backups"},
		{"configuration.restore.job.fullname", "Job", "configuration-restore-backup"},
		{"configuration.download.job.fullname", "Job", "configuration-download-backup"},
		{"configuration.upload.job.fullname", "Job", "configuration-upload-backup"},
		{"victoriaMetrics.list.job.fullname", "Job", "victoria-metrics-list-backups"},
		{"victoriaMetrics.restore.job.fullname", "Job", "victoria-metrics-restore-backup"},
		{"stackgraph.restore.persistentvolumeclaim.fullname", "PersistentVolumeClaim", "stackgraph-restore-backup"},
		{"hbase.configmap.fullname", "ConfigMap", "nightly-backup-sts-hbase-backup"},
		{"hbase.secret.fullname", "Secret", "nightly-backup-sts-hbase-backup"},
	}
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	names := string(data)
	renames := map[string]string{}
	for _, fixture := range fixtures {
		name := "stackstate.backup." + fixture.helper
		explicit := "explicit-" + strings.ToLower(strings.ReplaceAll(fixture.helper, ".", "-"))
		renames[fixture.kind+"/"+fixture.legacy] = explicit
		replacement := `{{- define "` + name + `" -}}` + explicit + `{{- end -}}`
		if strings.HasPrefix(fixture.helper, "hbase.") {
			// The parent overrides the shared-library interface in the copied chart.
			names += "\n" + replacement + "\n"
		} else {
			definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(name) + `" -\}\}.*?\{\{- end -\}\}`)
			require.Len(t, definition.FindAllString(names, -1), 1, name)
			names = definition.ReplaceAllString(names, replacement)
		}
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(names), 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	seen := map[string]bool{}
	for _, mode := range []string{"Mono", "Distributed"} {
		for _, enabled := range []bool{false, true} {
			for _, argo := range []bool{false, true} {
				for _, upgrade := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/enabled=%t/argo=%t/upgrade=%t", mode, enabled, argo, upgrade), func(t *testing.T) {
						options := apiResourceNameTestOptions(map[string]string{
							"hbase.deployment.mode":                             mode,
							"global.backup.enabled":                             fmt.Sprint(enabled),
							"deployment.compatibleWithArgoCD":                   fmt.Sprint(argo),
							"backup.stackGraph.scheduled.implementation":        "all",
							"backup.stackGraph.scheduled.tempData.size":         "12Gi",
							"backup.stackGraph.scheduled.tempData.storageClass": "scratch-v1",
							"backup.stackGraph.v2.tempData.size":                "13Gi",
							"backup.stackGraph.v2.tempData.storageClass":        "scratch-v2",
							"backup.stackGraph.restore.tempData.size":           "14Gi",
							"backup.stackGraph.restore.tempData.storageClass":   "restore-scratch",
							"backup.configuration.scheduled.pvc.size":           "15Gi",
							"backup.configuration.scheduled.pvc.storageClass":   "saved-settings",
						})
						options.ValuesFiles = []string{fullValues}
						var args []string
						if upgrade {
							args = append(args, "--is-upgrade")
						}
						before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
						after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
						require.NoError(t, err)
						expected := backupNamingDocuments(t, before, renames, seen)
						actual := backupNamingDocuments(t, after, nil, nil)
						assert.Equal(t, expected, actual, "Only the deliberately renamed resources and their references may differ")
					})
				}
			}
		}
	}
	for key := range renames {
		assert.True(t, seen[key], "resource was not exercised: %s", key)
	}
}

// Parse generated Jobs by name OR generateName, retaining every Argo CD hook.
// Embedded YAML is compared as data structures so missed references are caught
// without treating YAML formatting as application behavior.
func backupNamingDocuments(t *testing.T, output string, renames map[string]string, seen map[string]bool) map[string]interface{} {
	t.Helper()
	result := map[string]interface{}{}
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(output), 4096)
	for {
		var object unstructured.Unstructured
		err := decoder.Decode(&object)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if len(object.Object) == 0 || object.GetKind() == "Pod" {
			continue
		} // Random Helm test Pod suffix.
		backupNamingObject(t, object.Object, renames, seen)
		name := object.GetName()
		if name == "" {
			name = object.GetGenerateName()
		}
		key := object.GetKind() + "/" + object.GetNamespace() + "/" + name
		// Some dependencies emit the same pull Secret. Retain every
		// occurrence rather than dropping duplicate rendered resources.
		objects, _ := result[key].([]interface{})
		result[key] = append(objects, object.Object)
	}
	return result
}

var backupNamingTimestamp = regexp.MustCompile(`[0-9]{2}t[0-9]{6}$`)

func backupNamingObject(t *testing.T, object map[string]interface{}, renames map[string]string, seen map[string]bool) {
	t.Helper()
	kind, _ := object["kind"].(string)
	metadata, _ := object["metadata"].(map[string]interface{})
	for _, field := range []string{"name", "generateName"} {
		if name, ok := metadata[field].(string); ok {
			if kind == "Job" && field == "name" {
				name = backupNamingTimestamp.ReplaceAllString(name, "<timestamp>")
			}
			key := kind + "/" + name
			if replacement, ok := renames[key]; ok {
				name = replacement
				seen[key] = true
			}
			metadata[field] = name
		}
	}
	if kind == "ConfigMap" {
		if data, ok := object["data"].(map[string]interface{}); ok {
			for key, value := range data {
				text, ok := value.(string)
				if !ok {
					continue
				}
				if strings.HasSuffix(key, ".yaml") && (strings.HasPrefix(key, "job-") || strings.HasPrefix(key, "pvc-")) {
					var embedded map[string]interface{}
					require.NoError(t, k8syaml.Unmarshal([]byte(text), &embedded))
					backupNamingObject(t, embedded, renames, seen)
					data[key] = embedded
				} else if key == "config" {
					// Backup CLI configuration contains logging ConfigMap and settings PVC references.
					var config map[string]interface{}
					require.NoError(t, k8syaml.Unmarshal([]byte(text), &config))
					backupNamingReferences(config, renames)
					data[key] = config
				}
			}
		}
	}
	backupNamingReferences(object, renames)
}

func backupNamingReferences(value interface{}, renames map[string]string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, value := range node {
			switch key {
			case "configMap", "configMapRef", "secret", "secretRef", "persistentVolumeClaim":
				if ref, ok := value.(map[string]interface{}); ok {
					kind, field := "ConfigMap", "name"
					if key == "secret" || key == "secretRef" {
						kind = "Secret"
					}
					if key == "persistentVolumeClaim" {
						kind = "PersistentVolumeClaim"
						field = "claimName"
					}
					if name, ok := ref[field].(string); ok {
						if replacement, ok := renames[kind+"/"+name]; ok {
							ref[field] = replacement
						}
					}
				}
			case "loggingConfigConfigMap", "pvc":
				kind := "ConfigMap"
				if key == "pvc" {
					kind = "PersistentVolumeClaim"
				}
				if name, ok := value.(string); ok {
					if replacement, ok := renames[kind+"/"+name]; ok {
						node[key] = replacement
					}
				}
			case "value":
				// The backup initializer exposes this Service as an environment variable.
				if text, ok := value.(string); ok {
					old := "nightly-suse-observability-backup-stackpacks"
					if replacement, ok := renames["Service/"+old]; ok && text == "http://"+old+":7090" {
						node[key] = "http://" + replacement + ":7090"
					}
				}
			}
			backupNamingReferences(value, renames)
		}
	case []interface{}:
		for _, item := range node {
			backupNamingReferences(item, renames)
		}
	}
}
