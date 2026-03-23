package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type expectedVCTStorage struct {
	vctName string
	size    string
}

func TestStorageSizing(t *testing.T) {
	expectedStatefulsetStorage := map[string]map[string]expectedVCTStorage{
		"10-nonha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "50Gi"},
			"suse-observability-hbase-stackgraph":     {vctName: "data", size: "50Gi"},
			"suse-observability-hbase-tephra-mono":    {vctName: "snapshot", size: "1Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "60Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "50Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"20-nonha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "50Gi"},
			"suse-observability-hbase-stackgraph":     {vctName: "data", size: "50Gi"},
			"suse-observability-hbase-tephra-mono":    {vctName: "snapshot", size: "1Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "50Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"50-nonha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "50Gi"},
			"suse-observability-hbase-stackgraph":     {vctName: "data", size: "100Gi"},
			"suse-observability-hbase-tephra-mono":    {vctName: "snapshot", size: "1Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "50Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"100-nonha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "100Gi"},
			"suse-observability-hbase-stackgraph":     {vctName: "data", size: "100Gi"},
			"suse-observability-hbase-tephra-mono":    {vctName: "snapshot", size: "1Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "50Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"150-ha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "100Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "200Gi"},
			"suse-observability-hbase-hdfs-dn":        {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-nn":        {vctName: "data", size: "20Gi"},
			"suse-observability-hbase-hdfs-snn":       {vctName: "data", size: "20Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-victoria-metrics-1":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"250-ha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-dn":        {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-nn":        {vctName: "data", size: "20Gi"},
			"suse-observability-hbase-hdfs-snn":       {vctName: "data", size: "20Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-victoria-metrics-1":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"500-ha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-dn":        {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-nn":        {vctName: "data", size: "20Gi"},
			"suse-observability-hbase-hdfs-snn":       {vctName: "data", size: "20Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "400Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-victoria-metrics-1":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"4000-ha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-dn":        {vctName: "data", size: "1000Gi"},
			"suse-observability-hbase-hdfs-nn":        {vctName: "data", size: "20Gi"},
			"suse-observability-hbase-hdfs-snn":       {vctName: "data", size: "20Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "400Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "400Gi"},
			"suse-observability-victoria-metrics-1":   {vctName: "server-volume", size: "400Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"10-nonha-storage-overrides": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "200Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "100Gi"},
			"suse-observability-hbase-stackgraph":     {vctName: "data", size: "300Gi"},
			"suse-observability-hbase-tephra-mono":    {vctName: "snapshot", size: "888Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "150Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "100Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "20Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "999Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "888Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "666Gi"},
		},
		"150-ha-storage-overrides": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "500Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "400Gi"},
			"suse-observability-hbase-hdfs-dn":        {vctName: "data", size: "500Gi"},
			"suse-observability-hbase-hdfs-nn":        {vctName: "data", size: "444Gi"},
			"suse-observability-hbase-hdfs-snn":       {vctName: "data", size: "555Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "200Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "500Gi"},
			"suse-observability-victoria-metrics-1":   {vctName: "server-volume", size: "500Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "16Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "999Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "888Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "666Gi"},
		},
		"default-nonha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-stackgraph":     {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-tephra-mono":    {vctName: "snapshot", size: "1Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
		"default-ha": {
			"suse-observability-clickhouse-shard0":    {vctName: "data", size: "50Gi"},
			"suse-observability-elasticsearch-master": {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-dn":        {vctName: "data", size: "250Gi"},
			"suse-observability-hbase-hdfs-nn":        {vctName: "data", size: "20Gi"},
			"suse-observability-hbase-hdfs-snn":       {vctName: "data", size: "20Gi"},
			"suse-observability-kafka":                {vctName: "data", size: "100Gi"},
			"suse-observability-victoria-metrics-0":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-victoria-metrics-1":   {vctName: "server-volume", size: "250Gi"},
			"suse-observability-zookeeper":            {vctName: "data", size: "8Gi"},
			"suse-observability-vmagent":              {vctName: "tmpdata", size: "10Gi"},
			"suse-observability-workload-observer":    {vctName: "data", size: "1Gi"},
			"suse-observability-ai-assistant":         {vctName: "data", size: "1Gi"},
		},
	}

	testCases := []struct {
		name        string
		valuesFiles []string
		expected    map[string]expectedVCTStorage
	}{
		// Legacy sizing profiles
		{
			name:        "10-nonha",
			valuesFiles: []string{"values/values_sizing_10_nonha.yaml"},
			expected:    expectedStatefulsetStorage["10-nonha"],
		},
		{
			name:        "20-nonha",
			valuesFiles: []string{"values/values_sizing_20_nonha.yaml"},
			expected:    expectedStatefulsetStorage["20-nonha"],
		},
		{
			name:        "50-nonha",
			valuesFiles: []string{"values/values_sizing_50_nonha.yaml"},
			expected:    expectedStatefulsetStorage["50-nonha"],
		},
		{
			name:        "100-nonha",
			valuesFiles: []string{"values/values_sizing_100_nonha.yaml"},
			expected:    expectedStatefulsetStorage["100-nonha"],
		},
		{
			name:        "150-ha",
			valuesFiles: []string{"values/values_sizing_150_ha.yaml"},
			expected:    expectedStatefulsetStorage["150-ha"],
		},
		{
			name:        "250-ha",
			valuesFiles: []string{"values/values_sizing_250_ha.yaml"},
			expected:    expectedStatefulsetStorage["250-ha"],
		},
		{
			name:        "500-ha",
			valuesFiles: []string{"values/values_sizing_500_ha.yaml"},
			expected:    expectedStatefulsetStorage["500-ha"],
		},
		{
			name:        "4000-ha",
			valuesFiles: []string{"values/values_sizing_4000_ha.yaml"},
			expected:    expectedStatefulsetStorage["4000-ha"],
		},
		// Global sizing profiles
		{
			name:        "global-10-nonha",
			valuesFiles: []string{"values/global_sizing_10_nonha.yaml"},
			expected:    expectedStatefulsetStorage["10-nonha"],
		},
		{
			name:        "global-20-nonha",
			valuesFiles: []string{"values/global_sizing_20_nonha.yaml"},
			expected:    expectedStatefulsetStorage["20-nonha"],
		},
		{
			name:        "global-50-nonha",
			valuesFiles: []string{"values/global_sizing_50_nonha.yaml"},
			expected:    expectedStatefulsetStorage["50-nonha"],
		},
		{
			name:        "global-100-nonha",
			valuesFiles: []string{"values/global_sizing_100_nonha.yaml"},
			expected:    expectedStatefulsetStorage["100-nonha"],
		},
		{
			name:        "global-150-ha",
			valuesFiles: []string{"values/global_sizing_150_ha.yaml"},
			expected:    expectedStatefulsetStorage["150-ha"],
		},
		{
			name:        "global-250-ha",
			valuesFiles: []string{"values/global_sizing_250_ha.yaml"},
			expected:    expectedStatefulsetStorage["250-ha"],
		},
		{
			name:        "global-500-ha",
			valuesFiles: []string{"values/global_sizing_500_ha.yaml"},
			expected:    expectedStatefulsetStorage["500-ha"],
		},
		{
			name:        "global-4000-ha",
			valuesFiles: []string{"values/global_sizing_4000_ha.yaml"},
			expected:    expectedStatefulsetStorage["4000-ha"],
		},
		{
			name:        "default-nonha",
			valuesFiles: []string{"values/default_nonha_sizing.yaml"},
			expected:    expectedStatefulsetStorage["default-nonha"],
		},
		{
			name:        "default-ha",
			valuesFiles: []string{"values/default_ha_sizing.yaml"},
			expected:    expectedStatefulsetStorage["default-ha"],
		},
		// Storage override tests
		{
			name:        "global-10-nonha-storage-overrides",
			valuesFiles: []string{"values/global_sizing_10_nonha_storage_override.yaml"},
			expected:    expectedStatefulsetStorage["10-nonha-storage-overrides"],
		},
		{
			name:        "global-150-ha-storage-overrides",
			valuesFiles: []string{"values/global_sizing_150_ha_storage_override.yaml"},
			expected:    expectedStatefulsetStorage["150-ha-storage-overrides"],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFiles...)
			resources := helmtestutil.NewKubernetesResources(t, output)

			// Check that all statefulsets with VCTs are covered by the test
			t.Run("completeness", func(t *testing.T) {
				for name, ss := range resources.Statefulsets {
					if len(ss.Spec.VolumeClaimTemplates) > 0 {
						if _, ok := tc.expected[name]; !ok {
							t.Errorf("StatefulSet %q has volumeClaimTemplates but is not covered by the test, please add it", name)
						}
					}
				}
				for name := range tc.expected {
					if _, ok := resources.Statefulsets[name]; !ok {
						t.Errorf("StatefulSet %q is listed in the test but not rendered by the chart, please remove it", name)
					}
				}
			})

			for name, expected := range tc.expected {
				t.Run("statefulset-"+name, func(t *testing.T) {
					ss, exists := resources.Statefulsets[name]
					require.True(t, exists, "StatefulSet %s should exist", name)
					assertVCTStorage(t, name, ss.Spec.VolumeClaimTemplates, expected)
				})
			}
		})
	}
}

func assertVCTStorage(t *testing.T, ssName string, vcts []corev1.PersistentVolumeClaim, expected expectedVCTStorage) {
	t.Helper()
	var found bool
	for _, vct := range vcts {
		if vct.Name == expected.vctName {
			found = true
			actual := vct.Spec.Resources.Requests[corev1.ResourceStorage]
			assert.Equal(t, resource.MustParse(expected.size), actual,
				"StatefulSet %s VCT %q storage should be %s", ssName, expected.vctName, expected.size)
			break
		}
	}
	require.True(t, found, "StatefulSet %s should have a volumeClaimTemplate named %q", ssName, expected.vctName)
}
