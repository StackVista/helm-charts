package test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	chartName       = "codeartifact-proxy"
	statefulSetName = "codeartifact-proxy"
	nginxContainer  = "codeartifact-proxy"
	sidecarName     = "codeartifact-token-refresh"
	bootstrapName   = "codeartifact-token-bootstrap"
	cacheVolume     = "nginx-cache"
	cacheMountPath  = "/var/cache/nginx"
	nginxTmpPath    = "/tmp"
)

func render(t *testing.T, files ...string) helmtestutil.KubernetesResources {
	output := helmtestutil.RenderHelmTemplate(t, chartName, append([]string{"values/required.yaml"}, files...)...)
	return helmtestutil.NewKubernetesResources(t, output)
}

func renderWith(t *testing.T, setValues map[string]string) helmtestutil.KubernetesResources {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, chartName, &helm.Options{
		ValuesFiles: []string{"values/required.yaml"},
		SetValues:   setValues,
	})
	return helmtestutil.NewKubernetesResources(t, output)
}

func requireStatefulSet(t *testing.T, resources helmtestutil.KubernetesResources) appsv1.StatefulSet {
	statefulSet, ok := resources.Statefulsets[statefulSetName]
	require.True(t, ok, "codeartifact-proxy StatefulSet should exist")
	return statefulSet
}

func requirePodSpec(t *testing.T, resources helmtestutil.KubernetesResources) corev1.PodSpec {
	return requireStatefulSet(t, resources).Spec.Template.Spec
}

func containerByName(t *testing.T, containers []corev1.Container, name string) corev1.Container {
	for _, container := range containers {
		if container.Name == name {
			return container
		}
	}
	require.Failf(t, "container not found", "no container named %q", name)
	return corev1.Container{}
}

func mountByName(t *testing.T, mounts []corev1.VolumeMount, name string) corev1.VolumeMount {
	for _, mount := range mounts {
		if mount.Name == name {
			return mount
		}
	}
	require.Failf(t, "volume mount not found", "no volume mount named %q", name)
	return corev1.VolumeMount{}
}

func serverBlock(t *testing.T, resources helmtestutil.KubernetesResources) string {
	configMap, ok := resources.ConfigMaps["codeartifact-proxy-nginx-config"]
	require.True(t, ok, "nginx-config ConfigMap should exist")
	return configMap.Data["default.conf"]
}

func TestTokenRefreshSidecarAndBootstrapPresent(t *testing.T) {
	podSpec := requirePodSpec(t, render(t))

	require.Len(t, podSpec.InitContainers, 1, "exactly one initContainer expected")
	bootstrap := containerByName(t, podSpec.InitContainers, bootstrapName)
	assert.Contains(t, bootstrap.Args, "/scripts/refresh-token.sh")
	assert.Contains(t, bootstrap.Env, corev1.EnvVar{Name: "RUN_ONCE", Value: "true"})

	require.Len(t, podSpec.Containers, 2, "nginx plus token-refresh sidecar expected")
	sidecar := containerByName(t, podSpec.Containers, sidecarName)
	assert.Contains(t, sidecar.Args, "/scripts/refresh-token.sh")
	assert.Contains(t, sidecar.Env, corev1.EnvVar{Name: "RUN_ONCE", Value: "false"})
	require.NotNil(t, sidecar.ReadinessProbe, "sidecar should report token validity through a readinessProbe")
}

func TestAuthIncludeVolumeIsMemoryBacked(t *testing.T) {
	podSpec := requirePodSpec(t, render(t))

	var authVolume *corev1.Volume
	for i := range podSpec.Volumes {
		if podSpec.Volumes[i].Name == "codeartifact-auth" {
			authVolume = &podSpec.Volumes[i]
		}
	}
	require.NotNil(t, authVolume, "codeartifact-auth volume should exist")
	require.NotNil(t, authVolume.EmptyDir, "codeartifact-auth should be an emptyDir")
	assert.Equal(t, corev1.StorageMediumMemory, authVolume.EmptyDir.Medium,
		"the CodeArtifact token include must never touch node disk")

	nginx := containerByName(t, podSpec.Containers, nginxContainer)
	for _, mount := range nginx.VolumeMounts {
		if mount.Name == "codeartifact-auth" {
			assert.True(t, mount.ReadOnly, "nginx must not be able to write the auth include")
		}
	}
}

