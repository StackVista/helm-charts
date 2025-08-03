package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	expectedClickhouseConfig = "stackstate.traces.clickHouse ="
	monolithComponent        = "suse-observability-server"
	serverServiceName        = "suse-observability-server-headless"
)

var expectedRoleBindingSplitEnabled = []rbacv1.RoleBinding{
	{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-get-pods",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-api",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-authorization-sync",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-checks",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-initializer",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-slicing",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-state",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-sync",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-notification",
			},
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-health-sync",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-get-pods",
		},
	},
}

var expectedRoleBindingSplitDisabled = []rbacv1.RoleBinding{
	{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-get-pods",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "suse-observability-server",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-get-pods",
		},
	},
}

var expectedClusterRoleBindingsSplitEnabled = []rbacv1.ClusterRoleBinding{
	{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-authentication",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "suse-observability-api",
				Namespace: "suse-observability",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "system:auth-delegator",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
	{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-authorization",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "suse-observability-api",
				Namespace: "suse-observability",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "suse-observability-authorization",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
}

var expectedClusterRoleBindingsSplitDisabled = []rbacv1.ClusterRoleBinding{
	{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-authentication",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "suse-observability-server",
				Namespace: "suse-observability",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "system:auth-delegator",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
	{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-authorization",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "suse-observability-server",
				Namespace: "suse-observability",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "suse-observability-authorization",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
}

var splitComponents = []string{
	"suse-observability-api",
	"suse-observability-checks",
	"suse-observability-authorization-sync",
	"suse-observability-health-sync",
	"suse-observability-initializer",
	"suse-observability-notification",
	"suse-observability-slicing",
	"suse-observability-state",
	"suse-observability-sync",
}

var splitComponentServices = []string{
	"suse-observability-api-headless",
	"suse-observability-checks",
	"suse-observability-authorization-sync",
	"suse-observability-health-sync",
	"suse-observability-initializer",
	"suse-observability-notification",
	"suse-observability-slicing",
	"suse-observability-state",
	"suse-observability-sync",
}

var splitComponentPdbs = []string{
	"suse-observability-api",
	"suse-observability-checks",
	"suse-observability-authorization-sync",
	"suse-observability-health-sync",
	"suse-observability-notification",
	"suse-observability-state",
	"suse-observability-sync",
}

func TestFeaturesTracesDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	expected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "API deployment should have traces feature flag enabled by default")

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	assert.Contains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig, "API configmap should contain traces ClickHouse configuration by default")
}

func TestFeaturesTracesDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.traces": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	notExpected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
	assert.NotContains(t, deployment.Spec.Template.Spec.Containers[0].Env, notExpected, "API deployment should not have traces feature flag when disabled")

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	assert.NotContains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig, "API configmap should not contain traces ClickHouse configuration when traces disabled")
}

func TestFeaturesTracesExperimentalOverFeatures(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.traces":     "true",
			"stackstate.experimental.traces": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	notExpected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
	assert.NotContains(t, deployment.Spec.Template.Spec.Containers[0].Env, notExpected, "API deployment should not have traces feature flag when experimental override disables it")

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	assert.NotContains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig, "API configmap should not contain traces ClickHouse configuration when experimental override disables traces")
}

func TestFeaturesDashboardsDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	notExpected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_dashboards", Value: "true"}
	assert.NotContains(t, deployment.Spec.Template.Spec.Containers[0].Env, notExpected, "API deployment should not have dashboards feature flag enabled by default")
}

func TestFeaturesDashboardsDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.dashboards": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	expected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_dashboards", Value: "true"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "API deployment should have dashboards feature flag when enabled")
}

func TestFeaturesDashboardsExperimentalOverFeatures(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.dashboards":     "false",
			"stackstate.experimental.dashboards": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	expected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_dashboards", Value: "true"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "API deployment should have dashboards feature flag when experimental override enables it")
}

func TestFeaturesServerSplitDefault(t *testing.T) {
	setValues := map[string]string{
		"stackstate.k8sAuthorization.enabled":                      "true",
		"stackstate.components.all.metrics.servicemonitor.enabled": "true",
	}
	resources := renderHelmTemplateForServerSplitTest(t, setValues)

	assertMonolithResourcesExist(t, resources, false)
	assertSplitComponentResourcesExist(t, resources, true)
	assertRoleBindings(t, resources, true)
	assertReceiverDeploymentSplit(t, resources, true)
	assertBackupCronjobSplit(t, resources, true)
}

