package test

import (
	"encoding/base64"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestCommonSecretReferencesFollowDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	definition := regexp.MustCompile(`(?s)\{\{- define "stackstate.common.secret.fullname" -\}\}.*?\{\{- end -\}\}`)
	require.Len(t, definition.FindAll(data, -1), 1)
	data = definition.ReplaceAll(data, []byte(`{{- define "stackstate.common.secret.fullname" -}}explicit-common-secret{{- end -}}`))
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	ldapValues, err := filepath.Abs("values/ldap_authentication.yaml")
	require.NoError(t, err)
	inline := map[string]string{
		"stackstate.java.trustStore":                           "java-trust",
		"stackstate.java.trustStorePassword":                   "java-password",
		"stackstate.authentication.ldap.ssl.trustStore":        "ldap-trust",
		"stackstate.authentication.ldap.ssl.trustCertificates": "ldap-certs",
	}
	blobs := map[string]string{
		"javaTrustStore": "java-trust", "javaTrustStorePassword": "java-password",
		"ldapTrustStore": "ldap-trust", "ldapTrustCertificates": "ldap-certs",
	}
	external := map[string]string{
		"stackstate.java.trustStoreFromExternalSecret.name":                           "customer-java",
		"stackstate.java.trustStoreFromExternalSecret.key":                            "java-data",
		"stackstate.java.trustStoreFromExternalSecret.passwordKey":                    "java-password-key",
		"stackstate.authentication.ldap.ssl.trustStoreFromExternalSecret.name":        "customer-ldap",
		"stackstate.authentication.ldap.ssl.trustStoreFromExternalSecret.key":         "ldap-data",
		"stackstate.authentication.ldap.ssl.trustCertificatesFromExternalSecret.name": "customer-certs",
		"stackstate.authentication.ldap.ssl.trustCertificatesFromExternalSecret.key":  "cert-data",
	}
	for _, split := range []bool{false, true} {
		for _, scenario := range []string{"empty", "env-only", "inline", "encoded", "external", "mixed-java", "mixed-ldap"} {
			t.Run(fmt.Sprintf("%s/split=%t", scenario, split), func(t *testing.T) {
				values := map[string]string{"stackstate.features.server.split": fmt.Sprint(split)}
				expectedData := map[string]string{}
				if scenario != "empty" {
					values["stackstate.components.all.extraEnv.secret.SHARED_SETTING"] = "shared-value"
					expectedData["SHARED_SETTING"] = "shared-value"
				}
				withTrust := scenario != "empty" && scenario != "env-only"
				if withTrust {
					maps.Copy(values, inline)
					maps.Copy(expectedData, blobs)
				}
				switch scenario {
				case "encoded":
					for key, value := range inline {
						if key == "stackstate.java.trustStorePassword" {
							continue
						}
						values[key] = ""
						values[key+"Base64Encoded"] = base64.StdEncoding.EncodeToString([]byte(value))
					}
				case "external":
					maps.Copy(values, external)
					for key := range blobs {
						delete(expectedData, key)
					}
				case "mixed-java":
					values["stackstate.java.trustStoreFromExternalSecret.name"] = "customer-java"
					values["stackstate.authentication.ldap.ssl.trustCertificatesFromExternalSecret.name"] = "customer-certs"
					delete(expectedData, "javaTrustStore")
					delete(expectedData, "ldapTrustCertificates")
				case "mixed-ldap":
					values["stackstate.authentication.ldap.ssl.trustStoreFromExternalSecret.name"] = "customer-ldap"
					delete(expectedData, "ldapTrustStore")
				}
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{fullValues}
				if withTrust {
					options.ValuesFiles = append(options.ValuesFiles, ldapValues)
				}
				before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
				require.NoError(t, err)
				after := helmtestutil.NewKubernetesResources(t, output)
				legacy, explicit := "nightly-suse-observability-common", "explicit-common-secret"
				require.Contains(t, before.Secrets, legacy)
				require.Contains(t, after.Secrets, explicit)
				expectedSecret := before.Secrets[legacy]
				expectedSecret.Name = explicit
				assert.Equal(t, expectedSecret, after.Secrets[explicit])
				require.Len(t, after.Secrets[explicit].Data, len(expectedData))
				for key, value := range expectedData {
					assert.Equal(t, value, string(after.Secrets[explicit].Data[key]), key)
				}
				delete(before.Secrets, legacy)
				delete(after.Secrets, explicit)
				assert.NotContains(t, after.Secrets, legacy)
				assert.Equal(t, before.Secrets, after.Secrets)
				references := 0
				for name, deployment := range before.Deployments {
					require.Contains(t, after.Deployments, name)
					expected := deployment.DeepCopy()
					references += rewriteCommonSecretPod(t, &expected.Spec.Template, after.Deployments[name].Spec.Template, legacy, explicit)
					assert.Equal(t, *expected, after.Deployments[name])
				}
				assert.Len(t, after.Deployments, len(before.Deployments))
				for name, statefulset := range before.Statefulsets {
					require.Contains(t, after.Statefulsets, name)
					expected := statefulset.DeepCopy()
					references += rewriteCommonSecretPod(t, &expected.Spec.Template, after.Statefulsets[name].Spec.Template, legacy, explicit)
					assert.Equal(t, *expected, after.Statefulsets[name])
				}
				assert.Len(t, after.Statefulsets, len(before.Statefulsets))
				afterJobs := commonSecretJobsByStableName(t, after.Jobs)
				for stableName, job := range commonSecretJobsByStableName(t, before.Jobs) {
					require.Contains(t, afterJobs, stableName)
					expected := job.DeepCopy()
					actual := afterJobs[stableName]
					expected.Name = actual.Name // Only the existing timestamp suffix may differ.
					references += rewriteCommonSecretPod(t, &expected.Spec.Template, actual.Spec.Template, legacy, explicit)
					assert.Equal(t, *expected, actual)
				}
				assert.Len(t, afterJobs, len(before.Jobs))
				for name, job := range before.CronJobs {
					require.Contains(t, after.CronJobs, name)
					expected := job.DeepCopy()
					references += rewriteCommonSecretPod(t, &expected.Spec.JobTemplate.Spec.Template, after.CronJobs[name].Spec.JobTemplate.Spec.Template, legacy, explicit)
					assert.Equal(t, *expected, after.CronJobs[name])
				}
				assert.Len(t, after.CronJobs, len(before.CronJobs))
				if scenario == "empty" {
					assert.Zero(t, references)
				} else {
					assert.Positive(t, references)
				}
				if withTrust {
					// S3Proxy uses a direct Secret volume, while core services
					// use projected volumes. Both must follow the resolver.
					require.Contains(t, after.Deployments, "suse-observability-s3proxy")
					s3 := after.Deployments["suse-observability-s3proxy"].Spec.Template.Spec
					volume := findVolume(t, s3.Volumes, "common-secrets")
					require.NotNil(t, volume.Secret)
					if scenario == "external" || scenario == "mixed-java" {
						assert.Equal(t, "customer-java", volume.Secret.SecretName)
					} else {
						assert.Equal(t, explicit, volume.Secret.SecretName)
					}
				}
				assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
				assert.Equal(t, before.DaemonSets, after.DaemonSets)
				assert.Equal(t, before.Services, after.Services)
				assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
				assert.Equal(t, before.Roles, after.Roles)
				assert.Equal(t, before.RoleBindings, after.RoleBindings)
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
			})
		}
	}
}

