package test

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// exemptedContainers lists the containers that cannot satisfy the restricted Pod Security
// Standard. Every entry must also be documented in the required-permissions page, because
// customers running restricted PSA or CIS RKE2 have to exempt the namespace or disable the
// container.
var exemptedContainers = map[string]string{
	"configure-sysctl": "raises vm.max_map_count for Elasticsearch, needs privileged and root; " +
		"disable with elasticsearch.sysctlInitContainer.enabled=false",
}

// exemptedGatedContainers are only rendered by the feature gates in
// values/restricted_security_context_gates.yaml, and all of them chown a data directory as
// root before the main container starts, so dropping capabilities would defeat their purpose.
var exemptedGatedContainers = map[string]string{
	"volume-permissions": "chowns the data volume as root for ClickHouse, Kafka and ZooKeeper; " +
		"off by default behind <subchart>.volumePermissions.enabled",
	"datanode-init":  "HDFS volume chown as root, off by default behind hbase.hdfs.volumePermissions",
	"namenode-init":  "HDFS volume chown as root, off by default behind hbase.hdfs.volumePermissions",
	"snamenode-init": "HDFS volume chown as root, off by default behind hbase.hdfs.volumePermissions",
}

func TestRestrictedSecurityContext(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/restricted_security_context.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	helmtestutil.AssertRestrictedSecurityContext(t, resources, exemptedContainers)
}

func TestRbacAgentHasNoFixedPodIdentity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/restricted_security_context.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-rbac-agent"]
	require.True(t, ok)

	podSpec := deployment.Spec.Template.Spec
	require.NotNil(t, podSpec.SecurityContext)
	assert.Nil(t, podSpec.SecurityContext.RunAsUser)
	assert.Nil(t, podSpec.SecurityContext.RunAsGroup)
	assert.Nil(t, podSpec.SecurityContext.FSGroup)
	require.NotNil(t, podSpec.SecurityContext.RunAsNonRoot)
	assert.True(t, *podSpec.SecurityContext.RunAsNonRoot)

	require.Len(t, podSpec.Containers, 1)
	container := podSpec.Containers[0]
	assert.Contains(t, container.Env, corev1.EnvVar{Name: "STS_ROLE_TYPE", Value: "instance"})
}

// A container behind an unset feature gate is invisible to the render above, which is how the
// router-mode upgrade hooks stayed non-compliant. This covers every gate that adds a container
// or changes an existing one's securityContext; the list was derived by rendering the chart
// once per `enabled: false` value in the chart and its subcharts, and the other gates produced
// no new containers.
func TestRestrictedSecurityContextFeatureGates(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability",
		"values/restricted_security_context.yaml",
		"values/restricted_security_context_gates.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	exempted := maps.Clone(exemptedContainers)
	maps.Copy(exempted, exemptedGatedContainers)

	helmtestutil.AssertRestrictedSecurityContext(t, resources, exempted)
}