// Caching is a primary function of this proxy, so the cache must be a per-replica
// volumeClaimTemplate that survives pod restarts and rollouts - not an emptyDir discarded with the
// pod.
func TestCacheIsAPersistentVolumeClaimTemplate(t *testing.T) {
	resources := render(t)
	statefulSet := requireStatefulSet(t, resources)

	require.Len(t, statefulSet.Spec.VolumeClaimTemplates, 1, "exactly one cache claim template expected")
	claim := statefulSet.Spec.VolumeClaimTemplates[0]
	assert.Equal(t, cacheVolume, claim.Name)
	assert.Equal(t, []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}, claim.Spec.AccessModes)
	assert.False(t, claim.Spec.Resources.Requests.Storage().IsZero(), "the cache claim needs a size")

	podSpec := statefulSet.Spec.Template.Spec
	for _, volume := range podSpec.Volumes {
		require.NotEqual(t, cacheVolume, volume.Name,
			"the cache must come from the claim template, not a pod volume")
	}

	nginx := containerByName(t, podSpec.Containers, nginxContainer)
	mount := mountByName(t, nginx.VolumeMounts, cacheVolume)
	assert.Equal(t, cacheMountPath, mount.MountPath)
	assert.False(t, mount.ReadOnly, "nginx writes cache entries here")

	// Both cache zones must live on that persistent mount, or the cache silently reverts to being
	// pod-lifetime only.
	config := serverBlock(t, resources)
	assert.Contains(t, config, "proxy_cache_path "+cacheMountPath+"/codeartifact-assets ")
	assert.Contains(t, config, "proxy_cache_path "+cacheMountPath+"/codeartifact-metadata ")
}

// The nginx container has readOnlyRootFilesystem: true, and nginx mkdir()s every configured
// *_temp_path at load time. Any temp path left at the image default resolves under the read-only
// root and kills the process with `mkdir() ... failed (30: Read-only file system)`, so the chart
// must declare all of them - and the pid file - on a writable mount.
func TestNginxWritablePathsAreAllOnMountedVolumes(t *testing.T) {
	resources := render(t)
	podSpec := requirePodSpec(t, resources)
	nginx := containerByName(t, podSpec.Containers, nginxContainer)

	require.NotNil(t, nginx.SecurityContext.ReadOnlyRootFilesystem)
	require.True(t, *nginx.SecurityContext.ReadOnlyRootFilesystem)

	writableMounts := []string{}
	for _, mount := range nginx.VolumeMounts {
		if !mount.ReadOnly {
			writableMounts = append(writableMounts, mount.MountPath)
		}
	}
	assert.Contains(t, writableMounts, nginxTmpPath, "nginx needs a writable temp directory")

	configMap, ok := resources.ConfigMaps["codeartifact-proxy-nginx-config"]
	require.True(t, ok, "nginx-config ConfigMap should exist")
	nginxConf := configMap.Data["nginx.conf"]
	require.NotEmpty(t, nginxConf, "the chart must own the whole nginx.conf to redirect the temp paths")

	// Every directive nginx mkdir()s or writes to at startup, plus its pid file.
	for _, directive := range []string{
		"pid",
		"client_body_temp_path",
		"proxy_temp_path",
		"fastcgi_temp_path",
		"uwsgi_temp_path",
		"scgi_temp_path",
	} {
		pattern := regexp.MustCompile(`(?m)^\s*` + directive + `\s+(\S+?);`)
		match := pattern.FindStringSubmatch(nginxConf)
		require.NotNil(t, match, "nginx.conf must set %s explicitly, or nginx uses its read-only compiled-in default", directive)

		onWritableMount := false
		for _, mountPath := range writableMounts {
			if strings.HasPrefix(match[1], mountPath+"/") {
				onWritableMount = true
			}
		}
		assert.True(t, onWritableMount, "%s is %q, which is not under any writable mount %v", directive, match[1], writableMounts)
	}
}

