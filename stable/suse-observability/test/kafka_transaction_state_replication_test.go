package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

type kafkaReassignment struct {
	Version    int `json:"version"`
	Partitions []struct {
		Topic     string `json:"topic"`
		Partition int    `json:"partition"`
		Replicas  []int  `json:"replicas"`
	} `json:"partitions"`
}

func TestKafkaTopicCreateMigratesTransactionStateReplication(t *testing.T) {
	mockBin := t.TempDir()
	reassignmentCapture := filepath.Join(t.TempDir(), "reassignment.json")
	reassignmentState := filepath.Join(t.TempDir(), "reassignment-started")
	commandLog := filepath.Join(t.TempDir(), "commands.log")

	writeKafkaMock(t, mockBin, "kafka-topics.sh", `#!/bin/bash
printf 'topics %s\n' "$*" >> "$COMMAND_LOG"
if [[ "$*" == *"--topic __transaction_state --describe"* ]]; then
  printf '%s\n' \
    'Topic: __transaction_state Partition: 0 Leader: 0 Replicas: 0 Isr: 0' \
    'Topic: __transaction_state Partition: 1 Leader: 1 Replicas: 1 Isr: 1' \
    'Topic: __transaction_state Partition: 2 Leader: 2 Replicas: 2 Isr: 2'
else
  printf 'Topic: existing Partition: 0 Leader: 0 Replicas: 0,1 Isr: 0,1\n'
fi
`)
	writeKafkaMock(t, mockBin, "kafka-broker-api-versions.sh", `#!/bin/bash
printf 'brokers %s\n' "$*" >> "$COMMAND_LOG"
printf '%s\n' \
  'kafka-0:9092 (id: 0 rack: null) -> (' \
  'kafka-1:9092 (id: 1 rack: null) -> (' \
  'kafka-2:9092 (id: 2 rack: null) -> ('
`)
	writeKafkaMock(t, mockBin, "kafka-reassign-partitions.sh", `#!/bin/bash
printf 'reassign %s\n' "$*" >> "$COMMAND_LOG"
verify=false
for argument in "$@"; do
  if [[ "$argument" == "--verify" ]]; then
    verify=true
  fi
done
while (( $# > 0 )); do
  if [[ "$1" == "--reassignment-json-file" ]]; then
    cp "$2" "$REASSIGNMENT_CAPTURE"
    shift 2
  else
    shift
  fi
done
if [[ "$verify" == true ]]; then
  if [[ -f "$REASSIGNMENT_STATE" ]]; then
    printf 'Reassignment of partition __transaction_state-0 is completed.\n'
  else
    printf 'There is no active reassignment of partition __transaction_state-0, but replica set is 0 rather than 0,1.\n'
  fi
else
  touch "$REASSIGNMENT_STATE"
fi
`)
	writeKafkaMock(t, mockBin, "kafka-configs.sh", "#!/bin/bash\nprintf 'configs %s\\n' \"$*\" >> \"$COMMAND_LOG\"\n")

	command := exec.Command("bash", "../scripts/job-kafka-topic-create.sh")
	command.Env = append(os.Environ(),
		"PATH="+mockBin+":"+os.Getenv("PATH"),
		"KAFKA_BROKERS=kafka:9092",
		"KAFKA_REPLICAS=3",
		"KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR=2",
		"REASSIGNMENT_CAPTURE="+reassignmentCapture,
		"REASSIGNMENT_STATE="+reassignmentState,
		"COMMAND_LOG="+commandLog,
	)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	content, err := os.ReadFile(reassignmentCapture)
	require.NoError(t, err)
	var reassignment kafkaReassignment
	require.NoError(t, json.Unmarshal(content, &reassignment))
	assert.Equal(t, 1, reassignment.Version)
	require.Len(t, reassignment.Partitions, 3)
	assert.Equal(t, "__transaction_state", reassignment.Partitions[0].Topic)
	assert.Equal(t, 0, reassignment.Partitions[0].Partition)
	assert.Equal(t, []int{0, 1}, reassignment.Partitions[0].Replicas)
	assert.Equal(t, "__transaction_state", reassignment.Partitions[1].Topic)
	assert.Equal(t, 1, reassignment.Partitions[1].Partition)
	assert.Equal(t, []int{1, 2}, reassignment.Partitions[1].Replicas)
	assert.Equal(t, "__transaction_state", reassignment.Partitions[2].Topic)
	assert.Equal(t, 2, reassignment.Partitions[2].Partition)
	assert.Equal(t, []int{2, 0}, reassignment.Partitions[2].Replicas)

	logContent, err := os.ReadFile(commandLog)
	require.NoError(t, err)
	commandLogText := string(logContent)
	assert.Contains(t, commandLogText, "--execute")
	assert.NotContains(t, commandLogText, "--throttle")
	assert.Contains(t, commandLogText, "--verify --preserve-throttles")
	require.Contains(t, commandLogText, "configs ")
	require.Contains(t, commandLogText, "brokers ")
	assert.Less(t, strings.LastIndex(commandLogText, "configs "), strings.Index(commandLogText, "brokers "))
}

