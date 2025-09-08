package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
)

type resourceUsage struct {
	profile string

	totalCpuRequests *resource.Quantity
	totalCpuLimits   *resource.Quantity
	totalMemRequests *resource.Quantity
	totalMemLimits   *resource.Quantity

	totalStorage *resource.Quantity
}

const VERSION = "version: "

func TestCalculateResourceUsage(t *testing.T) {
	f, err := os.Create("resource_usage.txt")
	if err != nil {
		t.Fatal(err)
	}

	valuesChart, err := os.ReadFile("../stable/suse-observability-values/Chart.yaml")
	if err != nil {
		t.Fatal(err)
	}
	valuesVersion := ""
	for _, line := range strings.Split(string(valuesChart), "\n") {
		if len(line) > len(VERSION) && line[:len(VERSION)] == VERSION {
			valuesVersion = line[len(VERSION):]
		}
	}

	observabilityChart, err := os.ReadFile("../stable/suse-observability/Chart.yaml")
	if err != nil {
		t.Fatal(err)
	}
	observabilityVersion := ""
	for _, line := range strings.Split(string(observabilityChart), "\n") {
		if len(line) > len(VERSION) && line[:len(VERSION)] == VERSION {
			observabilityVersion = line[len(VERSION):]
		}
	}

	f.WriteString("\n")
	fmt.Fprintf(f, "suse-observability-values version: %s\n", valuesVersion)
	fmt.Fprintf(f, "suse-observability version: %s\n", observabilityVersion)
	for _, profile := range []string{
		"trial",
		"10-nonha",
		"20-nonha",
		"50-nonha",
		"100-nonha",
		"150-ha",
		"250-ha",
		"500-ha",
		"4000-ha",
	} {
		fmt.Fprintf(f, "\nProfile: %s\n", profile)
		usage := CalculateResourceUsage(t, profile)
		fmt.Fprintf(f, "  cpu request: %v\n", usage.totalCpuRequests)
		fmt.Fprintf(f, "  cpu limit: %v\n", usage.totalCpuLimits)
		fmt.Fprintf(f, "  mem request: %v\n", usage.totalMemRequests)
		fmt.Fprintf(f, "  mem limit: %v\n", usage.totalMemLimits)
		fmt.Fprintf(f, "  storage: %v\n", usage.totalStorage)
	}
	f.Close()
}

func CalculateResourceUsage(t *testing.T, profile string) resourceUsage {
	values := []string{"values/base.yaml"}

	output, err := helm.RenderTemplateE(t,
		&helm.Options{
			Logger:      logger.Discard,
			ValuesFiles: values,
			SetValues:   map[string]string{"sizing.profile": profile},
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

	result := resourceUsage{
		profile:          profile,
		totalCpuRequests: resource.NewMilliQuantity(0, resource.DecimalSI),
		totalCpuLimits:   resource.NewMilliQuantity(0, resource.DecimalSI),
		totalMemRequests: resource.NewQuantity(0, resource.DecimalSI),
		totalMemLimits:   resource.NewQuantity(0, resource.DecimalSI),
		totalStorage:     resource.NewQuantity(0, resource.DecimalSI),
	}

	deployments := resources.Deployments
	for _, deployment := range deployments {
		replicas := int64(1)
		if deployment.Spec.Replicas != nil {
			replicas = int64(*deployment.Spec.Replicas)
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			cpuReq := container.Resources.Requests.Cpu().DeepCopy()
			cpuReq.Mul(replicas)
			result.totalCpuRequests.Add(cpuReq)
			cpuLim := container.Resources.Limits.Cpu().DeepCopy()
			cpuLim.Mul(replicas)
			result.totalCpuLimits.Add(cpuLim)
			memReq := container.Resources.Requests.Memory().DeepCopy()
			memReq.Mul(replicas)
			result.totalMemRequests.Add(memReq)
			memLim := container.Resources.Limits.Memory().DeepCopy()
			memLim.Mul(replicas)
			result.totalMemLimits.Add(memLim)

			ephemeral := container.Resources.Requests.StorageEphemeral().DeepCopy()
			ephemeral.Mul(replicas)
			result.totalStorage.Add(ephemeral)
		}
		for _, vol := range deployment.Spec.Template.Spec.Volumes {
			if vol.Ephemeral != nil {
				vct := vol.Ephemeral.VolumeClaimTemplate
				storage := vct.Spec.Resources.Requests.Storage().DeepCopy()
				storage.Mul(replicas)
				result.totalStorage.Add(storage)
			}
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
			result.totalCpuRequests.Add(cpuReq)
			cpuLim := container.Resources.Limits.Cpu().DeepCopy()
			cpuLim.Mul(replicas)
			result.totalCpuLimits.Add(cpuLim)
			memReq := container.Resources.Requests.Memory().DeepCopy()
			memReq.Mul(replicas)
			result.totalMemRequests.Add(memReq)
			memLim := container.Resources.Limits.Memory().DeepCopy()
			memLim.Mul(replicas)
			result.totalMemLimits.Add(memLim)

			ephemeral := container.Resources.Requests.StorageEphemeral().DeepCopy()
			ephemeral.Mul(replicas)
			result.totalStorage.Add(ephemeral)
		}
		for _, vol := range statefulset.Spec.Template.Spec.Volumes {
			if vol.Ephemeral != nil {
				vct := vol.Ephemeral.VolumeClaimTemplate
				storage := vct.Spec.Resources.Requests.Storage().DeepCopy()
				storage.Mul(replicas)
				result.totalStorage.Add(storage)
			}
		}
		for _, vct := range statefulset.Spec.VolumeClaimTemplates {
			storage := vct.Spec.Resources.Requests.Storage().DeepCopy()
			storage.Mul(replicas)
			result.totalStorage.Add(storage)
		}
	}

	pvcs := resources.PersistentVolumeClaims
	for _, pvc := range pvcs {
		result.totalStorage.Add(pvc.Spec.Resources.Requests.Storage().DeepCopy())
	}

	cms := resources.ConfigMaps
	for _, configMap := range cms {
		_, found := configMap.ObjectMeta.Labels["stackstate.com/backup-scripts"]
		if found {
			for key, data := range configMap.Data {
				if key[len(key)-5:] == ".yaml" {
					backupResources := helmtestutil.NewKubernetesResources(t, data)
					pvcs := backupResources.PersistentVolumeClaims
					for _, pvc := range pvcs {
						result.totalStorage.Add(pvc.Spec.Resources.Requests.Storage().DeepCopy())
					}
				}
			}
		}
	}

	return result
}