// nginx runs against the chart's own config, so the base image's nginx.conf and conf.d/ (which
// declare a conflicting default server block) are never read.
func TestNginxRunsAgainstChartOwnedConfig(t *testing.T) {
	resources := render(t)
	podSpec := requirePodSpec(t, resources)
	nginx := containerByName(t, podSpec.Containers, nginxContainer)

	mount := mountByName(t, nginx.VolumeMounts, "nginx-config")
	assert.Equal(t, "/etc/nginx/config", mount.MountPath)
	assert.True(t, mount.ReadOnly)

	configMap, ok := resources.ConfigMaps["codeartifact-proxy-nginx-entrypoint"]
	require.True(t, ok, "nginx-entrypoint ConfigMap should exist")
	entrypoint := configMap.Data["nginx-entrypoint.sh"]
	assert.Contains(t, entrypoint, `nginx -c "$CONFIG_FILE" -e /dev/stderr -t`,
		"the entrypoint should validate the config and fail fast rather than start a pod that cannot serve")
	assert.Contains(t, entrypoint, `nginx -c "$CONFIG_FILE" -e /dev/stderr -g 'daemon off;'`)
	assert.Contains(t, entrypoint, "mkdir -p",
		"nginx only mkdir()s the leaf temp directories, so the entrypoint must create the tree")
}

// The AWS CLI writes under $HOME while assuming the IRSA role, and the root filesystem is
// read-only, so HOME must point at the sidecar's writable /tmp.
func TestSidecarContainersHaveWritableHome(t *testing.T) {
	podSpec := requirePodSpec(t, render(t))

	for _, container := range append(append([]corev1.Container{}, podSpec.InitContainers...), containerByName(t, podSpec.Containers, sidecarName)) {
		if container.Name == nginxContainer {
			continue
		}
		home := ""
		for _, env := range container.Env {
			if env.Name == "HOME" {
				home = env.Value
			}
		}
		require.NotEmpty(t, home, "%s must set HOME away from the read-only root", container.Name)
		mount := mountByName(t, container.VolumeMounts, "sidecar-tmp")
		assert.Equal(t, home, mount.MountPath, "%s HOME should be its writable volume", container.Name)
		assert.False(t, mount.ReadOnly)
	}
}

func TestBothContainersAreHardened(t *testing.T) {
	podSpec := requirePodSpec(t, render(t))

	containers := append(append([]corev1.Container{}, podSpec.Containers...), podSpec.InitContainers...)
	for _, container := range containers {
		securityContext := container.SecurityContext
		require.NotNil(t, securityContext, "%s should have a securityContext", container.Name)
		require.NotNil(t, securityContext.RunAsNonRoot)
		assert.True(t, *securityContext.RunAsNonRoot, "%s should run as non-root", container.Name)
		require.NotNil(t, securityContext.ReadOnlyRootFilesystem)
		assert.True(t, *securityContext.ReadOnlyRootFilesystem, "%s should have a read-only root filesystem", container.Name)
		require.NotNil(t, securityContext.AllowPrivilegeEscalation)
		assert.False(t, *securityContext.AllowPrivilegeEscalation, "%s should not allow privilege escalation", container.Name)
		require.NotNil(t, securityContext.Capabilities)
		assert.Equal(t, []corev1.Capability{"ALL"}, securityContext.Capabilities.Drop,
			"%s should drop all capabilities", container.Name)
	}
}

// The sidecar reloads nginx through a token-generation marker on the shared volume, so a shared PID
// namespace (which would also expose /proc/<pid>/environ across containers) is never needed.
func TestProcessNamespaceIsNotShared(t *testing.T) {
	podSpec := requirePodSpec(t, render(t))

	require.NotNil(t, podSpec.ShareProcessNamespace)
	assert.False(t, *podSpec.ShareProcessNamespace)
}

