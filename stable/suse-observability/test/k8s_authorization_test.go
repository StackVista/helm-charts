package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var expectedRoles = map[string]v1.Role{
	"suse-observability-instance-basic-access": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-basic-access",
		},
		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"views", "settings", "metricbindings", "systemnotifications"},
				Verbs:     []string{"get"},
			},
		},
	},
	"suse-observability-instance-troubleshooter": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-troubleshooter",
		},
		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"views", "monitors", "notifications", "stackpackconfigurations", "dashboards"},
				Verbs:     []string{"create", "update", "get", "delete"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"agents", "apitokens", "metrics", "metricbindings", "settings", "stackpacks", "systemnotifications", "topology", "traces"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"visualizationsettings"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"componentactions", "monitors"},
				Verbs:     []string{"execute"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"favoriteviews", "favoritedashboards"},
				Verbs:     []string{"delete", "create"},
			},
		},
	},
	"suse-observability-instance-recommended-access": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-recommended-access",
		},
		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"apitokens", "stackpacks"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"visualizationsettings"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"favoriteviews", "favoritedashboards"},
				Verbs:     []string{"delete", "create"},
			},
		},
	},
	"suse-observability-instance-admin-saas": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-admin",
		},
		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"views", "metricbindings", "monitors", "notifications", "permissions", "servicetokens", "stackpackconfigurations", "dashboards"},
				Verbs:     []string{"create", "update", "get", "delete"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"agents", "apitokens", "metrics", "syncdata", "systemnotifications", "topology", "traces", "stackpacks"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"syncdata", "visualizationsettings"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"componentactions", "monitors"},
				Verbs:     []string{"execute"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"favoriteviews", "syncdata", "favoritedashboards"},
				Verbs:     []string{"delete"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"favoriteviews", "favoritedashboards"},
				Verbs:     []string{"create"},
			},
		},
	},
	"suse-observability-instance-admin-selfhosted": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-admin",
		},
		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"views", "permissions", "servicetokens", "settings", "stackpackconfigurations", "monitors", "notifications", "dashboards"},
				Verbs:     []string{"create", "update", "get", "delete"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"settings"},
				Verbs:     []string{"unlock"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"stackpacks"},
				Verbs:     []string{"create", "get"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"agents", "apitokens", "metrics", "metricbindings", "syncdata", "systemnotifications", "topology", "topicmessages", "traces"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"syncdata", "visualizationsettings"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"componentactions", "monitors", "restrictedscripts"},
				Verbs:     []string{"execute"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"favoriteviews", "favoritedashboards"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"favoriteviews", "syncdata", "favoritedashboards"},
				Verbs:     []string{"delete"},
			},
		},
	},
	"suse-observability-instance-observer": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-observer",
		},
		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"instance.observability.cattle.io"},
				Resources: []string{"metrics", "topology", "traces"},
				Verbs:     []string{"get"},
			},
		},
	},
}

var expectedRoleBindings = map[string]v1.RoleBinding{
	"suse-observability-instance-basic-access": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-basic-access",
		},
		Subjects: []v1.Subject{
			{
				Kind: "Group",
				Name: "system:authenticated",
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-instance-basic-access",
		},
	},
	"suse-observability-instance-troubleshooter": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-troubleshooter",
		},
		Subjects: []v1.Subject{
			{
				Kind: "Group",
				Name: "suse-observability-instance-troubleshooter",
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-instance-troubleshooter",
		},
	},
	"suse-observability-instance-admin": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-admin",
		},
		Subjects: []v1.Subject{
			{
				Kind: "Group",
				Name: "suse-observability-instance-admin",
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-instance-admin",
		},
	},
	"suse-observability-instance-recommended-access": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-recommended-access",
		},
		Subjects: []v1.Subject{
			{
				Kind: "Group",
				Name: "system:authenticated",
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-instance-recommended-access",
		},
	},
	"suse-observability-instance-observer": {
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "suse-observability-instance-observer",
		},
		Subjects: []v1.Subject{
			{
				Kind: "Group",
				Name: "suse-observability-instance-observer",
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "suse-observability-instance-observer",
		},
	},
}

