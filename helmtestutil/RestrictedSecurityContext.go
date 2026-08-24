package helmtestutil

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// podSpecRef pairs a pod spec with enough identity to name it in a failure message.
type podSpecRef struct {
	kind string
	name string
	spec corev1.PodSpec
}

// AssertRestrictedSecurityContext checks that every container and init container in
// resources satisfies the four securityContext settings the restricted Pod Security
// Standard requires, as documented for the SUSE Observability charts:
//
//	capabilities.drop            contains ALL
//	allowPrivilegeEscalation     false
//	seccompProfile.type          RuntimeDefault
//	runAsNonRoot                 true
//
// The first two are container-only fields; the last two may be inherited from the pod.
//
// exemptContainers holds container names that cannot satisfy the profile, keyed by name
// rather than by workload because several workload names contain a timestamp or a random
// suffix. An entry that matches no container fails the test, so the list cannot go stale
// while the exemption silently keeps applying.
func AssertRestrictedSecurityContext(t *testing.T, resources KubernetesResources, exemptContainers map[string]string) {
	t.Helper()

	matched := map[string]bool{}
	checked := 0

	for _, ref := range collectPodSpecs(resources) {
		for _, group := range []struct {
			label      string
			containers []corev1.Container
		}{
			{"initContainer", ref.spec.InitContainers},
			{"container", ref.spec.Containers},
		} {
			for _, container := range group.containers {
				if _, exempt := exemptContainers[container.Name]; exempt {
					matched[container.Name] = true
					continue
				}
				checked++
				where := fmt.Sprintf("%s/%s %s %s", ref.kind, ref.name, group.label, container.Name)
				for _, violation := range restrictedViolations(ref.spec.SecurityContext, container.SecurityContext) {
					assert.Fail(t, "securityContext does not satisfy the restricted Pod Security Standard",
						"%s: %s\nAdd the restricted securityContext to this container, or add %q to the exemption list and document it in the required-permissions page.",
						where, violation, container.Name)
				}
			}
		}
	}

	assert.NotZero(t, checked, "no containers were found to check — the render is probably empty")

	for name, reason := range exemptContainers {
		assert.True(t, matched[name],
			"container %q is exempted (%s) but no longer exists in the rendered chart — remove the exemption", name, reason)
	}
}

func restrictedViolations(pod *corev1.PodSecurityContext, container *corev1.SecurityContext) []string {
	var violations []string

	if container == nil || container.Capabilities == nil || !containsCapability(container.Capabilities.Drop, "ALL") {
		violations = append(violations, "capabilities.drop must contain ALL")
	}
	if container == nil || container.AllowPrivilegeEscalation == nil || *container.AllowPrivilegeEscalation {
		violations = append(violations, "allowPrivilegeEscalation must be false")
	}

	seccomp := (*corev1.SeccompProfile)(nil)
	if container != nil && container.SeccompProfile != nil {
		seccomp = container.SeccompProfile
	} else if pod != nil && pod.SeccompProfile != nil {
		seccomp = pod.SeccompProfile
	}
	if seccomp == nil || seccomp.Type != corev1.SeccompProfileTypeRuntimeDefault {
		violations = append(violations, "seccompProfile.type must be RuntimeDefault")
	}

	runAsNonRoot := (*bool)(nil)
	if container != nil && container.RunAsNonRoot != nil {
		runAsNonRoot = container.RunAsNonRoot
	} else if pod != nil && pod.RunAsNonRoot != nil {
		runAsNonRoot = pod.RunAsNonRoot
	}
	if runAsNonRoot == nil || !*runAsNonRoot {
		violations = append(violations, "runAsNonRoot must be true")
	}

	return violations
}

func containsCapability(capabilities []corev1.Capability, want corev1.Capability) bool {
	for _, capability := range capabilities {
		if capability == want {
			return true
		}
	}
	return false
}

func collectPodSpecs(resources KubernetesResources) []podSpecRef {
	var refs []podSpecRef

	for _, r := range resources.Deployments {
		refs = append(refs, podSpecRef{"Deployment", r.Name, r.Spec.Template.Spec})
	}
	for _, r := range resources.Statefulsets {
		refs = append(refs, podSpecRef{"StatefulSet", r.Name, r.Spec.Template.Spec})
	}
	for _, r := range resources.DaemonSets {
		refs = append(refs, podSpecRef{"DaemonSet", r.Name, r.Spec.Template.Spec})
	}
	for _, r := range resources.Jobs {
		refs = append(refs, podSpecRef{"Job", r.Name, r.Spec.Template.Spec})
	}
	for _, r := range resources.CronJobs {
		refs = append(refs, podSpecRef{"CronJob", r.Name, cronJobPodSpec(r)})
	}
	for _, r := range resources.Pods {
		refs = append(refs, podSpecRef{"Pod", r.Name, r.Spec})
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].kind != refs[j].kind {
			return refs[i].kind < refs[j].kind
		}
		return refs[i].name < refs[j].name
	})

	return refs
}

func cronJobPodSpec(cronJob batchv1beta1.CronJob) corev1.PodSpec {
	return cronJob.Spec.JobTemplate.Spec.Template.Spec
}
