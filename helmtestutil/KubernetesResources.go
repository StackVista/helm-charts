package helmtestutil

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesResources parsed from a multi-document template string
type KubernetesResources struct {
	ConfigMaps          []corev1.ConfigMap
	Pdbs                []policyv1.PodDisruptionBudget
	ServiceAccounts     []corev1.ServiceAccount
	Secrets             []corev1.Secret
	ClusterRoles        []rbacv1.ClusterRole
	ClusterRoleBindings []rbacv1.ClusterRoleBinding
	Services            []corev1.Service
	Deployments         []appsv1.Deployment
	Statefulsets        []appsv1.StatefulSet
	Jobs                []batchv1.Job
	Pods                []corev1.Pod
}

// NewKubernetesResources creates a new instance of KubernetesResources by parsing the helmOutput and
// loading the data into the correct types resulting in some (limited) type checks on the data as well
func NewKubernetesResources(t *testing.T, helmOutput string) KubernetesResources {
	configMaps := make([]corev1.ConfigMap, 0)
	pdbs := make([]policyv1.PodDisruptionBudget, 0)
	serviceAccounts := make([]corev1.ServiceAccount, 0)
	secrets := make([]corev1.Secret, 0)
	clusterRoles := make([]rbacv1.ClusterRole, 0)
	clusterRoleBindings := make([]rbacv1.ClusterRoleBinding, 0)
	services := make([]corev1.Service, 0)
	deployments := make([]appsv1.Deployment, 0)
	statefulsets := make([]appsv1.StatefulSet, 0)
	jobs := make([]batchv1.Job, 0)
	pods := make([]corev1.Pod, 0)

	// The K8S unmarshalling only can do a single document
	// So we split them first by the yaml document separator
	separateFiles := strings.Split(helmOutput, "\n---\n")

	for _, v := range separateFiles {
		var metadata metav1.TypeMeta
		helm.UnmarshalK8SYaml(t, v, &metadata)
		switch metadata.Kind {
		case "ConfigMap":
			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(t, v, &configMap)
			configMaps = append(configMaps, configMap)
		case "PodDisruptionBudget":
			var resource policyv1.PodDisruptionBudget
			helm.UnmarshalK8SYaml(t, v, &resource)
			pdbs = append(pdbs, resource)
		case "ServiceAccount":
			var resource corev1.ServiceAccount
			helm.UnmarshalK8SYaml(t, v, &resource)
			serviceAccounts = append(serviceAccounts, resource)
		case "Secret":
			var resource corev1.Secret
			helm.UnmarshalK8SYaml(t, v, &resource)
			secrets = append(secrets, resource)
		case "ClusterRole":
			var resource rbacv1.ClusterRole
			helm.UnmarshalK8SYaml(t, v, &resource)
			clusterRoles = append(clusterRoles, resource)
		case "ClusterRoleBinding":
			var resource rbacv1.ClusterRoleBinding
			helm.UnmarshalK8SYaml(t, v, &resource)
			clusterRoleBindings = append(clusterRoleBindings, resource)
		case "Service":
			var resource corev1.Service
			helm.UnmarshalK8SYaml(t, v, &resource)
			services = append(services, resource)
		case "Deployment":
			var resource appsv1.Deployment
			helm.UnmarshalK8SYaml(t, v, &resource)
			deployments = append(deployments, resource)
		case "StatefulSet":
			var resource appsv1.StatefulSet
			helm.UnmarshalK8SYaml(t, v, &resource)
			statefulsets = append(statefulsets, resource)
		case "Job":
			var resource batchv1.Job
			helm.UnmarshalK8SYaml(t, v, &resource)
			jobs = append(jobs, resource)
		case "Pod":
			var resource corev1.Pod
			helm.UnmarshalK8SYaml(t, v, &resource)
			pods = append(pods, resource)
		case "List":
			// Skip for now
		default:
			t.Error("Found unknown kind " + metadata.Kind)
		}
	}

	return KubernetesResources{
		ConfigMaps:          configMaps,
		Pdbs:                pdbs,
		ServiceAccounts:     serviceAccounts,
		Secrets:             secrets,
		ClusterRoles:        clusterRoles,
		ClusterRoleBindings: clusterRoleBindings,
		Services:            services,
		Deployments:         deployments,
		Statefulsets:        statefulsets,
		Jobs:                jobs,
		Pods:                pods,
	}
}