func TestFeaturesServerSplitEnabled(t *testing.T) {
	setValues := map[string]string{
		"stackstate.features.server.split":                         "true",
		"stackstate.k8sAuthorization.enabled":                      "true",
		"stackstate.components.all.metrics.servicemonitor.enabled": "true",
	}
	resources := renderHelmTemplateForServerSplitTest(t, setValues)

	assertMonolithResourcesExist(t, resources, false)
	assertSplitComponentResourcesExist(t, resources, true)
	assertRoleBindings(t, resources, true)
	assertReceiverDeploymentSplit(t, resources, true)
	assertBackupCronjobSplit(t, resources, true)
}

func TestFeaturesServerSplitDisabled(t *testing.T) {
	setValues := map[string]string{
		"stackstate.features.server.split":                         "false",
		"stackstate.k8sAuthorization.enabled":                      "true",
		"stackstate.components.all.metrics.servicemonitor.enabled": "true",
	}
	resources := renderHelmTemplateForServerSplitTest(t, setValues)

	assertMonolithResourcesExist(t, resources, true)
	assertSplitComponentResourcesExist(t, resources, false)
	assertRoleBindings(t, resources, false)
	assertReceiverDeploymentSplit(t, resources, false)
	assertBackupCronjobSplit(t, resources, false)
}

func TestFeaturesServerSplitExperimentalOverFeatures(t *testing.T) {
	setValues := map[string]string{
		"stackstate.features.server.split":                         "false",
		"stackstate.experimental.server.split":                     "true",
		"stackstate.k8sAuthorization.enabled":                      "true",
		"stackstate.components.all.metrics.servicemonitor.enabled": "true",
	}
	resources := renderHelmTemplateForServerSplitTest(t, setValues)

	assertMonolithResourcesExist(t, resources, false)
	assertSplitComponentResourcesExist(t, resources, true)
	assertRoleBindings(t, resources, true)
	assertReceiverDeploymentSplit(t, resources, true)
	assertBackupCronjobSplit(t, resources, true)
}

func renderHelmTemplateForServerSplitTest(t *testing.T, setValues map[string]string) *helmtestutil.KubernetesResources {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: setValues,
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	return &resources
}

func assertMonolithResourcesExist(t *testing.T, resources *helmtestutil.KubernetesResources, shouldExist bool) {
	_, ok := resources.Deployments[monolithComponent]
	assert.Equal(t, shouldExist, ok, "%s deployment existence", monolithComponent)

	_, ok = resources.ConfigMaps[monolithComponent]
	assert.Equal(t, shouldExist, ok, "%s configMap existence", monolithComponent)

	logConfigMapName := monolithComponent + "-log"
	_, ok = resources.ConfigMaps[logConfigMapName]
	assert.Equal(t, shouldExist, ok, "%s configMap existence", logConfigMapName)

	_, ok = resources.Pdbs[monolithComponent]
	assert.Equal(t, shouldExist, ok, "%s podDisruptionBudget existence", monolithComponent)

	_, ok = resources.Services[serverServiceName]
	assert.Equal(t, shouldExist, ok, "%s Service existence", serverServiceName)

	_, ok = resources.ServiceAccounts[monolithComponent]
	assert.Equal(t, shouldExist, ok, "%s ServiceAccount existence", monolithComponent)

	_, ok = resources.Secrets[monolithComponent]
	assert.Equal(t, shouldExist, ok, "%s Secret existence", monolithComponent)

	_, ok = resources.ServiceMonitors[monolithComponent]
	assert.Equal(t, shouldExist, ok, "%s ServiceMonitor existence", monolithComponent)
}

func assertSplitComponentResourcesExist(t *testing.T, resources *helmtestutil.KubernetesResources, shouldExist bool) {
	for _, component := range splitComponents {
		_, ok := resources.Deployments[component]
		assert.Equal(t, shouldExist, ok, "%s deployment existence", component)

		_, ok = resources.ConfigMaps[component]
		assert.Equal(t, shouldExist, ok, "%s configMap existence", component)

		logConfigMapName := component + "-log"
		_, ok = resources.ConfigMaps[logConfigMapName]
		assert.Equal(t, shouldExist, ok, "%s configMap existence", logConfigMapName)

		_, ok = resources.ServiceAccounts[component]
		assert.Equal(t, shouldExist, ok, "%s ServiceAccount existence", component)

		_, ok = resources.Secrets[component]
		assert.Equal(t, shouldExist, ok, "%s Secret existence", component)

		_, ok = resources.ServiceMonitors[component]
		assert.Equal(t, shouldExist, ok, "%s ServiceMonitor existence", component)
	}

	for _, component := range splitComponentPdbs {
		_, ok := resources.Pdbs[component]
		assert.Equal(t, shouldExist, ok, "%s podDisruptionBudget existence", component)
	}

	for _, component := range splitComponentServices {
		_, ok := resources.Services[component]
		assert.Equal(t, shouldExist, ok, "%s Service existence", component)
	}
}

