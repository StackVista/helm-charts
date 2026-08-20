package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func findVolume(t *testing.T, volumes []corev1.Volume, name string) corev1.Volume {
	for _, volume := range volumes {
		if volume.Name == name {
			return volume
		}
	}
	require.FailNow(t, "volume "+name+" not found")
	return corev1.Volume{}
}

func TestJavaTrustStoreFromExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/java_trust_store_external_secret.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment := resources.Deployments["suse-observability-api"]
	container := deployment.Spec.Template.Spec.Containers[0]

	assert.Contains(t, container.Args, "-Djavax.net.ssl.trustStore=/opt/docker/secrets/java-cacerts")
	assert.Contains(t, container.Args, "-Djavax.net.ssl.trustStoreType=jks")
	assert.Contains(t, container.Args, "-Djavax.net.ssl.trustStorePassword=$(JAVA_TRUSTSTORE_PASSWORD)")

	volume := findVolume(t, deployment.Spec.Template.Spec.Volumes, "service-secrets-volume")
	require.NotNil(t, volume.Projected, "the trust store must be mounted from a projected volume")
	require.Len(t, volume.Projected.Sources, 1)
	assert.Equal(t, "my-java-truststore", volume.Projected.Sources[0].Secret.Name)
	assert.Equal(t, "cacerts", volume.Projected.Sources[0].Secret.Items[0].Key)
	assert.Equal(t, "java-cacerts", volume.Projected.Sources[0].Secret.Items[0].Path)

	var passwordEnv *corev1.EnvVar
	for i, env := range container.Env {
		if env.Name == "JAVA_TRUSTSTORE_PASSWORD" {
			passwordEnv = &container.Env[i]
		}
	}
	require.NotNil(t, passwordEnv)
	assert.Equal(t, "my-java-truststore", passwordEnv.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "cacerts-password", passwordEnv.ValueFrom.SecretKeyRef.Key)

	// The whole point of the external secret: the blob must not be copied into a chart-managed
	// secret, because that puts it back into the Helm release secret via the rendered manifest.
	commonSecret := resources.Secrets["suse-observability-common"]
	assert.NotContains(t, commonSecret.Data, "javaTrustStore")
	assert.NotContains(t, commonSecret.Data, "javaTrustStorePassword")
}

func TestJavaTrustStoreFromExternalSecretWithInlinePassword(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.java.trustStoreFromExternalSecret.name": "my-java-truststore",
			"stackstate.java.trustStorePassword":                "the password",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment := resources.Deployments["suse-observability-api"]
	volume := findVolume(t, deployment.Spec.Template.Spec.Volumes, "service-secrets-volume")
	assert.Equal(t, "my-java-truststore", volume.Projected.Sources[0].Secret.Name)
	assert.Equal(t, "java-cacerts", volume.Projected.Sources[0].Secret.Items[0].Key, "key should default to java-cacerts")

	var passwordEnv *corev1.EnvVar
	for i, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "JAVA_TRUSTSTORE_PASSWORD" {
			passwordEnv = &deployment.Spec.Template.Spec.Containers[0].Env[i]
		}
	}
	require.NotNil(t, passwordEnv)
	assert.Equal(t, "suse-observability-common", passwordEnv.ValueFrom.SecretKeyRef.Name, "password falls back to the chart-managed secret")

	commonSecret := resources.Secrets["suse-observability-common"]
	assert.NotContains(t, commonSecret.Data, "javaTrustStore")
	assert.Contains(t, commonSecret.Data, "javaTrustStorePassword")
}

// Guards the migration path: a user who adds the external secret but leaves the old inline value in
// place must not end up with the trust store still counted against the release secret size limit.
func TestJavaTrustStoreExternalSecretTakesPrecedenceOverInlineValue(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml", "values/java_trust_store_external_secret.yaml"},
		SetValues: map[string]string{
			"stackstate.java.trustStore":         "inline trust store that should be ignored",
			"stackstate.java.trustStorePassword": "inline password that should be ignored",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment := resources.Deployments["suse-observability-api"]
	volume := findVolume(t, deployment.Spec.Template.Spec.Volumes, "service-secrets-volume")
	require.Len(t, volume.Projected.Sources, 1)
	assert.Equal(t, "my-java-truststore", volume.Projected.Sources[0].Secret.Name)
	assert.Equal(t, "cacerts", volume.Projected.Sources[0].Secret.Items[0].Key)

	var passwordEnv *corev1.EnvVar
	for i, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "JAVA_TRUSTSTORE_PASSWORD" {
			passwordEnv = &deployment.Spec.Template.Spec.Containers[0].Env[i]
		}
	}
	require.NotNil(t, passwordEnv)
	assert.Equal(t, "my-java-truststore", passwordEnv.ValueFrom.SecretKeyRef.Name)

	commonSecret := resources.Secrets["suse-observability-common"]
	assert.NotContains(t, commonSecret.Data, "javaTrustStore", "the inline value must not be copied into the chart-managed secret")
	assert.NotContains(t, commonSecret.Data, "javaTrustStorePassword")
}