func TestK8sAuthzDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 11, len(resources.Roles), "Default configuration should generate exactly 11 roles")
	ok := assert.Contains(t, resources.Roles, "suse-observability-instance-basic-access")
	if ok {
		checkRole(t, expectedRoles["suse-observability-instance-basic-access"], resources.Roles["suse-observability-instance-basic-access"])
	}
	assert.Contains(t, resources.Roles, "suse-observability-instance-observer", "Default config should include observer role")
	assert.Contains(t, resources.Roles, "suse-observability-instance-troubleshooter", "Default config should include troubleshooter role")
	assert.Contains(t, resources.Roles, "suse-observability-instance-admin", "Default config should include admin role")
	assert.Contains(t, resources.Roles, "suse-observability-instance-recommended-access", "Default config should include recommended-access role")

	require.Equal(t, 11, len(resources.RoleBindings), "Default configuration should generate exactly 11 role bindings")
	ok = assert.Contains(t, resources.RoleBindings, "suse-observability-instance-basic-access")
	if ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-basic-access"], resources.RoleBindings["suse-observability-instance-basic-access"])
	}
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-observer", "Default config should include observer role binding")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-troubleshooter", "Default config should include troubleshooter role binding")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-admin", "Default config should include admin role binding")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-recommended-access", "Default config should include recommended-access role binding")
}

func TestK8sAuthzSaas(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.deployment.mode": "Saas",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 11, len(resources.Roles), "SaaS mode with k8s-authz feature enabled should generate 11 roles")
	if ok := assert.Contains(t, resources.Roles, "suse-observability-instance-basic-access"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-basic-access"], resources.Roles["suse-observability-instance-basic-access"])
	}
	checkFeatureSaasRoles(t, resources.Roles)

	require.Equal(t, 11, len(resources.RoleBindings), "SaaS mode with k8s-authz feature enabled should generate 11 role bindings")
	if ok := assert.Contains(t, resources.RoleBindings, "suse-observability-instance-basic-access"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-basic-access"], resources.RoleBindings["suse-observability-instance-basic-access"])
	}
	checkFeatureSaasRoleBindings(t, resources.RoleBindings)

}

func TestK8sAuthzFeatureFlagDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.role-k8s-authz": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 7, len(resources.Roles), "Default configuration should generate exactly 7 roles")
	ok := assert.Contains(t, resources.Roles, "suse-observability-instance-basic-access")
	if ok {
		checkRole(t, expectedRoles["suse-observability-instance-basic-access"], resources.Roles["suse-observability-instance-basic-access"])
	}
	assert.NotContains(t, resources.Roles, "suse-observability-instance-observer", "Default config should not include observer role")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-troubleshooter", "Default config should not include troubleshooter role")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-admin", "Default config should not include admin role")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-recommended-access", "Default config should not include recommended-access role")

	require.Equal(t, 7, len(resources.RoleBindings), "Default configuration should generate exactly 7 role bindings")
	ok = assert.Contains(t, resources.RoleBindings, "suse-observability-instance-basic-access")
	if ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-basic-access"], resources.RoleBindings["suse-observability-instance-basic-access"])
	}
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-observer", "Default config should not include observer role binding")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-troubleshooter", "Default config should not include troubleshooter role binding")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-admin", "Default config should not include admin role binding")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-recommended-access", "Default config should not include recommended-access role binding")
}

func TestK8sAuthzExperimentalOverFeaturesDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.role-k8s-authz":     "true",
			"stackstate.experimental.role-k8s-authz": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 6, len(resources.Roles), "Experimental override (false) should disable k8s-authz feature, generating only 6 roles")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-observer", "Experimental override should disable observer role")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-troubleshooter", "Experimental override should disable troubleshooter role")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-admin", "Experimental override should disable admin role")
	assert.NotContains(t, resources.Roles, "suse-observability-instance-recommended-access", "Experimental override should disable recommended-access role")

	require.Equal(t, 6, len(resources.RoleBindings), "Experimental override (false) should disable k8s-authz feature, generating only 6 role bindings")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-observer", "Experimental override should disable observer role binding")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-troubleshooter", "Experimental override should disable troubleshooter role binding")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-admin", "Experimental override should disable admin role binding")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-instance-recommended-access", "Experimental override should disable recommended-access role binding")

}

func TestK8sAuthzExperimentalOverFeaturesEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.role-k8s-authz":     "false",
			"stackstate.experimental.role-k8s-authz": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 10, len(resources.Roles), "Experimental override (true) should enable k8s-authz feature, generating 10 roles")
	assert.Contains(t, resources.Roles, "suse-observability-instance-observer", "Experimental override should enable observer role")
	assert.Contains(t, resources.Roles, "suse-observability-instance-troubleshooter", "Experimental override should enable troubleshooter role")
	assert.Contains(t, resources.Roles, "suse-observability-instance-admin", "Experimental override should enable admin role")
	assert.Contains(t, resources.Roles, "suse-observability-instance-recommended-access", "Experimental override should enable recommended-access role")

	require.Equal(t, 10, len(resources.RoleBindings), "Experimental override (true) should enable k8s-authz feature, generating 10 role bindings")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-observer", "Experimental override should enable observer role binding")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-troubleshooter", "Experimental override should enable troubleshooter role binding")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-admin", "Experimental override should enable admin role binding")
	assert.Contains(t, resources.RoleBindings, "suse-observability-instance-recommended-access", "Experimental override should enable recommended-access role binding")

}

