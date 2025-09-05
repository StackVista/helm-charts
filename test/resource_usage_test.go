package test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
)

func TestCalculateResourceUsage(t *testing.T) {
	CalculateResourceUsage(t, "trial")
	CalculateResourceUsage(t, "10-nonha")
	CalculateResourceUsage(t, "20-nonha")
	CalculateResourceUsage(t, "50-nonha")
	CalculateResourceUsage(t, "100-nonha")
	CalculateResourceUsage(t, "150-ha")
	CalculateResourceUsage(t, "250-ha")
	CalculateResourceUsage(t, "500-ha")
	CalculateResourceUsage(t, "4000-ha")
}

func CalculateResourceUsage(t *testing.T, size string) {
	values := []string{"values/base.yaml"}

	output, err := helm.RenderTemplateE(t,
		&helm.Options{
			Logger:      logger.Discard,
			ValuesFiles: values,
			SetValues:   map[string]string{"sizing.profile": size},
		},
		"../stable/suse-observability-values", "test-values", []string{},
	)
	require.NoErrorf(t, err, "Failed to render values")

	parts := strings.Split(output, "\n---\n")
	files := []string{}

	for _, part := range parts {
		// write values to temporary file
		f, err := os.CreateTemp("", "tmpfile-")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())
		data := []byte(part)
		if _, err := f.Write(data); err != nil {
			t.Fatal(err)
		}
		f.Close()
		files = append(files, f.Name())
	}

	full, err := helm.RenderTemplateE(t,
		&helm.Options{
			Logger:      logger.Discard,
			ValuesFiles: files,
		},
		"../stable/suse-observability", "test-sizing", []string{},
	)
	require.NoErrorf(t, err, "Failed to render values")

	resources := helmtestutil.NewKubernetesResources(t, full)
	require.Equal(t, 1, len(resources.Pods), "Unexpected number of pods, expected 1")

	totalCpuRequests := resource.NewMilliQuantity(0, resource.DecimalSI)
	totalCpuLimits := resource.NewMilliQuantity(0, resource.DecimalSI)
	totalMemRequests := resource.NewQuantity(0, resource.DecimalSI)
	totalMemLimits := resource.NewQuantity(0, resource.DecimalSI)

	deployments := resources.Deployments
	for _, deployment := range deployments {
		replicas := int64(1)
		if deployment.Spec.Replicas != nil {
			replicas = int64(*deployment.Spec.Replicas)
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			cpuReq := container.Resources.Requests.Cpu().DeepCopy()
			cpuReq.Mul(replicas)
			totalCpuRequests.Add(cpuReq)
			cpuLim := container.Resources.Limits.Cpu().DeepCopy()
			cpuLim.Mul(replicas)
			totalCpuLimits.Add(cpuLim)
			memReq := container.Resources.Requests.Memory().DeepCopy()
			memReq.Mul(replicas)
			totalMemRequests.Add(memReq)
			memLim := container.Resources.Limits.Memory().DeepCopy()
			memLim.Mul(replicas)
			totalMemLimits.Add(memLim)
		}
	}

	statefulsets := resources.Statefulsets
	for _, statefulset := range statefulsets {
		replicas := int64(1)
		if statefulset.Spec.Replicas != nil {
			replicas = int64(*statefulset.Spec.Replicas)
		}
		for _, container := range statefulset.Spec.Template.Spec.Containers {
			cpuReq := container.Resources.Requests.Cpu().DeepCopy()
			cpuReq.Mul(replicas)
			totalCpuRequests.Add(cpuReq)
			cpuLim := container.Resources.Limits.Cpu().DeepCopy()
			cpuLim.Mul(replicas)
			totalCpuLimits.Add(cpuLim)
			memReq := container.Resources.Requests.Memory().DeepCopy()
			memReq.Mul(replicas)
			totalMemRequests.Add(memReq)
			memLim := container.Resources.Limits.Memory().DeepCopy()
			memLim.Mul(replicas)
			totalMemLimits.Add(memLim)
		}
	}

	t.Logf("Resource usage for profile %s\n", size)
	t.Logf("  cpu request: %v\n", totalCpuRequests)
	t.Logf("  cpu limit: %v\n", totalCpuLimits)
	t.Logf("  mem request: %v\n", totalMemRequests)
	t.Logf("  mem limit: %v\n", totalMemLimits)

	/*
		var valuesMap map[string]interface{}
		err = yaml.Unmarshal([]byte(output), &valuesMap)
		assert.NoError(t, err)

		decoder := yaml.NewDecoder(strings.NewReader(output))

		for {
			var doc map[string]interface{}
			err := decoder.Decode(&doc)
			if err != nil {
				if err.Error() == "EOF" {
					break // End of documents
				}
				assert.NoError(t, err)
			}

			err = mergo.Merge(&valuesMap, doc, mergo.WithOverride)
			require.NoErrorf(t, err, "Failed to merge map with values: %v", valuesMap)

			fmt.Printf("Document: %+v\n", doc)
		}

		fmt.Printf("valuesMap: %+#v\n", valuesMap)
	*/
}
