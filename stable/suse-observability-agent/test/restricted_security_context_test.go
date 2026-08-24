package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// exemptedContainers lists the containers that cannot satisfy the restricted Pod Security
// Standard. Every entry must also be documented in the required-privileges page, because
// customers running restricted PSA or CIS RKE2 have to exempt the agent namespace.
var exemptedContainers = map[string]string{
	"node-agent":    "hostPID and container runtime sockets for process-to-container mapping",
	"process-agent": "privileged, for eBPF injection into every network namespace and conntrack reads",
	"logs-agent":    "reads container logs from the host filesystem, needs the spc_t SELinux type",
}

func TestRestrictedSecurityContext(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/restricted_security_context.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	helmtestutil.AssertRestrictedSecurityContext(t, resources, exemptedContainers)
}

// fullPrivilegesMode is a debugging escape hatch, so it is expected to break the profile for
// the agent containers it targets. The rest of the chart must stay compliant.
func TestRestrictedSecurityContextFullPrivilegesMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/restricted_security_context.yaml"},
		SetValues:   map[string]string{"all.fullPrivilegesMode.enabled": "true"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	privileged := map[string]string{
		"cluster-agent":            "fullPrivilegesMode",
		"suse-observability-agent": "fullPrivilegesMode",
	}
	for name, reason := range exemptedContainers {
		privileged[name] = reason
	}

	helmtestutil.AssertRestrictedSecurityContext(t, resources, privileged)
}

// remote-kube-cache and the header injector run images that declare no USER, so they carry an
// explicit runAsUser to satisfy runAsNonRoot on plain Kubernetes. OpenShift's restricted-v2 SCC
// assigns a UID from the namespace range instead and rejects an explicit one outside it, and
// these service accounts are not covered by the node-agent SCC. The UID must therefore be
// removable without losing the rest of the profile.
func TestRestrictedSecurityContextUidOverridableForOpenShift(t *testing.T) {
	uidBearing := map[string]string{
		"remote-kube-cache":    "remoteKubeCache.containerSecurityContext.runAsUser",
		"http-header-injector": "httpHeaderInjectorWebhook.containerSecurityContext.runAsUser",
		"webhook-cert-setup":   "httpHeaderInjectorWebhook.containerSecurityContext.runAsUser",
		"webhook-cert-delete":  "httpHeaderInjectorWebhook.containerSecurityContext.runAsUser",
	}

	t.Run("uid present by default", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
			"values/restricted_security_context.yaml")
		for name, sc := range containerSecurityContexts(t, output) {
			if _, ok := uidBearing[name]; !ok {
				continue
			}
			require.NotNil(t, sc.RunAsUser, "%s needs a UID on plain Kubernetes: its image declares no USER", name)
			assert.EqualValues(t, 65534, *sc.RunAsUser, "%s should default to 65534", name)
		}
	})

	t.Run("uid removable for openshift", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
			ValuesFiles: []string{"values/restricted_security_context.yaml"},
			SetValues: map[string]string{
				"remoteKubeCache.containerSecurityContext.runAsUser":           "null",
				"httpHeaderInjectorWebhook.containerSecurityContext.runAsUser": "null",
			},
		})
		resources := helmtestutil.NewKubernetesResources(t, output)

		found := 0
		for name, sc := range containerSecurityContexts(t, output) {
			if _, ok := uidBearing[name]; !ok {
				continue
			}
			found++
			assert.Nil(t, sc.RunAsUser, "%s should carry no explicit UID once the value is nulled", name)
			require.NotNil(t, sc.RunAsNonRoot, "%s must keep runAsNonRoot", name)
			assert.True(t, *sc.RunAsNonRoot, "%s must keep runAsNonRoot even without a UID", name)
		}
		assert.Equal(t, len(uidBearing), found, "expected every UID-bearing container to render")

		// Nulling the UID must not weaken anything else.
		helmtestutil.AssertRestrictedSecurityContext(t, resources, exemptedContainers)
	})
}

// containerSecurityContexts maps container name to securityContext across every workload in a
// render. Container names are unique enough here, and workload names are not stable.
func containerSecurityContexts(t *testing.T, output string) map[string]*corev1.SecurityContext {
	t.Helper()
	resources := helmtestutil.NewKubernetesResources(t, output)

	out := map[string]*corev1.SecurityContext{}
	collect := func(spec corev1.PodSpec) {
		for _, c := range append(append([]corev1.Container{}, spec.InitContainers...), spec.Containers...) {
			out[c.Name] = c.SecurityContext
		}
	}
	for _, r := range resources.Deployments {
		collect(r.Spec.Template.Spec)
	}
	for _, r := range resources.Statefulsets {
		collect(r.Spec.Template.Spec)
	}
	for _, r := range resources.DaemonSets {
		collect(r.Spec.Template.Spec)
	}
	for _, r := range resources.Jobs {
		collect(r.Spec.Template.Spec)
	}

	return out
}