func TestServerBlockRoutes(t *testing.T) {
	config := serverBlock(t, render(t))

	assert.Contains(t, config, "location /maven/releases/ {")
	assert.Contains(t, config, "location /maven/snapshots/ {")
	assert.Contains(t, config, "location /pypi/ {")
	assert.Contains(t, config, `location ~* ^/pypi/simple/[^/]+/?$ {`)
	assert.Contains(t, config, `include /etc/nginx/config/proxy-common.inc;`)
	assert.Contains(t, config, "/maven/packages/")
	assert.Contains(t, config, "/maven/packages-snapshot/")
	assert.Contains(t, config, "/pypi/packages/")

	// Generic packages are fetched with SigV4 at their call sites, never through this proxy.
	assert.NotContains(t, config, "/generic/")
}

// nginx rejects a proxy_pass with a URI part inside a regex location, so the metadata/index
// locations must map the client path with `rewrite ... break` instead.
func TestServerBlockRegexLocationsHaveNoProxyPassUri(t *testing.T) {
	config := serverBlock(t, render(t))

	inRegexLocation := false
	for _, line := range strings.Split(config, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "location ~"):
			inRegexLocation = true
		case strings.HasPrefix(trimmed, "location "):
			inRegexLocation = false
		case inRegexLocation && strings.HasPrefix(trimmed, "proxy_pass "):
			target := strings.TrimSuffix(strings.TrimPrefix(trimmed, "proxy_pass "), ";")
			assert.NotContains(t, strings.TrimPrefix(target, "https://"), "/",
				"proxy_pass in a regex location must have no URI part: %s", trimmed)
		case inRegexLocation && strings.HasPrefix(trimmed, "rewrite "):
			assert.Contains(t, trimmed, "break;", "the path mapping must not re-run location matching")
		}
	}
}

// nginx accumulates proxy_set_header values within a level, and drops the ones with an empty value,
// so this default guarantees the client's own credentials are not forwarded even if the
// sidecar-written include were ever missing.
func TestProxyCommonClearsClientAuthorizationBeforeIncludingToken(t *testing.T) {
	configMap, ok := render(t).ConfigMaps["codeartifact-proxy-nginx-config"]
	require.True(t, ok)
	common := configMap.Data["proxy-common.inc"]

	clearIndex := strings.Index(common, `proxy_set_header Authorization "";`)
	includeIndex := strings.Index(common, "include /etc/nginx/auth/codeartifact-auth.conf;")
	require.NotEqual(t, -1, clearIndex, "the empty-Authorization default should be present")
	require.NotEqual(t, -1, includeIndex, "the sidecar-written auth include should be present")
	assert.Less(t, clearIndex, includeIndex, "the sidecar include must come last so its token wins")
}

// The only `Authorization "Basic ..."` allowed in the manifests is the sidecar script's printf
// format string; a rendered literal value would mean a token had entered Helm.
func TestNoAuthorizationTokenInRenderedManifests(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, chartName, "values/required.yaml")

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, `Authorization "Basic`) {
			continue
		}
		assert.Contains(t, line, `Basic %s`, "unexpected literal Authorization value: %s", line)
	}
	assert.NotContains(t, output, "authorizationToken\":")
}

func TestServiceAccountCarriesIrsaAnnotation(t *testing.T) {
	const roleArn = "arn:aws:iam::000000000000:role/StackStateRoleCodeArtifactProxy"
	resources := renderWith(t, map[string]string{
		`serviceAccount.annotations.eks\.amazonaws\.com/role-arn`: roleArn,
	})

	serviceAccount, ok := resources.ServiceAccounts["codeartifact-proxy"]
	require.True(t, ok, "ServiceAccount should exist")
	assert.Equal(t, roleArn, serviceAccount.Annotations["eks.amazonaws.com/role-arn"])
	require.NotNil(t, serviceAccount.AutomountServiceAccountToken)
	assert.True(t, *serviceAccount.AutomountServiceAccountToken, "IRSA needs the projected token")
}

func TestPlaceholderEndpointIsRejected(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, chartName)
	require.Contains(t, err.Error(), "codeartifact_packages_maven_endpoint")
}

func TestPlaceholderDigestIsRejected(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, chartName, &helm.Options{
		ValuesFiles: []string{"values/required.yaml"},
		SetValues: map[string]string{
			"image.digest": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "image.digest")
}
