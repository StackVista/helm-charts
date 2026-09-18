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

func TestDatastoreImageDefaults(t *testing.T) {
	var defaults struct {
		Elasticsearch struct {
			ImageTag string `yaml:"imageTag"`
		}
		Hbase struct {
			Stackgraph struct{ Version string }
		}
	}
	data, err := os.ReadFile("../values.yaml")
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(data, &defaults))
	var localDefaults struct {
		Stackgraph struct{ Version string }
	}
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
			repositories := []string{"stackgraph-console", "tephra-server"}
			if mode == "Mono" {
				repositories = append(repositories, "stackgraph-hbase")
			} else {
				repositories = append(repositories, "hbase-master", "hbase-regionserver")
			}
			for _, repository := range repositories {
				require.Contains(t, images, "my.registry.com/stackstate/"+repository+":2.5-"+defaults.Hbase.Stackgraph.Version)
			}
			require.Contains(t, images, "my.registry.com/stackstate/elasticsearch:"+defaults.Elasticsearch.ImageTag)
		})
	}
}
