package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const maxSilenceTimeEnvName = "CONFIG_FORCE_stackstate_silencing_maxSilenceTime"

func findDeployment(t *testing.T, deployments map[string]appsv1.Deployment, name string) appsv1.Deployment {
	t.Helper()
	deployment, ok := deployments[name]
	if !ok {
		t.Fatalf("deployment %s not found", name)
	}
	return deployment
}

func envNames(env []v1.EnvVar) []string {
	names := make([]string, 0, len(env))
	for _, e := range env {
		names = append(names, e.Name)
	}
	return names
}

func TestSilencingMaxSilenceTimeDefaultsToTwentyFourHoursOnApi(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	api := findDeployment(t, resources.Deployments, "suse-observability-api")
	AssertEnvValue(t, api.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: maxSilenceTimeEnvName, Value: "24 hours"})
}

func TestSilencingMaxSilenceTimeDefaultsToTwentyFourHoursOnServer(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	server := findDeployment(t, resources.Deployments, "suse-observability-server")
	AssertEnvValue(t, server.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: maxSilenceTimeEnvName, Value: "24 hours"})
}

func TestSilencingMaxSilenceTimeOverrideOnApi(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/silencing_max_silence_time.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	api := findDeployment(t, resources.Deployments, "suse-observability-api")
	AssertEnvValue(t, api.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: maxSilenceTimeEnvName, Value: "2 hours"})
}

func TestSilencingMaxSilenceTimeOverrideOnServer(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml", "values/silencing_max_silence_time.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	server := findDeployment(t, resources.Deployments, "suse-observability-server")
	AssertEnvValue(t, server.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: maxSilenceTimeEnvName, Value: "2 hours"})
}

func TestSilencingMaxSilenceTimeOmittedWhenUnset(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/silencing_max_silence_time_unset.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	api := findDeployment(t, resources.Deployments, "suse-observability-api")
	assert.NotContains(t, envNames(api.Spec.Template.Spec.Containers[0].Env), maxSilenceTimeEnvName)
}

func TestSilencingMaxSilenceTimeIsNotSetOnOtherServices(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	for name, deployment := range resources.Deployments {
		if name == "suse-observability-api" {
			continue
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			assert.NotContains(t, envNames(container.Env), maxSilenceTimeEnvName,
				"%s should not carry the silencing maxSilenceTime override", name)
		}
	}
}
