package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestRbacResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definitions := string(data)
	names := map[string]string{}
	for _, resource := range []string{
		"deployment", "serviceaccount", "clusterrole", "clusterrolebinding",
		"api-key.secret", "pull.secret", "url.configmap", "clusterName.configmap", "customCertificates.configmap",
	} {
		header := `define "kubernetes-rbac-agent.` + resource + `.fullname"`
		require.Equal(t, 1, strings.Count(definitions, header))
		definitions = strings.Replace(definitions, header, `define "test.original.kubernetes-rbac-agent.`+resource+`.fullname"`, 1)
		names[resource] = "explicit-" + strings.ToLower(strings.ReplaceAll(resource, ".", "-"))
		definitions += "\n{{- " + header + " -}}" + names[resource] + "{{- end -}}\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))
	valuesFile, err := filepath.Abs("../linter_values.yaml")
	require.NoError(t, err)
	seen := map[string]bool{}
	for _, scenario := range []struct {
		name   string
		values map[string]string
	}{
		{"internal", map[string]string{
			"apiKey": "test-key", "global.customCertificates.enabled": "true",
			"global.customCertificates.pemData": "test-certificate",
		}},
		{"external-configmaps", map[string]string{
			"url.value": "", "clusterName.value": "",
			"url.fromConfigMap": "customer-url", "clusterName.fromConfigMap": "customer-cluster",
			"global.customCertificates.enabled": "true", "global.customCertificates.configMapName": "customer-certificates",
		}},
		{"external-secrets", map[string]string{
			"apiKey":                   "generated-but-unused",
			"global.apiKey.fromSecret": "\\{\\{ .Release.Name }}-customer-key",
			"global.url.fromSecret":    "customer-url", "global.clusterName.fromSecret": "customer-cluster",
		}},
		{"external-and-internal", map[string]string{
			"apiKey": "generated-but-unused", "global.apiKey.fromSecret": "customer-key",
			"url.fromConfigMap": "customer-url", "clusterName.fromConfigMap": "customer-cluster",
		}},
		{"platform-pull-secret", map[string]string{
			"global.suseObservability.pullSecret.username": "test",
			"global.suseObservability.pullSecret.password": "test-password",
		}},
	} {
		for _, namespace := range []string{"observability", "customer"} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/upgrade=%t", scenario.name, namespace, upgrade), func(t *testing.T) {
					options := &helm.Options{
						ValuesFiles: []string{valuesFile}, SetValues: scenario.values,
						KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
					}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.NewKubernetesResources(t,
						helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "orders", options, args...))
					output, err := helm.RenderTemplateE(t, options, chart, "orders", nil, args...)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)
					localName := "orders-rbac-agent"
					globalName := namespace + "-orders-rbac-agent"
					// Compare complete resources, changing only fixture identities,
					// references, and the existing certificate ConfigMap checksum.
					expected := before
					require.Contains(t, expected.Deployments, localName)
					originalDeployment := expected.Deployments[localName]
					deployment := originalDeployment.DeepCopy()
					delete(expected.Deployments, localName)
					deployment.Name = names["deployment"]
					deployment.Spec.Template.Name = names["deployment"]
					deployment.Spec.Template.Spec.ServiceAccountName = names["serviceaccount"]
					expected.Deployments[deployment.Name] = *deployment
					seen["deployment"] = true

					account := expected.ServiceAccounts[localName]
					delete(expected.ServiceAccounts, localName)
					account.Name = names["serviceaccount"]
					expected.ServiceAccounts[account.Name] = account
					seen["serviceaccount"] = true
					role := expected.ClusterRoles[globalName]
					delete(expected.ClusterRoles, globalName)
					role.Name = names["clusterrole"]
					expected.ClusterRoles[role.Name] = role
					seen["clusterrole"] = true
					require.Contains(t, expected.ClusterRoleBindings, globalName)
					originalBinding := expected.ClusterRoleBindings[globalName]
					binding := originalBinding.DeepCopy()
					delete(expected.ClusterRoleBindings, globalName)
					binding.Name = names["clusterrolebinding"]
					binding.RoleRef.Name = names["clusterrole"]
					require.Len(t, binding.Subjects, 1)
					binding.Subjects[0].Name = names["serviceaccount"]
					expected.ClusterRoleBindings[binding.Name] = *binding
					seen["clusterrolebinding"] = true

					secrets := map[string]string{
						localName + "-api-key":     names["api-key.secret"],
						localName + "-pull-secret": names["pull.secret"],
					}
					for old, name := range secrets {
						if secret, ok := expected.Secrets[old]; ok {
							delete(expected.Secrets, old)
							secret.Name = name
							expected.Secrets[name] = secret
							seen[name] = true
						}
					}
					configs := map[string]string{
						localName + "-url":                          names["url.configmap"],
						localName + "-cluster-name":                 names["clusterName.configmap"],
						"kubernetes-rbac-agent-custom-certificates": names["customCertificates.configmap"],
					}
					for old, name := range configs {
						if config, ok := expected.ConfigMaps[old]; ok {
							delete(expected.ConfigMaps, old)
							config.Name = name
							expected.ConfigMaps[name] = config
							seen[name] = true
						}
					}
					pod := &deployment.Spec.Template.Spec
					for i := range pod.ImagePullSecrets {
						if name, ok := secrets[pod.ImagePullSecrets[i].Name]; ok {
							pod.ImagePullSecrets[i].Name = name
						}
					}
					for i := range pod.Containers[0].Env {
						ref := pod.Containers[0].Env[i].ValueFrom
						if ref != nil && ref.SecretKeyRef != nil {
							if name, ok := secrets[ref.SecretKeyRef.Name]; ok {
								ref.SecretKeyRef.Name = name
							}
						}
					}
					for _, source := range pod.Containers[0].EnvFrom {
						if source.ConfigMapRef != nil {
							if name, ok := configs[source.ConfigMapRef.Name]; ok {
								source.ConfigMapRef.Name = name
							}
						}
					}
					for _, volume := range pod.Volumes {
						if volume.ConfigMap != nil {
							if name, ok := configs[volume.ConfigMap.Name]; ok {
								volume.ConfigMap.Name = name
							}
						}
					}
					const checksum = "checksum/custom-certificates"
					if hash, ok := deployment.Spec.Template.Annotations[checksum]; ok {
						require.Contains(t, after.Deployments, deployment.Name)
						actualHash := after.Deployments[deployment.Name].Spec.Template.Annotations[checksum]
						assert.NotEqual(t, hash, actualHash)
						deployment.Spec.Template.Annotations[checksum] = actualHash
					}
					expected.Deployments[deployment.Name] = *deployment
					assert.Equal(t, expected, after)
				})
			}
		}
	}
	for resource, name := range names {
		assert.True(t, seen[resource] || seen[name], "helper not exercised: %s", resource)
	}
}