func rewriteCommonSecretPod(t *testing.T, expected *corev1.PodTemplateSpec, actual corev1.PodTemplateSpec, oldName, newName string) int {
	t.Helper()
	references := 0
	rename := func(name *string) {
		if *name == oldName {
			*name = newName
			references++
		}
	}
	for _, containers := range [][]corev1.Container{expected.Spec.Containers, expected.Spec.InitContainers} {
		for i := range containers {
			for j := range containers[i].Env {
				if value := containers[i].Env[j].ValueFrom; value != nil && value.SecretKeyRef != nil {
					rename(&value.SecretKeyRef.Name)
				}
			}
			for j := range containers[i].EnvFrom {
				if ref := containers[i].EnvFrom[j].SecretRef; ref != nil {
					rename(&ref.Name)
				}
			}
		}
	}
	for i := range expected.Spec.Volumes {
		volume := &expected.Spec.Volumes[i]
		if volume.Secret != nil {
			rename(&volume.Secret.SecretName)
		}
		if volume.Projected != nil {
			for j := range volume.Projected.Sources {
				if secret := volume.Projected.Sources[j].Secret; secret != nil {
					rename(&secret.Name)
				}
			}
		}
	}
	for _, key := range []string{"checksum/common-env", "checksum/common-secret"} {
		if previous, ok := expected.Annotations[key]; ok {
			require.Contains(t, actual.Annotations, key)
			assert.NotEqual(t, previous, actual.Annotations[key], "Deliberate test rename must affect existing checksum annotations")
			expected.Annotations[key] = actual.Annotations[key]
		}
	}
	return references
}

func commonSecretJobsByStableName(t *testing.T, jobs map[string]batchv1.Job) map[string]batchv1.Job {
	t.Helper()
	result := map[string]batchv1.Job{}
	timestamp := regexp.MustCompile(`[0-9]{2}t[0-9]{6}$`)
	for name, job := range jobs {
		stableName := timestamp.ReplaceAllString(name, "<timestamp>")
		require.NotContains(t, result, stableName)
		result[stableName] = job
	}
	return result
}