func TestKafkaTopicCreateDefersMigrationWithoutFailingTopicReconciliation(t *testing.T) {
	mockBin := t.TempDir()
	commandLog := filepath.Join(t.TempDir(), "commands.log")

	writeKafkaMock(t, mockBin, "kafka-topics.sh", `#!/bin/bash
if [[ "$*" == *"--topic __transaction_state --describe"* ]]; then
  printf 'Topic: __transaction_state Partition: 0 Leader: 0 Replicas: 0 Isr: 0\n'
else
  printf 'Topic: existing Partition: 0 Leader: 0 Replicas: 0,1 Isr: 0,1\n'
fi
`)
	writeKafkaMock(t, mockBin, "kafka-configs.sh", "#!/bin/bash\nprintf 'configured\\n' >> \"$COMMAND_LOG\"\n")
	writeKafkaMock(t, mockBin, "kafka-broker-api-versions.sh", "#!/bin/bash\nprintf 'attempt\\n' >> \"$COMMAND_LOG\"\nprintf 'broker discovery failed\\n' >&2\nexit 1\n")
	writeKafkaMock(t, mockBin, "kafka-reassign-partitions.sh", "#!/bin/bash\nexit 1\n")
	writeKafkaMock(t, mockBin, "sleep", "#!/bin/bash\nexit 0\n")

	command := exec.Command("bash", "../scripts/job-kafka-topic-create.sh")
	command.Env = append(os.Environ(),
		"PATH="+mockBin+":"+os.Getenv("PATH"),
		"KAFKA_BROKERS=kafka:9092",
		"KAFKA_REPLICAS=3",
		"KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR=2",
		"KAFKA_TRANSACTION_STATE_REASSIGNMENT_ATTEMPTS=2",
		"KAFKA_TRANSACTION_STATE_REASSIGNMENT_RETRY_INTERVAL=0",
		"COMMAND_LOG="+commandLog,
	)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	assert.Contains(t, string(output), "the next chart run will retry it")

	logContent, err := os.ReadFile(commandLog)
	require.NoError(t, err)
	assert.Contains(t, string(logContent), "configured")
	assert.Equal(t, 2, strings.Count(string(logContent), "attempt\n"))
}

func TestKafkaTopicCreateStopsRetryingAfterSharedDeadline(t *testing.T) {
	mockBin := t.TempDir()
	reassignmentState := filepath.Join(t.TempDir(), "reassignment-started")
	commandLog := filepath.Join(t.TempDir(), "commands.log")

	writeKafkaMock(t, mockBin, "kafka-topics.sh", `#!/bin/bash
if [[ "$*" == *"--topic __transaction_state --describe"* ]]; then
  printf 'Topic: __transaction_state Partition: 0 Leader: 0 Replicas: 0 Isr: 0\n'
else
  printf 'Topic: existing Partition: 0 Leader: 0 Replicas: 0,1 Isr: 0,1\n'
fi
`)
	writeKafkaMock(t, mockBin, "kafka-configs.sh", "#!/bin/bash\nprintf 'configs %s\\n' \"$*\" >> \"$COMMAND_LOG\"\n")
	writeKafkaMock(t, mockBin, "kafka-broker-api-versions.sh", `#!/bin/bash
printf 'brokers\n' >> "$COMMAND_LOG"
printf '%s\n' \
  'kafka-0:9092 (id: 0 rack: null) -> (' \
  'kafka-1:9092 (id: 1 rack: null) -> ('
`)
	writeKafkaMock(t, mockBin, "kafka-reassign-partitions.sh", `#!/bin/bash
printf 'reassign %s\n' "$*" >> "$COMMAND_LOG"
if [[ "$*" == *"--execute"* ]]; then
  touch "$REASSIGNMENT_STATE"
elif [[ -f "$REASSIGNMENT_STATE" ]]; then
  printf 'Reassignment of partition __transaction_state-0 is still in progress.\n'
else
  printf 'There is no active reassignment of partition __transaction_state-0, but replica set is 0 rather than 0,1.\n'
fi
`)
	writeKafkaMock(t, mockBin, "sleep", "#!/bin/bash\n/bin/sleep 0.1\n")

	commandEnvironment := append(os.Environ(),
		"PATH="+mockBin+":"+os.Getenv("PATH"),
		"KAFKA_BROKERS=kafka:9092",
		"KAFKA_REPLICAS=3",
		"KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR=2",
		"KAFKA_TRANSACTION_STATE_REASSIGNMENT_TIMEOUT=1",
		"REASSIGNMENT_STATE="+reassignmentState,
		"COMMAND_LOG="+commandLog,
	)

	firstRun := exec.Command("bash", "../scripts/job-kafka-topic-create.sh")
	firstRun.Env = commandEnvironment
	output, err := firstRun.CombinedOutput()
	require.NoError(t, err, string(output))
	assert.Contains(t, string(output), "Timed out waiting")
	assert.NotContains(t, string(output), "Retrying __transaction_state")

	logContent, err := os.ReadFile(commandLog)
	require.NoError(t, err)
	commandLogText := string(logContent)
	assert.Equal(t, 1, strings.Count(commandLogText, "--execute"))
	assert.Equal(t, 1, strings.Count(commandLogText, "brokers\n"))
	assert.NotContains(t, commandLogText, "--throttle")
}

