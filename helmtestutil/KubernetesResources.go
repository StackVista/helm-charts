package helmtestutil

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesResources parsed from a multi-document template string
type KubernetesResources struct {
	ClusterRoles           map[string]rbacv1.ClusterRole
	ClusterRoleBindings    map[string]rbacv1.ClusterRoleBinding
	ConfigMaps             map[string]corev1.ConfigMap
	CronJobs               map[string]batchv1beta1.CronJob
	DaemonSets             map[string]appsv1.DaemonSet
	Deployments            map[string]appsv1.Deployment
	Ingresses              map[string]networkingv1.Ingress
	Jobs                   map[string]batchv1.Job
	PersistentVolumeClaims map[string]corev1.PersistentVolumeClaim
	Pods                   map[string]corev1.Pod
	Pdbs                   map[string]policyv1.PodDisruptionBudget
	Roles                  map[string]rbacv1.Role
	RoleBindings           map[string]rbacv1.RoleBinding
	Secrets                map[string]corev1.Secret
	Services               map[string]corev1.Service
	ServiceAccounts        map[string]corev1.ServiceAccount
	Statefulsets           map[string]appsv1.StatefulSet
	ServiceMonitors        map[string]promv1.ServiceMonitor
	Unmapped               map[string]string
}

// NewKubernetesResources creates a new instance of KubernetesResources by parsing the helmOutput and
// loading the data into the correct types resulting in some (limited) type checks on the data as well
func NewKubernetesResources(t *testing.T, helmOutput string) KubernetesResources {
	clusterRoles := make(map[string]rbacv1.ClusterRole)
	clusterRoleBindings := make(map[string]rbacv1.ClusterRoleBinding)
	configMaps := make(map[string]corev1.ConfigMap)
	cronJobs := make(map[string]batchv1beta1.CronJob)
	daemonSets := make(map[string]appsv1.DaemonSet)
	deployments := make(map[string]appsv1.Deployment)
	ingresses := make(map[string]networkingv1.Ingress)
	jobs := make(map[string]batchv1.Job)
	persistentVolumeClaims := make(map[string]corev1.PersistentVolumeClaim)
	pods := make(map[string]corev1.Pod)
	pdbs := make(map[string]policyv1.PodDisruptionBudget)
	roles := make(map[string]rbacv1.Role)
	roleBindings := make(map[string]rbacv1.RoleBinding)
	secrets := make(map[string]corev1.Secret)
	services := make(map[string]corev1.Service)
	serviceAccounts := make(map[string]corev1.ServiceAccount)
	statefulsets := make(map[string]appsv1.StatefulSet)
	serviceMonitors := make(map[string]promv1.ServiceMonitor)
	unmapped := map[string]string{}

	// The K8S unmarshalling only can do a single document
	// So we split them first by the yaml document separator
	separateFiles := strings.Split(helmOutput, "\n---\n")

	for _, v := range separateFiles {
		if helmDocWithoutResources(v) {
			// Skip empty resources, e.g. the helm chart has empty Multidoc sections, like
			// ---
			// # Source: suse-observability/charts/anomaly-detection/templates/anomaly-detection-manager.yaml
			// ---
			continue
		}
		var partial metav1.PartialObjectMetadata
		err := helm.UnmarshalK8SYamlE(t, v, &partial)
		if err != nil {
			// See whether we got some warnings
			whitelisted := []string{
				// Known issue in zookeeper chart 10.x
				"warning: skipped value for kafka.zookeeper.topologySpreadConstraints: Not a table.",
				// Older helm verison reports this warning:
				"warning: skipped value for topologySpreadConstraints: Not a table.",
				// Known issue in zookeeper chart:
				"warning: skipped value for hbase.zookeeper.updateStrategy: Not a table.",
				// Known issue in zookeeper chart:
				"warning: skipped value for clickhouse.zookeeper.topologySpreadConstraints: Not a table.",
				// Unknown issue for OpenTelemetry exporters configuration
				"warning: cannot overwrite table with non table for suse-observability.opentelemetry-collector.config.exporters.logging (map[])",
				"warning: cannot overwrite table with non table for suse-observability.opentelemetry-collector.config.receivers.jaeger (map[protocols:map[grpc:map[endpoint:${env:MY_POD_IP}:14250] thrift_compact:map[endpoint:${env:MY_POD_IP}:6831] thrift_http:map[endpoint:${env:MY_POD_IP}:14268]]])",
				"warning: cannot overwrite table with non table for suse-observability.opentelemetry-collector.config.receivers.prometheus (map[config:map[scrape_configs:[map[job_name:opentelemetry-collector scrape_interval:10s static_configs:[map[targets:[${env:MY_POD_IP}:8888]]]]]]])",
				"warning: cannot overwrite table with non table for suse-observability.opentelemetry-collector.config.receivers.zipkin (map[endpoint:${env:MY_POD_IP}:9411])",
			}

			for _, line := range strings.Split(v, "\n") {
				if !contains(whitelisted, line) && line != "" {
					t.Logf("Unknown resource or unexpected error/warning: %s", v)
					assert.NoError(t, err)
				}
			}

			continue
		}

		switch partial.Kind {
		case "ClusterRole":
			var resource rbacv1.ClusterRole
			helm.UnmarshalK8SYaml(t, v, &resource)
			clusterRoles[resource.Name] = resource
		case "ClusterRoleBinding":
			var resource rbacv1.ClusterRoleBinding
			helm.UnmarshalK8SYaml(t, v, &resource)
			clusterRoleBindings[resource.Name] = resource
		case "ConfigMap":
			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(t, v, &configMap)
			configMaps[configMap.Name] = configMap
		case "CronJob":
			var resource batchv1beta1.CronJob
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "CronJob failed to parse: "+v)
			cronJobs[resource.Name] = resource
		case "DaemonSet":
			var resource appsv1.DaemonSet
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "DaemonSet failed to parse: "+v)
			daemonSets[resource.Name] = resource
		case "Deployment":
			var resource appsv1.Deployment
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "Deployment failed to parse: "+v)
			deployments[resource.Name] = resource
		case "Ingress":
			var resource networkingv1.Ingress
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "Ingress failed to parse: "+v)
			ingresses[resource.Name] = resource
		case "Job":
			var resource batchv1.Job
			helm.UnmarshalK8SYaml(t, v, &resource)
			jobs[resource.Name] = resource
		case "List":
			// Skip for now
		case "PersistentVolumeClaim":
			var resource corev1.PersistentVolumeClaim
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "PersistentVolumeClaim failed to parse: "+v)
			persistentVolumeClaims[resource.Name] = resource
		case "Pod":
			var resource corev1.Pod
			helm.UnmarshalK8SYaml(t, v, &resource)
			pods[resource.Name] = resource
		case "PodDisruptionBudget":
			var resource policyv1.PodDisruptionBudget
			helm.UnmarshalK8SYaml(t, v, &resource)
			pdbs[resource.Name] = resource
		case "Role":
			var resource rbacv1.Role
			helm.UnmarshalK8SYaml(t, v, &resource)
			roles[resource.Name] = resource
		case "RoleBinding":
			var resource rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(t, v, &resource)
			roleBindings[resource.Name] = resource
		case "Secret":
			var resource corev1.Secret
			helm.UnmarshalK8SYaml(t, v, &resource)
			secrets[resource.Name] = resource
		case "Service":
			var resource corev1.Service
			helm.UnmarshalK8SYaml(t, v, &resource)
			services[resource.Name] = resource
		case "ServiceAccount":
			var resource corev1.ServiceAccount
			helm.UnmarshalK8SYaml(t, v, &resource)
			serviceAccounts[resource.Name] = resource
		case "StatefulSet":
			var resource appsv1.StatefulSet
			helm.UnmarshalK8SYaml(t, v, &resource)
			statefulsets[resource.Name] = resource
		case "ServiceMonitor":
			var resource promv1.ServiceMonitor
			helm.UnmarshalK8SYaml(t, v, &resource)
			serviceMonitors[resource.Name] = resource
		default:
			if partial.Kind != "" || partial.APIVersion != "" {
				t.Logf("Found unknown kind '%s/%s' in content\n%s\n This can be caused by an incorrect k8s resource type in the helm template or when using a custom resource type.", partial.APIVersion, partial.Kind, v)
				unmapped[partial.Name] = v
			} else {
				t.Logf("Skipping empty resource: %s", v)
			}
		}
	}

	return KubernetesResources{
		ClusterRoles:           clusterRoles,
		ClusterRoleBindings:    clusterRoleBindings,
		ConfigMaps:             configMaps,
		DaemonSets:             daemonSets,
		Deployments:            deployments,
		Ingresses:              ingresses,
		CronJobs:               cronJobs,
		Jobs:                   jobs,
		PersistentVolumeClaims: persistentVolumeClaims,
		Pods:                   pods,
		Pdbs:                   pdbs,
		Roles:                  roles,
		RoleBindings:           roleBindings,
		Secrets:                secrets,
		Services:               services,
		ServiceAccounts:        serviceAccounts,
		Statefulsets:           statefulsets,
		ServiceMonitors:        serviceMonitors,
		Unmapped:               unmapped,
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.Contains(str, v) {
			return true
		}
	}

	return false
}

func helmDocWithoutResources(doc string) bool {
	return !strings.Contains(doc, "\n") && strings.HasPrefix(doc, "# Source:")
}
