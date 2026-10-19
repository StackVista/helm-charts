package test

import (
	"strconv"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestKubernetesStateMetricsAgentTags(t *testing.T) {
	for _, tc := range []struct {
		name        string
		labels      bool
		annotations bool
	}{
		{name: "empty"},
		{name: "labels", labels: true},
		{name: "annotations", annotations: true},
		{name: "both", labels: true, annotations: true},
	} {
		for _, clusterCheck := range []bool{false, true} {
			t.Run(tc.name+"/clusterCheck="+strconv.FormatBool(clusterCheck), func(t *testing.T) {
				values := map[string]string{
					"clusterAgent.collection.kubeStateMetrics.clusterCheck": strconv.FormatBool(clusterCheck),
				}
				labels := map[string]map[string]string{}
				annotations := map[string]map[string]string{}
				if tc.labels {
					values[`clusterAgent.collection.kubeStateMetrics.labelsAsTags.node.node-role\.kubernetes\.io/control-plane`] = "kube_node_role_control_plane"
					values["clusterAgent.collection.kubeStateMetrics.labelsAsTags.node.team"] = "node_team"
					values["clusterAgent.collection.kubeStateMetrics.labelsAsTags.pod.app"] = "pod_app"
					labels = map[string]map[string]string{
						"node": {"node-role.kubernetes.io/control-plane": "kube_node_role_control_plane", "team": "node_team"},
						"pod":  {"app": "pod_app"},
					}
				}
				if tc.annotations {
					values[`clusterAgent.collection.kubeStateMetrics.annotationsAsTags.node.example\.com/owner`] = "node_owner"
					values["clusterAgent.collection.kubeStateMetrics.annotationsAsTags.pod.owner"] = "pod_owner"
					values["clusterAgent.collection.kubeStateMetrics.annotationsAsTags.pod.team"] = "pod_team"
					annotations = map[string]map[string]string{
						"node": {"example.com/owner": "node_owner"},
						"pod":  {"owner": "pod_owner", "team": "pod_team"},
					}
				}

				resources := renderCompat(t, &helm.Options{
					ValuesFiles: []string{"values/minimal.yaml"},
					SetValues:   values,
				})
				conf, ok := clusterAgentKSMConf(t, resources)
				require.True(t, ok)
				var config struct {
					ClusterCheck bool `json:"cluster_check"`
					Instances    []struct {
						Collectors         []string                     `json:"collectors"`
						LabelsAsTags       map[string]map[string]string `json:"labels_as_tags"`
						AnnotationsAsTags  map[string]map[string]string `json:"annotations_as_tags"`
						SkipLeaderElection bool                         `json:"skip_leader_election"`
					} `json:"instances"`
				}
				require.NoError(t, yaml.Unmarshal([]byte(conf), &config))
				require.Len(t, config.Instances, 1)
				assert.Equal(t, labels, config.Instances[0].LabelsAsTags)
				assert.Equal(t, annotations, config.Instances[0].AnnotationsAsTags)
				assert.Contains(t, config.Instances[0].Collectors, "nodes")
				assert.Equal(t, clusterCheck, config.ClusterCheck)
				assert.Equal(t, clusterCheck, config.Instances[0].SkipLeaderElection)
			})
		}
	}
}