func TestK8sAuthzFeatureFlagEnabledSelfHosted(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.role-k8s-authz": "true",
			"stackstate.deployment.mode":         "SelfHosted",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 10, len(resources.Roles), "SelfHosted mode with k8s-authz feature enabled should generate 10 roles")
	if ok := assert.Contains(t, resources.Roles, "suse-observability-instance-basic-access"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-basic-access"], resources.Roles["suse-observability-instance-basic-access"])
	}

	checkFeatureSelfHostedRoles(t, resources.Roles)

	require.Equal(t, 10, len(resources.RoleBindings), "SelfHosted mode with k8s-authz feature enabled should generate 10 role bindings")
	if ok := assert.Contains(t, resources.RoleBindings, "suse-observability-instance-basic-access"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-basic-access"], resources.RoleBindings["suse-observability-instance-basic-access"])
	}
	checkFeatureSelfHostedRoleBindings(t, resources.RoleBindings)
}

func checkRole(t *testing.T, expected, got v1.Role) {
	assert.Equal(t, expected.Name, got.Name, "Role name should match expected")
	assert.Equal(t, expected.APIVersion, got.APIVersion, "Role API version should match expected")
	assert.Equal(t, expected.Kind, got.Kind, "Role kind should match expected")
	assert.Equal(t, expected.Rules, got.Rules, "Role rules should match expected")
}

func checkRoleBinding(t *testing.T, expected, got v1.RoleBinding) {
	assert.Equal(t, expected.Name, got.Name, "RoleBinding name should match expected")
	assert.Equal(t, expected.APIVersion, got.APIVersion, "RoleBinding API version should match expected")
	assert.Equal(t, expected.Kind, got.Kind, "RoleBinding kind should match expected")
	assert.Equal(t, expected.RoleRef, got.RoleRef, "RoleBinding role reference should match expected")
	assert.Equal(t, expected.Subjects, got.Subjects, "RoleBinding subjects should match expected")
}

func checkFeatureSaasRoles(t *testing.T, roles map[string]v1.Role) {
	if ok := assert.Contains(t, roles, "suse-observability-instance-troubleshooter"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-troubleshooter"], roles["suse-observability-instance-troubleshooter"])
	}
	if ok := assert.Contains(t, roles, "suse-observability-instance-recommended-access"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-recommended-access"], roles["suse-observability-instance-recommended-access"])
	}
	if ok := assert.Contains(t, roles, "suse-observability-instance-admin"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-admin-saas"], roles["suse-observability-instance-admin"])
	}
	if ok := assert.Contains(t, roles, "suse-observability-instance-observer"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-observer"], roles["suse-observability-instance-observer"])
	}
}

func checkFeatureSaasRoleBindings(t *testing.T, roleBindings map[string]v1.RoleBinding) {
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-troubleshooter"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-troubleshooter"], roleBindings["suse-observability-instance-troubleshooter"])
	}
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-admin"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-admin"], roleBindings["suse-observability-instance-admin"])
	}
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-recommended-access"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-recommended-access"], roleBindings["suse-observability-instance-recommended-access"])
	}
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-observer"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-observer"], roleBindings["suse-observability-instance-observer"])
	}
}

func checkFeatureSelfHostedRoles(t *testing.T, roles map[string]v1.Role) {
	if ok := assert.Contains(t, roles, "suse-observability-instance-troubleshooter"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-troubleshooter"], roles["suse-observability-instance-troubleshooter"])
	}
	if ok := assert.Contains(t, roles, "suse-observability-instance-recommended-access"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-recommended-access"], roles["suse-observability-instance-recommended-access"])
	}
	if ok := assert.Contains(t, roles, "suse-observability-instance-admin"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-admin-selfhosted"], roles["suse-observability-instance-admin"])
	}
	if ok := assert.Contains(t, roles, "suse-observability-instance-observer"); ok {
		checkRole(t, expectedRoles["suse-observability-instance-observer"], roles["suse-observability-instance-observer"])
	}
}

func checkFeatureSelfHostedRoleBindings(t *testing.T, roleBindings map[string]v1.RoleBinding) {
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-troubleshooter"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-troubleshooter"], roleBindings["suse-observability-instance-troubleshooter"])
	}
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-admin"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-admin"], roleBindings["suse-observability-instance-admin"])
	}
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-recommended-access"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-recommended-access"], roleBindings["suse-observability-instance-recommended-access"])
	}
	if ok := assert.Contains(t, roleBindings, "suse-observability-instance-observer"); ok {
		checkRoleBinding(t, expectedRoleBindings["suse-observability-instance-observer"], roleBindings["suse-observability-instance-observer"])
	}
}
