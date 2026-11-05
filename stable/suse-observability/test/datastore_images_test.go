package test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

func TestStackgraphImageVersionSource(t *testing.T) {
	output, err := exec.Command("bash", "../../../updatecli/test-stackgraph-image-version.sh").CombinedOutput()
	require.NoError(t, err, "%s", output)
}

type graphImageDefaults struct {
	Image struct{ Repository string }
}

type hbaseImageDefaults struct {
	Stackgraph struct {
		Version string
		Image   struct{ Repository string }
	}
	Console graphImageDefaults
	Tephra  graphImageDefaults
	Hbase   struct {
		Master       graphImageDefaults
		Regionserver graphImageDefaults
	}
}

func TestDatastoreImageDefaults(t *testing.T) {
	var defaults struct {
		Elasticsearch struct {
			ImageTag string `yaml:"imageTag"`
		}
		Hbase hbaseImageDefaults
	}
	data, err := os.ReadFile("../values.yaml")
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(data, &defaults))
	var localDefaults hbaseImageDefaults
	data, err = os.ReadFile("../../../local/hbase/values.yaml")
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(data, &localDefaults))
	require.NotEmpty(t, defaults.Hbase.Stackgraph.Version)
	require.Equal(t, defaults.Hbase.Stackgraph.Version, localDefaults.Stackgraph.Version)
	require.NotEmpty(t, defaults.Elasticsearch.ImageTag)

	for _, mode := range []string{"Distributed", "Mono"} {
		t.Run(mode, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues: map[string]string{
					"hbase.deployment.mode": mode,
					"hbase.console.enabled": "true",
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			images := map[string]bool{}
			collect := func(spec corev1.PodSpec) {
				for _, c := range spec.Containers {
					images[c.Image] = true
				}
			}
			for _, sts := range resources.Statefulsets {
				collect(sts.Spec.Template.Spec)
			}
			for _, deployment := range resources.Deployments {
				collect(deployment.Spec.Template.Spec)
			}
			repositories := map[string]string{
				localDefaults.Console.Image.Repository: defaults.Hbase.Console.Image.Repository,
				localDefaults.Tephra.Image.Repository:  defaults.Hbase.Tephra.Image.Repository,
			}
			if mode == "Mono" {
				repositories[localDefaults.Stackgraph.Image.Repository] = defaults.Hbase.Stackgraph.Image.Repository
			} else {
				repositories[localDefaults.Hbase.Master.Image.Repository] = defaults.Hbase.Hbase.Master.Image.Repository
				repositories[localDefaults.Hbase.Regionserver.Image.Repository] = defaults.Hbase.Hbase.Regionserver.Image.Repository
			}
			for fallback, repository := range repositories {
				require.NotEmpty(t, fallback)
				if repository == "" {
					repository = fallback
				}
				require.Contains(t, images, "my.registry.com/"+repository+":2.5-"+defaults.Hbase.Stackgraph.Version)
			}
			require.Contains(t, images, "my.registry.com/stackstate/elasticsearch:"+defaults.Elasticsearch.ImageTag)
		})
	}
}
