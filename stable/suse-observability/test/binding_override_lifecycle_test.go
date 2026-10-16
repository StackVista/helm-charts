package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// Model Helm's identity matching and three-way patches offline. A different
// binding name is a replacement; a retained name must never patch roleRef.
// This verifies the immutable-field contract, not API/controller reconciliation.
func TestBindingNamingOverrideUpgradeRollback(t *testing.T) {
	for _, split := range []bool{false, true} {
		for _, namespace := range []string{"suse-observability", "tenant-a"} {
			for _, affixes := range []bool{false, true} {
				t.Run(fmt.Sprintf("split=%t/namespace=%s/affixes=%t", split, namespace, affixes), func(t *testing.T) {
					var previous *helmtestutil.KubernetesResources
					var previousPodName, previousAuthorizationName string
					for revision, override := range []string{"first-platform", "second-platform", "second-platform", "first-platform"} {
						options := apiResourceNameTestOptions(map[string]string{
							"fullnameOverride":                 override,
							"stackstate.features.server.split": fmt.Sprint(split),
						})
						options.KubectlOptions.Namespace = namespace
						prefix := override
						if affixes {
							options.SetValues["fullnamePrefix"] = "pre-"
							options.SetValues["global.fullnameSuffix"] = "-end"
							prefix = "pre-" + override + "-end"
						}
						output := renderRBACNamingChart(t, "..", "nightly", options, revision > 0)
						current := helmtestutil.NewKubernetesResources(t, output)
						podName := prefix + "-get-pods"
						authorizationName := prefix + "-authorization"
						authenticationName := "suse-observability-authentication"
						if namespace != "suse-observability" {
							authenticationName = "suse-observability-" + namespace + "-authentication"
						}
						require.Contains(t, current.RoleBindings, podName)
						require.Contains(t, current.Roles, podName)
						assert.Equal(t, podName, current.RoleBindings[podName].RoleRef.Name)
						require.Contains(t, current.ClusterRoleBindings, authorizationName)
						require.Contains(t, current.ClusterRoles, authorizationName)
						assert.Equal(t, authorizationName, current.ClusterRoleBindings[authorizationName].RoleRef.Name)
						require.Contains(t, current.ClusterRoleBindings, authenticationName)
						assert.Equal(t, "system:auth-delegator", current.ClusterRoleBindings[authenticationName].RoleRef.Name)
						if previous != nil {
							// Upgrade, repeated upgrade and rollback must all keep
							// roleRef unchanged on every binding Helm would patch.
							for name, before := range previous.RoleBindings {
								if after, exists := current.RoleBindings[name]; exists {
									assertBindingPatchPreservesRoleRef(t, before, after, before.RoleRef, rbacv1.RoleBinding{})
								}
							}
							for name, before := range previous.ClusterRoleBindings {
								if after, exists := current.ClusterRoleBindings[name]; exists {
									assertBindingPatchPreservesRoleRef(t, before, after, before.RoleRef, rbacv1.ClusterRoleBinding{})
								}
							}
							if revision != 2 {
								assert.NotEqual(t, previous.ClusterRoleBindings[authenticationName].Subjects, current.ClusterRoleBindings[authenticationName].Subjects,
									"Canonical authentication binding must allow mutable subject updates")
								assert.NotContains(t, previous.RoleBindings, podName, "Override changes require a replacement Pod-reader binding")
								assert.NotContains(t, previous.ClusterRoleBindings, authorizationName, "Override changes require a replacement authorization binding")
								assert.NotContains(t, current.RoleBindings, previousPodName)
								assert.NotContains(t, current.ClusterRoleBindings, previousAuthorizationName)
							} else {
								assertPodReaderBindingUpgradePatchUnchanged(t, previous.RoleBindings[podName], current.RoleBindings[podName])
								assertClusterBindingUpgradePatchUnchanged(t, previous.ClusterRoleBindings[authorizationName], current.ClusterRoleBindings[authorizationName])
							}
						}
						previous = &current
						previousPodName, previousAuthorizationName = podName, authorizationName
					}
				})
			}
		}
	}
}

func assertBindingPatchPreservesRoleRef(t *testing.T, before, after interface{}, roleRef rbacv1.RoleRef, schema interface{}) {
	t.Helper()
	original, err := json.Marshal(before)
	require.NoError(t, err)
	modified, err := json.Marshal(after)
	require.NoError(t, err)
	var live map[string]interface{}
	require.NoError(t, json.Unmarshal(original, &live))
	metadata := live["metadata"].(map[string]interface{})
	metadata["resourceVersion"] = "fixture-version"
	metadata["annotations"] = map[string]interface{}{"fixture.customer": "keep"}
	liveJSON, err := json.Marshal(live)
	require.NoError(t, err)
	patchMetadata, err := strategicpatch.NewPatchMetaFromStruct(schema)
	require.NoError(t, err)
	patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, liveJSON, patchMetadata, true)
	require.NoError(t, err)
	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(patch, &fields))
	assert.NotContains(t, fields, "roleRef", "Helm must not patch immutable roleRef")
	result, err := strategicpatch.StrategicMergePatch(liveJSON, patch, schema)
	require.NoError(t, err)
	var patched struct {
		RoleRef  rbacv1.RoleRef   `json:"roleRef"`
		Subjects []rbacv1.Subject `json:"subjects"`
	}
	require.NoError(t, json.Unmarshal(result, &patched))
	assert.Equal(t, roleRef, patched.RoleRef)
	var desired struct {
		Subjects []rbacv1.Subject `json:"subjects"`
	}
	require.NoError(t, json.Unmarshal(modified, &desired))
	assert.Equal(t, desired.Subjects, patched.Subjects, "Mutable subject updates must still apply")
}