func assertRoleBindings(t *testing.T, resources *helmtestutil.KubernetesResources, splitEnabled bool) {
	var expectedClusterRoleBindings []rbacv1.ClusterRoleBinding
	var expectedRoleBindings []rbacv1.RoleBinding

	if splitEnabled {
		expectedClusterRoleBindings = expectedClusterRoleBindingsSplitEnabled
		expectedRoleBindings = expectedRoleBindingSplitEnabled
	} else {
		expectedClusterRoleBindings = expectedClusterRoleBindingsSplitDisabled
		expectedRoleBindings = expectedRoleBindingSplitDisabled
	}

	for _, crb := range expectedClusterRoleBindings {
		if ok := assert.Contains(t, resources.ClusterRoleBindings, crb.Name, "%s clusterRoleBidnging should exist", crb.Name); ok {
			checkClusterRoleBinding(t, crb, resources.ClusterRoleBindings[crb.Name])
		}
	}

	for _, rb := range expectedRoleBindings {
		if ok := assert.Contains(t, resources.RoleBindings, rb.Name, "%s roleBidnging should exist", rb.Name); ok {
			checkRoleBinding(t, rb, resources.RoleBindings[rb.Name])
		}
	}
}

func checkClusterRoleBinding(t *testing.T, expected, got rbacv1.ClusterRoleBinding) {
	assert.Equal(t, expected.Name, got.Name, "ClusterRoleBinding name should match expected")
	assert.Equal(t, expected.APIVersion, got.APIVersion, "ClusterRoleBinding API version should match expected")
	assert.Equal(t, expected.Kind, got.Kind, "ClusterRoleBinding kind should match expected")
	assert.Equal(t, expected.RoleRef, got.RoleRef, "ClusterRoleBinding role reference should match expected")
	assert.Equal(t, expected.Subjects, got.Subjects, "ClusterRoleBinding subjects should match expected")
}

func assertReceiverDeploymentSplit(t *testing.T, resources *helmtestutil.KubernetesResources, splitEnabled bool) {
	if ok := assert.Contains(t, resources.Deployments, "suse-observability-receiver", "Receiver deployment should exist"); ok {
		deployment := resources.Deployments["suse-observability-receiver"]
		var serviceBaseUriPattern, authzBaseUriPattern string
		if splitEnabled {
			serviceBaseUriPattern = "http://suse-observability-api-headless:7070/internal/api/agents"
			authzBaseUriPattern = "http://suse-observability-authorization-sync:7075/rbac"

		} else {
			serviceBaseUriPattern = "http://suse-observability-server-headless:7070/internal/api/agents"
			authzBaseUriPattern = "http://suse-observability-server-headless:7075/rbac"
		}
		expected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_receiver_agentLeases_agentServiceBaseUri", Value: serviceBaseUriPattern}
		assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "Receiver is configured with the correct api service")

		expected = corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_receiver_authorizationSyncApi_authorizationSyncApiBaseUri", Value: authzBaseUriPattern}
		assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "Receiver is configured with the correct authz service")
	}
}

func assertBackupCronjobSplit(t *testing.T, resources *helmtestutil.KubernetesResources, splitEnabled bool) {
	cronjobs := []string{"suse-observability-backup-sg", "suse-observability-backup-conf"}
	for _, cronjob := range cronjobs {
		if ok := assert.Contains(t, resources.CronJobs, cronjob, "CronJob%s  should exist, cronjob"); ok {
			cronjob := resources.CronJobs[cronjob]
			var entryPointPattern string
			if splitEnabled {
				entryPointPattern = "suse-observability-initializer:1618"

			} else {
				entryPointPattern = "suse-observability-server-headless:7070"
			}

			command := strings.Join(cronjob.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Command, "\n")
			assert.Contains(t, command, entryPointPattern, "Cronjob %s initContainer is configured with the correct entry point", cronjob)
		}
	}
}
