package test

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestLicenseEmailSecretReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	for resource, name := range map[string]string{"license": "explicit-license-secret", "email": "explicit-email-secret"} {
		helper := "stackstate." + resource + ".secret.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+name+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, split := range []bool{false, true} {
		for _, tc := range []struct {
			name, licenseExternal, emailExternal string
			emailEnabled                         bool
		}{
			{name: "internal", emailEnabled: true},
			{name: "email-disabled"},
			{name: "external-license", licenseExternal: "customer-license", emailEnabled: true},
			{name: "external-email", emailExternal: "customer-smtp", emailEnabled: true},
			{name: "both-external", licenseExternal: "customer-license", emailExternal: "customer-smtp", emailEnabled: true},
			{name: "external-email-disabled", emailExternal: "customer-smtp"},
		} {
			t.Run(fmt.Sprintf("%s/split=%t", tc.name, split), func(t *testing.T) {
				options := apiResourceNameTestOptions(map[string]string{
					"stackstate.features.server.split":                fmt.Sprint(split),
					"stackstate.license.key":                          "fixture-license",
					"stackstate.license.fromExternalSecret":           tc.licenseExternal,
					"stackstate.email.enabled":                        fmt.Sprint(tc.emailEnabled),
					"stackstate.email.server.auth.fromExternalSecret": tc.emailExternal,
					"stackstate.email.server.auth.username":           "fixture-user",
					"stackstate.email.server.auth.password":           "fixture-password",
					"backup.configuration.enabled":                    "true",
				})
				before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
				options.ValuesFiles = []string{valuesFile}
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
				require.NoError(t, err)
				after := helmtestutil.NewKubernetesResources(t, output)

				renames := map[string]string{}
				checksums := map[string]bool{}
				expectedLicense := tc.licenseExternal
				expectedEmail := tc.emailExternal
				for _, secret := range []struct {
					suffix, external string
					enabled          bool
				}{
					{"license", tc.licenseExternal, true},
					{"email", tc.emailExternal, tc.emailEnabled},
				} {
					legacy := "nightly-suse-observability-" + secret.suffix
					explicit := "explicit-" + secret.suffix + "-secret"
					if secret.enabled && secret.external == "" {
						require.Contains(t, before.Secrets, legacy)
						require.Contains(t, after.Secrets, explicit)
						expected := before.Secrets[legacy]
						expected.Name = explicit
						assert.Equal(t, expected, after.Secrets[explicit], "Secret keys and data must be preserved")
						delete(before.Secrets, legacy)
						delete(after.Secrets, explicit)
						renames[legacy] = explicit
						checksums["checksum/"+secret.suffix+"-env"] = true
						if secret.suffix == "license" {
							expectedLicense = explicit
							assert.Equal(t, "fixture-license", string(expected.Data["LICENSE_KEY"]))
						} else {
							expectedEmail = explicit
							assert.Equal(t, "fixture-user", string(expected.Data["SMTP_USER_NAME"]))
							assert.Equal(t, "fixture-password", string(expected.Data["SMTP_PASSWORD"]))
						}
					} else {
						assert.NotContains(t, before.Secrets, legacy)
						assert.NotContains(t, after.Secrets, explicit)
					}
					assert.NotContains(t, after.Secrets, legacy)
					if secret.external != "" {
						assert.NotContains(t, after.Secrets, secret.external, "Customer Secrets must not be created by the chart")
					}
				}
				assert.Equal(t, before.Secrets, after.Secrets, "Authentication and API-key Secrets must stay unchanged")

				for name, deployment := range before.Deployments {
					require.Contains(t, after.Deployments, name)
					expected := deployment.DeepCopy()
					rewriteLicenseEmailSecretReferences(&expected.Spec.Template.Spec, renames)
					actual := after.Deployments[name]
					for key := range checksums {
						if previous, ok := expected.Spec.Template.Annotations[key]; ok {
							require.Contains(t, actual.Spec.Template.Annotations, key)
							assert.NotEqual(t, previous, actual.Spec.Template.Annotations[key], "Deliberately renamed Secrets change their manifest checksums")
							expected.Spec.Template.Annotations[key] = actual.Spec.Template.Annotations[key]
						}
					}
					assert.Equal(t, *expected, actual, name)
				}
				assert.Len(t, after.Deployments, len(before.Deployments))
				app := "nightly-suse-observability-server"
				if split {
					app = "nightly-suse-observability-api"
				}
				require.Contains(t, after.Deployments, app)
				container := after.Deployments[app].Spec.Template.Spec.Containers[0]
				assert.Contains(t, container.EnvFrom, corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: expectedLicense}}})
				if tc.emailEnabled {
					assert.Contains(t, container.EnvFrom, corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: expectedEmail}}})
				}

				backupRefs := 0
				for name, job := range before.CronJobs {
					require.Contains(t, after.CronJobs, name)
					expected := job.DeepCopy()
					rewriteLicenseEmailSecretReferences(&expected.Spec.JobTemplate.Spec.Template.Spec, renames)
					actual := after.CronJobs[name]
					assert.Equal(t, *expected, actual)
					for _, c := range actual.Spec.JobTemplate.Spec.Template.Spec.Containers {
						for _, env := range c.EnvFrom {
							if env.SecretRef != nil && env.SecretRef.Name == expectedLicense {
								backupRefs++
							}
						}
					}
				}
				assert.Positive(t, backupRefs, "Configuration backup must reference the selected license Secret")
				assert.Len(t, after.CronJobs, len(before.CronJobs))
				restoreRefs := 0
				for name, config := range before.ConfigMaps {
					require.Contains(t, after.ConfigMaps, name)
					expected := config
					expected.Data = maps.Clone(config.Data)
					for key, value := range expected.Data {
						for legacy, explicit := range renames {
							value = strings.ReplaceAll(value, legacy, explicit)
						}
						expected.Data[key] = value
						if strings.Contains(value, `name: "`+expectedLicense+`"`) {
							restoreRefs++
						}
					}
					assert.Equal(t, expected, after.ConfigMaps[name], "Embedded restore Job references must follow the selected Secret")
				}
				assert.Positive(t, restoreRefs)
				assert.Len(t, after.ConfigMaps, len(before.ConfigMaps))
				assert.Equal(t, before.Services, after.Services)
				assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
				assert.Equal(t, before.Roles, after.Roles)
				assert.Equal(t, before.RoleBindings, after.RoleBindings)
				assert.Equal(t, before.Statefulsets, after.Statefulsets)
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
			})
		}
	}
}

func rewriteLicenseEmailSecretReferences(pod *corev1.PodSpec, renames map[string]string) {
	for _, containers := range [][]corev1.Container{pod.Containers, pod.InitContainers} {
		for i := range containers {
			for j := range containers[i].EnvFrom {
				ref := containers[i].EnvFrom[j].SecretRef
				if ref != nil {
					if replacement, ok := renames[ref.Name]; ok {
						ref.Name = replacement
					}
				}
			}
		}
	}
}
