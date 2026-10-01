package test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLogsAgentOtelBinaryStartup(t *testing.T) {
	binary := os.Getenv("OTEL_AGENT_BINARY")
	if binary == "" {
		t.Skip("set OTEL_AGENT_BINARY to test rendered configuration startup")
	}
	binary, err := exec.LookPath(binary)
	require.NoError(t, err)
	binary, err = filepath.Abs(binary)
	require.NoError(t, err)
	for _, tc := range []struct {
		name, invalidQueue string
		fixed              bool
	}{
		{name: "fixed-promtail", fixed: true},
		{name: "reject-fixed-promtail-queue", fixed: true, invalidQueue: "stsk8slogs/promtail"},
		{name: "discovery-promtail"},
		{name: "reject-promtail-queue", invalidQueue: "stsk8slogs/promtail"},
		{name: "reject-native-queue", invalidQueue: "otlp_http/native"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var exported, discovered, unexpected atomic.Int64
			receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tc.fixed && r.URL.Path == "/stsAgent/features" {
					assert.Equal(t, "ApiKey synthetic-chart-startup", r.Header.Get("Authorization"))
					discovered.Add(1)
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"otel-logs":false}`)
					return
				}
				if r.URL.Path != "/stsAgent/logs/k8s" {
					unexpected.Add(1)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				assert.Equal(t, "synthetic-chart-startup", r.Header.Get("sts-api-key"))
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.NotEmpty(t, body)
				exported.Add(1)
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(receiver.Close)
			resources := renderLogsOtel(t, map[string]string{"stackstate.url": receiver.URL + "/stsAgent"})
			config := logsOtelConfig(t, resources)
			dir := t.TempDir()
			logsConfigMap(t, config, "extensions", "file_storage/logs")["directory"] = filepath.Join(dir, "checkpoints")
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			health := listener.Addr().String()
			require.NoError(t, listener.Close())
			controller := logsConfigMap(t, config, "extensions", "stslogsagent/logs")
			controller["health_endpoint"] = health
			if tc.fixed {
				logsConfigMap(t, config, "extensions")["stslogsagent/logs"] = map[string]interface{}{"health_endpoint": health}
				delete(logsConfigMap(t, config, "extensions"), "bearertokenauth/native")
				delivery := logsConfigMap(t, config, "connectors", "stslogsroute/logs")
				delete(delivery, "promtail_pipeline")
				delete(delivery, "native_pipeline")
				logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["delivery"] = delivery
				delete(config, "connectors")
				delete(logsConfigMap(t, config, "service", "pipelines"), "logs/promtail")
				logsConfigMap(t, config, "service", "pipelines", "logs/input")["exporters"] = []string{"stsk8slogs/promtail"}
				delete(logsConfigMap(t, config, "exporters"), "otlp_http/native")
				delete(logsConfigMap(t, config, "service", "pipelines"), "logs/native")
				logsConfigMap(t, config, "service")["extensions"] = []string{"file_storage/logs", "stslogsagent/logs"}
			} else {
				controller["state_directory"] = filepath.Join(dir, "controller")
				controller["termination_message_path"] = filepath.Join(dir, "termination.log")
			}
			filelog := logsConfigMap(t, config, "receivers", "file_log/pods")
			filelog["include"] = []string{filepath.Join(dir, "pods", "*", "*", "*.log")}
			filelog["exclude"] = []string{}
			// Filelog supplies filename metadata without a Kubernetes API in this process fixture.
			delete(logsConfigMap(t, config, "processors"), "k8s_attributes")
			logsConfigMap(t, config, "service", "pipelines", "logs/input")["processors"] = []string{"memory_limiter", "transform/static_pod", "transform/cluster"}
			logsConfigMap(t, config, "service", "telemetry")["metrics"] = map[string]interface{}{"level": "none"}
			if tc.invalidQueue != "" {
				logsConfigMap(t, config, "exporters", tc.invalidQueue, "sending_queue")["enabled"] = true
			}
			data, err := yaml.Marshal(config)
			require.NoError(t, err)
			configPath := filepath.Join(dir, "agent.yaml")
			require.NoError(t, os.WriteFile(configPath, data, 0600))
			podDir := filepath.Join(dir, "pods", "fixture_test_12345678-1234-1234-1234-123456789abc", "app")
			require.NoError(t, os.MkdirAll(podDir, 0755))
			record := time.Now().UTC().Format(time.RFC3339Nano) + " stdout F synthetic-chart-startup\n"
			require.NoError(t, os.WriteFile(filepath.Join(podDir, "0.log"), []byte(record), 0600))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "--config="+configPath, "--feature-gates=stanza.synchronousLogEmitter")
			cmd.Env = []string{"PATH=/usr/bin:/bin", "STS_API_KEY=synthetic-chart-startup", "RECEIVER_URL=" + receiver.URL + "/stsAgent", "NATIVE_OTLP_URL=" + receiver.URL + "/native", "PROMTAIL_LOGS_URL=" + receiver.URL + "/stsAgent/logs/k8s", "CLUSTER_NAME=some-k8s-cluster", "PROXY_URL=", "POD_NAME=agent", "POD_NAMESPACE=fixture", "K8S_NODE_NAME=node"}
			logPath := filepath.Join(dir, "collector.log")
			logFile, err := os.Create(logPath)
			require.NoError(t, err)
			defer logFile.Close()
			cmd.Stdout, cmd.Stderr = logFile, logFile
			require.NoError(t, cmd.Start())
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			t.Cleanup(func() {
				cancel()
				if t.Failed() {
					output, _ := os.ReadFile(logPath)
					t.Logf("collector: %s", output)
				}
			})
			if tc.invalidQueue != "" {
				select {
				case err := <-done:
					require.Error(t, err)
					output, err := os.ReadFile(logPath)
					require.NoError(t, err)
					assert.Contains(t, string(output), "sending_queue")
				case <-ctx.Done():
					t.Fatal("enabled queue was not rejected during startup")
				}
				assert.Zero(t, exported.Load())
			} else {
				client := &http.Client{Timeout: time.Second}
				require.Eventually(t, func() bool {
					for _, path := range []string{"/live", "/ready"} {
						resp, err := client.Get("http://" + health + path)
						if err != nil {
							return false
						}
						_ = resp.Body.Close()
						if resp.StatusCode != http.StatusOK {
							return false
						}
					}
					return exported.Load() > 0
				}, 20*time.Second, 100*time.Millisecond, "selected graph must become ready and export")
				require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))
				select {
				case err := <-done:
					require.NoError(t, err)
				case <-ctx.Done():
					t.Fatal("collector did not drain after SIGTERM")
				}
			}
			assert.Zero(t, unexpected.Load(), "no unexpected or native requests")
			if tc.fixed {
				assert.Zero(t, discovered.Load(), "fixed graph must not query features")
			} else if tc.invalidQueue == "" {
				assert.Positive(t, discovered.Load(), "B chart explicitly enables authenticated discovery")
			}
		})
	}
}
