package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestRbacAgentServiceAccountReferencesFollowDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_rbac-resource-names.tpl"), []byte(
		`{{- define "kubernetes-rbac-agent.serviceaccount.fullname" -}}explicit-rbac-account{{- end -}}`), 0600))
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, release := range []string{"suse-observability", "nightly"} {
		for _, split := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/split=%t/upgrade=%t", release, split, upgrade), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"stackstate.k8sAuthorization.enabled": "true",
						"stackstate.features.server.split":    fmt.Sprint(split),
					})
					options.ValuesFiles = []string{values}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, release, options, args...)
					after, err := helm.RenderTemplateE(t, options, chart, release, nil, args...)
					require.NoError(t, err)
					account := release + "-rbac-agent"
					seen := map[string]bool{}
					expected := backupNamingDocuments(t, before, map[string]string{
						"ServiceAccount/" + account: "explicit-rbac-account",
					}, seen)
					require.True(t, seen["ServiceAccount/"+account])
					rewriteRbacAgentAccountReferences(expected, account)
					// Includes both RoleBinding types and unchanged application
					// authorization subjects in API/server ConfigMaps.
					assert.Equal(t, expected, backupNamingDocuments(t, after, nil, nil))
				})
			}
		}
	}
}

func rewriteRbacAgentAccountReferences(value interface{}, old string) {
	switch node := value.(type) {
	case map[string]interface{}:
		if node["serviceAccountName"] == old {
			node["serviceAccountName"] = "explicit-rbac-account"
		}
		if subjects, ok := node["subjects"].([]interface{}); ok {
			for _, subject := range subjects {
				entry := subject.(map[string]interface{})
				if entry["kind"] == "ServiceAccount" && entry["name"] == old {
					entry["name"] = "explicit-rbac-account"
				}
			}
		}
		for _, child := range node {
			rewriteRbacAgentAccountReferences(child, old)
		}
	case []interface{}:
		for _, child := range node {
			rewriteRbacAgentAccountReferences(child, old)
		}
	}
}