// The two LDAP blobs are configured independently, so one may come from an external secret while the
// other stays inline.
func TestLdapTrustCertificatesFromExternalSecretWithInlineTrustStore(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml", "values/ldap_authentication.yaml"},
		SetValues: map[string]string{
			"stackstate.authentication.ldap.ssl.trustCertificatesFromExternalSecret.name": "ldap-certs",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	volume := findVolume(t, resources.Deployments["suse-observability-api"].Spec.Template.Spec.Volumes, "service-secrets-volume")
	require.Len(t, volume.Projected.Sources, 2)
	sources := map[string]string{}
	for _, source := range volume.Projected.Sources {
		sources[source.Secret.Items[0].Path] = source.Secret.Name
	}
	assert.Equal(t, map[string]string{
		"ldap-certificates.pem": "ldap-certs",
		"ldap-cacerts":          "suse-observability-common",
	}, sources)

	commonSecret := resources.Secrets["suse-observability-common"]
	assert.Contains(t, commonSecret.Data, "ldapTrustStore")
	assert.NotContains(t, commonSecret.Data, "ldapTrustCertificates")
}

// The LDAP mirror of TestJavaTrustStoreExternalSecretTakesPrecedenceOverInlineValue.
func TestLdapTrustStoreExternalSecretTakesPrecedenceOverInlineValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml", "values/ldap_authentication_external_secret.yaml"},
		SetValues: map[string]string{
			"stackstate.authentication.ldap.ssl.trustStore":        "inline trust store that should be ignored",
			"stackstate.authentication.ldap.ssl.trustCertificates": "inline certificates that should be ignored",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	volume := findVolume(t, resources.Deployments["suse-observability-api"].Spec.Template.Spec.Volumes, "service-secrets-volume")
	require.Len(t, volume.Projected.Sources, 2)
	for _, source := range volume.Projected.Sources {
		assert.Equal(t, "my-ldap-truststore", source.Secret.Name)
	}

	commonSecret := resources.Secrets["suse-observability-common"]
	assert.NotContains(t, commonSecret.Data, "ldapTrustStore", "the inline value must not be copied into the chart-managed secret")
	assert.NotContains(t, commonSecret.Data, "ldapTrustCertificates", "the inline value must not be copied into the chart-managed secret")
}

func TestNoTrustStoreConfigured(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment := resources.Deployments["suse-observability-api"]
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		assert.NotEqual(t, "service-secrets-volume", volume.Name, "no secrets volume without a trust store")
	}
	for _, mount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
		assert.NotEqual(t, "/opt/docker/secrets", mount.MountPath)
	}
	for _, arg := range deployment.Spec.Template.Spec.Containers[0].Args {
		assert.NotContains(t, arg, "javax.net.ssl.trustStore")
	}
	for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		assert.NotEqual(t, "JAVA_TRUSTSTORE_PASSWORD", env.Name)
	}
}

func TestLdapTrustStoreFromExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/ldap_authentication_external_secret.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap := resources.ConfigMaps["suse-observability-api"]
	assert.Contains(t, configMap.Data["application_stackstate.conf"], `trustStorePath = "/opt/docker/secrets/ldap-cacerts"`)
	assert.Contains(t, configMap.Data["application_stackstate.conf"], `trustCertificatesPath = "/opt/docker/secrets/ldap-certificates.pem"`)

	volume := findVolume(t, resources.Deployments["suse-observability-api"].Spec.Template.Spec.Volumes, "service-secrets-volume")
	require.Len(t, volume.Projected.Sources, 2)
	mounted := map[string]string{}
	for _, source := range volume.Projected.Sources {
		assert.Equal(t, "my-ldap-truststore", source.Secret.Name)
		mounted[source.Secret.Items[0].Path] = source.Secret.Items[0].Key
	}
	assert.Equal(t, map[string]string{"ldap-cacerts": "cacerts", "ldap-certificates.pem": "certificates.pem"}, mounted)

	commonSecret := resources.Secrets["suse-observability-common"]
	assert.NotContains(t, commonSecret.Data, "ldapTrustStore")
	assert.NotContains(t, commonSecret.Data, "ldapTrustCertificates")
}

func TestTrustStoresMixExternalAndInlineSecrets(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/ldap_authentication.yaml", "values/java_trust_store_external_secret.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	volume := findVolume(t, resources.Deployments["suse-observability-api"].Spec.Template.Spec.Volumes, "service-secrets-volume")
	require.Len(t, volume.Projected.Sources, 2)

	sources := map[string]string{}
	for _, source := range volume.Projected.Sources {
		sources[source.Secret.Items[0].Path] = source.Secret.Name
	}
	assert.Equal(t, map[string]string{
		"ldap-cacerts": "suse-observability-common",
		"java-cacerts": "my-java-truststore",
	}, sources)

	commonSecret := resources.Secrets["suse-observability-common"]
	assert.Contains(t, commonSecret.Data, "ldapTrustStore")
	assert.NotContains(t, commonSecret.Data, "javaTrustStore")
}

func TestS3ProxyJavaTrustStoreFromExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml", "values/java_trust_store_external_secret.yaml"},
		SetValues:   map[string]string{"global.backup.enabled": "true"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment := resources.Deployments["suse-observability-s3proxy"]
	container := deployment.Spec.Template.Spec.Containers[0]

	var javaOpts string
	var passwordEnv *corev1.EnvVar
	for i, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
		}
		if env.Name == "JAVA_TRUSTSTORE_PASSWORD" {
			passwordEnv = &container.Env[i]
		}
	}
	assert.Contains(t, javaOpts, "-Djavax.net.ssl.trustStore=/opt/s3proxy/secrets/java-cacerts")
	require.NotNil(t, passwordEnv)
	assert.Equal(t, "my-java-truststore", passwordEnv.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "cacerts-password", passwordEnv.ValueFrom.SecretKeyRef.Key)

	volume := findVolume(t, deployment.Spec.Template.Spec.Volumes, "common-secrets")
	require.NotNil(t, volume.Secret)
	assert.Equal(t, "my-java-truststore", volume.Secret.SecretName)
	assert.Equal(t, "cacerts", volume.Secret.Items[0].Key)
	assert.Equal(t, "java-cacerts", volume.Secret.Items[0].Path)
}
