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

func TestKafkaupOperatorResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definitions := string(data)
	for _, resource := range []string{"deployment", "serviceaccount", "role", "rolebinding", "configmap"} {
		header := `define "kafkaup-operator.` + resource + `.fullname"`
		require.Equal(t, 1, strings.Count(definitions, header))
		definitions = strings.Replace(definitions, header, `define "test.original.kafkaup-operator.`+resource+`.fullname"`, 1)
		definitions += "\n{{- " + header + " -}}explicit-" + resource + "{{- end -}}\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))

	for _, release := range []string{"kafkaup-operator", "nightly"} {
		for _, namespace := range []string{"observability", "tenant-a"} {
			for _, scenario := range []struct {
				name string
				set  map[string]string
			}{
				{"default", nil},
				{"fullname-override", map[string]string{"fullnameOverride": "customer-operator"}},
				{"prefix-and-suffix", map[string]string{
					"global.fullnamePrefix": "global-", "fullnamePrefix": "local-",
					"fullnameSuffix": "-local", "global.fullnameSuffix": "-global",
				}},
				{"long-name", map[string]string{"fullnameOverride": strings.Repeat("o", 70)}},
				{"external-kafka-and-registry", map[string]string{
					"kafkaSelectors.statefulSetName": "customer-{{ .Release.Name }}-kafka",
					"kafkaSelectors.podLabel.key":    "customer-label", "kafkaSelectors.podLabel.value": "external-kafka",
					"image.pullSecretUsername": "registry-user", "image.pullSecretPassword": "registry-password",
				}},
			} {
				for _, upgrade := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/upgrade=%t", release, namespace, scenario.name, upgrade), func(t *testing.T) {
						options := &helm.Options{SetValues: scenario.set, KubectlOptions: &k8s.KubectlOptions{Namespace: namespace}}
						var args []string
						if upgrade {
							args = append(args, "--is-upgrade")
						}
						before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, release, options, args...))
						output, err := helm.RenderTemplateE(t, options, chart, release, nil, args...)
						require.NoError(t, err)
						after := helmtestutil.NewKubernetesResources(t, output)
						require.Len(t, before.Deployments, 1)
						var old string
						for name := range before.Deployments {
							old = name
						}
						deployment := before.Deployments[old]
						require.Equal(t, old, deployment.Spec.Template.Spec.ServiceAccountName)
						require.Contains(t, before.ServiceAccounts, old)
						require.Contains(t, before.Roles, old)
						require.Contains(t, before.RoleBindings, old+"-binding")
						require.Contains(t, before.ConfigMaps, old+"-config")
						deployment.Name = "explicit-deployment"
						deployment.Spec.Template.Spec.ServiceAccountName = "explicit-serviceaccount"
						for i, volume := range deployment.Spec.Template.Spec.Volumes {
							if volume.ConfigMap != nil && volume.ConfigMap.Name == old+"-config" {
								deployment.Spec.Template.Spec.Volumes[i].ConfigMap.Name = "explicit-configmap"
							}
						}
						require.Contains(t, after.Deployments, deployment.Name)
						actualChecksum := after.Deployments[deployment.Name].Spec.Template.Annotations["checksum/configmap"]
						assert.NotEqual(t, deployment.Spec.Template.Annotations["checksum/configmap"], actualChecksum)
						deployment.Spec.Template.Annotations["checksum/configmap"] = actualChecksum
						delete(before.Deployments, old)
						before.Deployments[deployment.Name] = deployment

						config := before.ConfigMaps[old+"-config"]
						config.Name = "explicit-configmap"
						delete(before.ConfigMaps, old+"-config")
						before.ConfigMaps[config.Name] = config
						account := before.ServiceAccounts[old]
						account.Name = "explicit-serviceaccount"
						delete(before.ServiceAccounts, old)
						before.ServiceAccounts[account.Name] = account
						role := before.Roles[old]
						role.Name = "explicit-role"
						delete(before.Roles, old)
						before.Roles[role.Name] = role
						binding := before.RoleBindings[old+"-binding"]
						binding.Name = "explicit-rolebinding"
						binding.RoleRef.Name = role.Name
						binding.Subjects[0].Name = account.Name
						delete(before.RoleBindings, old+"-binding")
						before.RoleBindings[binding.Name] = binding
						assert.Equal(t, before, after, "Only dedicated identities/references and the existing ConfigMap checksum may differ")
					})
				}
			}
		}
	}
}