func TestKafkaTopicCreateResumesActiveTransactionStateReassignment(t *testing.T) {
	mockBin := t.TempDir()
	verifyCount := filepath.Join(t.TempDir(), "verify-count")
	commandLog := filepath.Join(t.TempDir(), "commands.log")

	writeKafkaMock(t, mockBin, "kafka-topics.sh", `#!/bin/bash
if [[ "$*" == *"--topic __transaction_state --describe"* ]]; then
  printf 'Topic: __transaction_state Partition: 0 Leader: 0 Replicas: 0 Isr: 0\n'
else
  printf 'Topic: existing Partition: 0 Leader: 0 Replicas: 0,1 Isr: 0,1\n'
fi
`)
	writeKafkaMock(t, mockBin, "kafka-configs.sh", "#!/bin/bash\nexit 0\n")
	writeKafkaMock(t, mockBin, "kafka-broker-api-versions.sh", `#!/bin/bash
printf '%s\n' \
  'kafka-0:9092 (id: 0 rack: null) -> (' \
  'kafka-1:9092 (id: 1 rack: null) -> (' \
  'kafka-2:9092 (id: 2 rack: null) -> ('
`)
	writeKafkaMock(t, mockBin, "kafka-reassign-partitions.sh", `#!/bin/bash
if [[ "$*" == *"--execute"* ]]; then
  printf 'execute\n' >> "$COMMAND_LOG"
  exit 1
fi
if [[ ! -f "$VERIFY_COUNT" ]]; then
  touch "$VERIFY_COUNT"
  printf 'Reassignment of partition __transaction_state-0 is still in progress.\n'
else
  printf 'Reassignment of partition __transaction_state-0 is completed.\n'
fi
`)
	writeKafkaMock(t, mockBin, "sleep", "#!/bin/bash\nexit 0\n")

	command := exec.Command("bash", "../scripts/job-kafka-topic-create.sh")
	command.Env = append(os.Environ(),
		"PATH="+mockBin+":"+os.Getenv("PATH"),
		"KAFKA_BROKERS=kafka:9092",
		"KAFKA_REPLICAS=3",
		"KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR=2",
		"VERIFY_COUNT="+verifyCount,
		"COMMAND_LOG="+commandLog,
	)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	logContent, err := os.ReadFile(commandLog)
	if !os.IsNotExist(err) {
		require.NoError(t, err)
		assert.NotContains(t, string(logContent), "execute")
	}
}

func TestKafkaTopicCreateReplicationFactorByProfile(t *testing.T) {
	tests := []struct {
		name       string
		valuesFile string
		expected   string
	}{
		{name: "non-ha", valuesFile: "values/values_sizing_10_nonha.yaml", expected: "1"},
		{name: "ha", valuesFile: "values/values_sizing_150_ha.yaml", expected: "2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", test.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)
			job := findJob(&resources, "topic-create")
			require.NotNil(t, job)

			for _, env := range job.Spec.Template.Spec.Containers[0].Env {
				if env.Name == "KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR" {
					assert.Equal(t, test.expected, env.Value)
					return
				}
			}
			t.Fatal("KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR was not rendered")
		})
	}
}

func writeKafkaMock(t *testing.T, directory, name, content string) {
	t.Helper()
	path := filepath.Join(directory, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
}
