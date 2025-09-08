package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// TestKafkaBasicRender tests that the Kafka chart renders successfully with default values
func TestKafkaBasicRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/default.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Validate that essential resources are created
	assert.NotEmpty(t, resources.Statefulsets, "StatefulSet should be created")
	assert.NotEmpty(t, resources.Services, "Services should be created")
	assert.NotEmpty(t, resources.ConfigMaps, "ConfigMaps should be created")
	assert.NotEmpty(t, resources.ServiceAccounts, "ServiceAccount should be created")
}

// TestKafkaDefaultValuesProducesExpectedResources tests that default values produce the expected Kubernetes resources
func TestKafkaDefaultValuesProducesExpectedResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test StatefulSet
	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	assert.Equal(t, "kafka", statefulset.Name, "StatefulSet should be named 'kafka'")
	assert.Equal(t, int32(1), *statefulset.Spec.Replicas, "Default replica count should be 1")

	// Validate container configuration
	containers := statefulset.Spec.Template.Spec.Containers
	assert.NotEmpty(t, containers, "StatefulSet should have containers")

	kafkaContainer := containers[0]
	assert.Equal(t, "kafka", kafkaContainer.Name, "Main container should be named 'kafka'")
	assert.Contains(t, kafkaContainer.Image, "bitnami/kafka", "Should use Bitnami Kafka image")

	// Test Services
	assert.GreaterOrEqual(t, len(resources.Services), 1, "Should have at least one Service")

	// Find the main Kafka service
	var kafkaService corev1.Service
	var headlessService corev1.Service

	for _, svc := range resources.Services {
		if svc.Spec.ClusterIP != "None" {
			kafkaService = svc
		} else {
			headlessService = svc
		}
	}

	// Validate main service
	assert.NotEmpty(t, kafkaService.Name, "Main Kafka service should exist")
	assert.Equal(t, corev1.ServiceTypeClusterIP, kafkaService.Spec.Type, "Default service type should be ClusterIP")

	// Check for expected ports
	portFound := false
	for _, port := range kafkaService.Spec.Ports {
		if port.Port == 9092 {
			portFound = true
			assert.Equal(t, "tcp-client", port.Name, "Client port should be named 'tcp-client'")
			break
		}
	}
	assert.True(t, portFound, "Service should expose port 9092 for client connections")

	// Validate headless service
	assert.NotEmpty(t, headlessService.Name, "Headless service should exist")
	assert.Equal(t, "None", headlessService.Spec.ClusterIP, "Headless service should have ClusterIP: None")

	// Test ServiceAccount
	assert.Len(t, resources.ServiceAccounts, 1, "Should have exactly one ServiceAccount")

	// Test ConfigMap (if created by default)
	if len(resources.ConfigMaps) > 0 {
		for _, cm := range resources.ConfigMaps {
			assert.NotEmpty(t, cm.Name, "ConfigMap should have a name")
			assert.NotNil(t, cm.Data, "ConfigMap should have data")
		}
	}
}

// TestKafkaWithCustomReplicas tests scaling Kafka with custom replica count
func TestKafkaWithCustomReplicas(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/custom-replicas.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Validate StatefulSet replica count
	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	assert.Equal(t, int32(3), *statefulset.Spec.Replicas, "Should have 3 replicas as specified in custom values")
}
