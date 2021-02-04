package helmtestutil

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesResources parsed from a multi-document template string
type KubernetesResources struct {
	ClusterRoles          []rbacv1.ClusterRole
	ClusterRoleBindings   []rbacv1.ClusterRoleBinding
	ConfigMaps            []corev1.ConfigMap
	CronJobs              []batchv1beta1.CronJob
	Deployments           []appsv1.Deployment
	Jobs                  []batchv1.Job
	PersistentVolumeClaim []corev1.PersistentVolumeClaim
	Pods                  []corev1.Pod
	Pdbs                  []policyv1.PodDisruptionBudget
	Roles                 []rbacv1.Role
	RoleBindings          []rbacv1.RoleBinding
	Secrets               []corev1.Secret
	Services              []corev1.Service
	ServiceAccounts       []corev1.ServiceAccount
	Statefulsets          []appsv1.StatefulSet
}

// NewKubernetesResources creates a new instance of KubernetesResources by parsing the helmOutput and
// loading the data into the correct types resulting in some (limited) type checks on the data as well
func NewKubernetesResources(t *testing.T, helmOutput string) KubernetesResources {
	clusterRoles := make([]rbacv1.ClusterRole, 0)
	clusterRoleBindings := make([]rbacv1.ClusterRoleBinding, 0)
	configMaps := make([]corev1.ConfigMap, 0)
	cronJobs := make([]batchv1beta1.CronJob, 0)
	deployments := make([]appsv1.Deployment, 0)
	jobs := make([]batchv1.Job, 0)
	persistentVolumeClaims := make([]corev1.PersistentVolumeClaim, 0)
	pods := make([]corev1.Pod, 0)
	pdbs := make([]policyv1.PodDisruptionBudget, 0)
	roles := make([]rbacv1.Role, 0)
	roleBindings := make([]rbacv1.RoleBinding, 0)
	secrets := make([]corev1.Secret, 0)
	services := make([]corev1.Service, 0)
	serviceAccounts := make([]corev1.ServiceAccount, 0)
	statefulsets := make([]appsv1.StatefulSet, 0)

	// The K8S unmarshalling only can do a single document
	// So we split them first by the yaml document separator
	separateFiles := strings.Split(helmOutput, "\n---\n")

	for _, v := range separateFiles {
		var metadata metav1.TypeMeta
		helm.UnmarshalK8SYaml(t, v, &metadata)
		switch metadata.Kind {
		case "ClusterRole":
			var resource rbacv1.ClusterRole
			helm.UnmarshalK8SYaml(t, v, &resource)
			clusterRoles = append(clusterRoles, resource)
		case "ClusterRoleBinding":
			var resource rbacv1.ClusterRoleBinding
			helm.UnmarshalK8SYaml(t, v, &resource)
			clusterRoleBindings = append(clusterRoleBindings, resource)
		case "ConfigMap":
			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(t, v, &configMap)
			configMaps = append(configMaps, configMap)
		case "CronJob":
			var resource batchv1beta1.CronJob
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "CronJob failed to parse: "+v)
			cronJobs = append(cronJobs, resource)
		case "Deployment":
			var resource appsv1.Deployment
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "Deployment failed to parse: "+v)
			deployments = append(deployments, resource)
		case "Job":
			var resource batchv1.Job
			helm.UnmarshalK8SYaml(t, v, &resource)
			jobs = append(jobs, resource)
		case "List":
			// Skip for now
		case "PersistentVolumeClaim":
			var resource corev1.PersistentVolumeClaim
			e := helm.UnmarshalK8SYamlE(t, v, &resource)
			assert.NoError(t, e, "PersistentVolumeClaim failed to parse: "+v)
			persistentVolumeClaims = append(persistentVolumeClaims, resource)
		case "Pod":
			var resource corev1.Pod
			helm.UnmarshalK8SYaml(t, v, &resource)
			pods = append(pods, resource)
		case "PodDisruptionBudget":
			var resource policyv1.PodDisruptionBudget
			helm.UnmarshalK8SYaml(t, v, &resource)
			pdbs = append(pdbs, resource)
		case "Role":
			var resource rbacv1.Role
			helm.UnmarshalK8SYaml(t, v, &resource)
			roles = append(roles, resource)
		case "RoleBinding":
			var resource rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(t, v, &resource)
			roleBindings = append(roleBindings, resource)
		case "Secret":
			var resource corev1.Secret
			helm.UnmarshalK8SYaml(t, v, &resource)
			secrets = append(secrets, resource)
		case "Service":
			var resource corev1.Service
			helm.UnmarshalK8SYaml(t, v, &resource)
			services = append(services, resource)
		case "ServiceAccount":
			var resource corev1.ServiceAccount
			helm.UnmarshalK8SYaml(t, v, &resource)
			serviceAccounts = append(serviceAccounts, resource)
		case "StatefulSet":
			var resource appsv1.StatefulSet
			helm.UnmarshalK8SYaml(t, v, &resource)
			statefulsets = append(statefulsets, resource)
		default:
			t.Error("Found unknown kind " + metadata.Kind + ". This can be caused by an incorrect k8s resource type in the helm template or when using a custom resource type.")
		}
	}

	return KubernetesResources{
		ClusterRoles:        clusterRoles,
		ClusterRoleBindings: clusterRoleBindings,
		ConfigMaps:          configMaps,
		Deployments:         deployments,
		Jobs:                jobs,
		Pods:                pods,
		Pdbs:                pdbs,
		Roles:               roles,
		RoleBindings:        roleBindings,
		Secrets:             secrets,
		Services:            services,
		ServiceAccounts:     serviceAccounts,
		Statefulsets:        statefulsets,
	}
}
