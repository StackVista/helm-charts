package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// The volumePermissions init containers chown the data volume as root. Their securityContext
// must come from the values verbatim, not from a merge with common.container: sprig's merge
// treats the values' runAsNonRoot: false as a zero value and lets common.container's true
// through, which combined with runAsUser: 0 makes the kubelet refuse to create the container
// (CreateContainerConfigError). That shipped broken from ~2021 until STAC-25627.
//
// hdfs.volumePermissions.enabled is the only switch: it used to need a second
// securityContext.enabled alongside it, and setting it alone left the chown running as the
// pod's non-root UID, where it fails with EPERM.
func TestHBaseHdfsVolumePermissionsSecurityContext(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, &helm.Options{
		ValuesFiles: []string{"values/distributed-mode.yaml"},
		SetValues:   map[string]string{"hdfs.volumePermissions.enabled": "true"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	wanted := map[string]string{
		"hdfs-dn":  "datanode-init",
		"hdfs-nn":  "namenode-init",
		"hdfs-snn": "snamenode-init",
	}

	checked := 0
	for _, sts := range resources.Statefulsets {
		for suffix, containerName := range wanted {
			if !strings.HasSuffix(sts.Name, suffix) {
				continue
			}

			var container *corev1.Container
			for i := range sts.Spec.Template.Spec.InitContainers {
				if sts.Spec.Template.Spec.InitContainers[i].Name == containerName {
					container = &sts.Spec.Template.Spec.InitContainers[i]
				}
			}
			require.NotNil(t, container, "init container %q not found in %q", containerName, sts.Name)

			sc := container.SecurityContext
			require.NotNil(t, sc, "%s should have a securityContext", containerName)

			require.NotNil(t, sc.RunAsNonRoot, "%s must set runAsNonRoot explicitly", containerName)
			assert.False(t, *sc.RunAsNonRoot,
				"%s runs as root to chown the data volume, so runAsNonRoot must be false; "+
					"true here means common.container leaked in through the merge", containerName)

			require.NotNil(t, sc.RunAsUser, "%s must set runAsUser", containerName)
			assert.EqualValues(t, 0, *sc.RunAsUser, "%s chowns as root", containerName)

			assert.Nil(t, sc.Capabilities,
				"%s must not drop capabilities: it needs CAP_CHOWN, and dropping ALL is "+
					"common.container leaking in through the merge", containerName)
			assert.Nil(t, sc.SeccompProfile,
				"%s should carry only the values securityContext; a seccompProfile means "+
					"common.container leaked in through the merge", containerName)

			checked++
		}
	}
	assert.Equal(t, len(wanted), checked, "expected one init container per HDFS StatefulSet")
}

// A leftover securityContext.enabled from before the two flags were collapsed must not leak
// into the rendered securityContext, where it would be an invalid field.
func TestHBaseHdfsVolumePermissionsIgnoresLegacyEnabledKey(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, &helm.Options{
		ValuesFiles: []string{"values/distributed-mode.yaml"},
		SetValues: map[string]string{
			"hdfs.volumePermissions.enabled":                 "true",
			"hdfs.volumePermissions.securityContext.enabled": "true",
		},
	})

	// Parsed untyped: decoding into corev1.SecurityContext would silently drop the stray key.
	keys := datanodeInitSecurityContextKeys(t, output)
	assert.ElementsMatch(t, []string{"runAsNonRoot", "runAsUser"}, keys,
		"the container securityContext should carry only the values fields, with the legacy "+
			"securityContext.enabled key omitted")
	assert.Contains(t, output, "chown -R", "the chown should still be in the init script")
}

// datanodeInitSecurityContextKeys returns the securityContext field names on datanode-init,
// read from the raw render so that fields Kubernetes would reject are still visible.
func datanodeInitSecurityContextKeys(t *testing.T, output string) []string {
	t.Helper()

	for _, doc := range strings.Split(output, "\n---\n") {
		var parsed struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				Template struct {
					Spec struct {
						InitContainers []struct {
							Name            string                 `json:"name"`
							SecurityContext map[string]interface{} `json:"securityContext"`
						} `json:"initContainers"`
					} `json:"spec"`
				} `json:"template"`
			} `json:"spec"`
		}
		if err := yaml.Unmarshal([]byte(doc), &parsed); err != nil {
			continue
		}
		if parsed.Kind != "StatefulSet" || !strings.HasSuffix(parsed.Metadata.Name, "hdfs-dn") {
			continue
		}
		for _, c := range parsed.Spec.Template.Spec.InitContainers {
			if c.Name != "datanode-init" {
				continue
			}
			keys := make([]string, 0, len(c.SecurityContext))
			for k := range c.SecurityContext {
				keys = append(keys, k)
			}

			return keys
		}
	}
	t.Fatal("datanode-init not found in the rendered output")

	return nil
}

// With the switch off, the init container keeps the ordinary restricted profile and no chown.
func TestHBaseHdfsVolumePermissionsDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	checked := 0
	for _, sts := range resources.Statefulsets {
		if !strings.HasSuffix(sts.Name, "hdfs-dn") {
			continue
		}
		for _, c := range sts.Spec.Template.Spec.InitContainers {
			sc := c.SecurityContext
			require.NotNil(t, sc, "%s should have a securityContext", c.Name)
			require.NotNil(t, sc.RunAsNonRoot)
			assert.True(t, *sc.RunAsNonRoot, "%s should stay non-root when the switch is off", c.Name)
			assert.Nil(t, sc.RunAsUser, "%s should not pin a UID when the switch is off", c.Name)
			require.NotNil(t, sc.Capabilities, "%s should drop capabilities when the switch is off", c.Name)
			checked++
		}
	}
	assert.NotZero(t, checked, "expected an hdfs-dn init container")
}
