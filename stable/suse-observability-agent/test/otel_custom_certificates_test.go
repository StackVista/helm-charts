package test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

type otelTrustWorkload struct {
	name      string
	pod       corev1.PodSpec
	config    map[string]interface{}
	exporter  string
	mountPath string
}

func renderOtelTrust(t *testing.T, transport, source, certificates string) []otelTrustWorkload {
	t.Helper()
	values := map[string]string{
		"stackstate.url":                            "https://platform.example/receiver/stsAgent",
		"otel.enabled":                              "true",
		"otel.telemetryGateway.enabled":             "true",
		"global.features.experimentalOtelLogsAgent": "true",
		"otelLogsAgent.exportMode":                  "native",
	}
	if transport == "grpc" {
		values["otel.platformGrpcOtlpEndpoint"] = "platform.example:4317"
	} else {
		values["otel.platformHttpOtlpEndpoint"] = "https://platform.example"
	}
	if source != "disabled" {
		values["global.customCertificates.enabled"] = "true"
		if source == "inline" {
			values["global.customCertificates.pemData"] = certificates
		} else {
			values["global.customCertificates.configMapName"] = "external-ca-certificates"
		}
	}
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml"},
		SetValues:   values,
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	workloads := []otelTrustWorkload{
		{name: "k8s-resource-collector", pod: resources.Deployments["suse-observability-agent-k8s-resource-collector"].Spec.Template.Spec, mountPath: "/etc/ssl/certs/custom"},
		{name: "otel-metrics-scraper", pod: resources.Statefulsets["suse-observability-agent-otel-metrics-scraper"].Spec.Template.Spec, mountPath: "/etc/ssl/certs/custom"},
		{name: "otel-telemetry-gateway", pod: resources.Deployments[otelTelemetryGatewayName].Spec.Template.Spec, mountPath: "/etc/pki/tls/certs"},
		{name: "otel-logs-agent", pod: resources.DaemonSets[otelLogsAgentName].Spec.Template.Spec, mountPath: "/etc/pki/tls/certs"},
	}
	for i := range workloads {
		workload := &workloads[i]
		require.Len(t, workload.pod.Containers, 1, workload.name)
		configName, configKey := "suse-observability-agent-"+workload.name+"-config", "config.yaml"
		if workload.name == "otel-logs-agent" {
			configName, configKey = otelLogsAgentName, "otel-logs.yaml"
		}
		require.NoError(t, yaml.Unmarshal([]byte(resources.ConfigMaps[configName].Data[configKey]), &workload.config))
		for name := range logsConfigMap(t, workload.config, "exporters") {
			if strings.HasPrefix(name, "otlp") {
				require.Empty(t, workload.exporter, "only one platform exporter")
				workload.exporter = name
			}
		}
		require.NotEmpty(t, workload.exporter)
		if transport == "grpc" {
			assert.True(t, strings.HasPrefix(workload.exporter, "otlp/"))
		} else {
			assert.False(t, strings.HasPrefix(workload.exporter, "otlp/"))
		}
	}
	if source == "inline" {
		assert.Equal(t, strings.TrimSpace(certificates), strings.TrimSpace(resources.ConfigMaps["suse-observability-agent-custom-certificates"].Data["tls.pem"]))
	} else {
		assert.NotContains(t, resources.ConfigMaps, "suse-observability-agent-custom-certificates")
	}
	return workloads
}

func TestOtelCustomCertificatesRendering(t *testing.T) {
	for _, transport := range []string{"http", "grpc"} {
		for _, source := range []string{"disabled", "inline", "external"} {
			t.Run(transport+"/"+source, func(t *testing.T) {
				for _, workload := range renderOtelTrust(t, transport, source, "synthetic-ca\n") {
					t.Run(workload.name, func(t *testing.T) {
						container := workload.pod.Containers[0]
						env := map[string]string{}
						for _, entry := range container.Env {
							env[entry.Name] = entry.Value
						}
						assert.NotContains(t, env, "SSL_CERT_FILE")
						assert.NotEqual(t, "true", env["SKIP_SSL_VALIDATION"])
						exporterTLS := logsConfigMap(t, workload.config, "exporters", workload.exporter, "tls")
						assert.NotEqual(t, true, exporterTLS["insecure_skip_verify"])
						assert.NotEqual(t, true, exporterTLS["insecure"])
						if source == "disabled" {
							assert.NotContains(t, env, "SSL_CERT_DIR")
						} else {
							directories := strings.Split(env["SSL_CERT_DIR"], ":")
							assert.Contains(t, directories, "/etc/ssl/certs")
							assert.Contains(t, directories, "/etc/pki/tls/certs")
							assert.Contains(t, directories, workload.mountPath)
						}
						var volumes, mounts int
						for _, volume := range workload.pod.Volumes {
							if volume.Name == "custom-certificates" {
								volumes++
								require.NotNil(t, volume.ConfigMap)
								expected := "suse-observability-agent-custom-certificates"
								if source == "external" {
									expected = "external-ca-certificates"
								}
								assert.Equal(t, expected, volume.ConfigMap.Name)
							}
						}
						for _, mount := range container.VolumeMounts {
							if mount.Name == "custom-certificates" {
								mounts++
								assert.Equal(t, workload.mountPath, mount.MountPath)
								assert.True(t, mount.ReadOnly)
							}
						}
						expected := 1
						if source == "disabled" {
							expected = 0
						}
						assert.Equal(t, expected, volumes)
						assert.Equal(t, expected, mounts)
					})
				}
			})
		}
	}
}

func TestOtelCustomCertificatesExport(t *testing.T) {
	binary := os.Getenv("OTEL_AGENT_BINARY")
	if binary == "" {
		t.Skip("set OTEL_AGENT_BINARY to validate chart exporter trust with the shipped collector")
	}
	binary, err := exec.LookPath(binary)
	require.NoError(t, err)
	binary, err = filepath.Abs(binary)
	require.NoError(t, err)
	certificate, caPEM := otelTrustCertificate(t, "127.0.0.1")
	_, unrelatedPEM := otelTrustCertificate(t, "127.0.0.1")
	wrongHostCertificate, wrongHostPEM := otelTrustCertificate(t, "192.0.2.1")
	for _, transport := range []string{"http", "grpc"} {
		for _, source := range []string{"inline", "external", "untrusted", "hostname", "system", "disabled"} {
			t.Run(transport+"/"+source, func(t *testing.T) {
				serverCertificate, trustedPEM := certificate, caPEM
				if source == "hostname" {
					serverCertificate, trustedPEM = wrongHostCertificate, wrongHostPEM
				}
				for _, workload := range renderOtelTrust(t, transport, source, string(trustedPEM)) {
					t.Run(workload.name, func(t *testing.T) {
						runOtelTrustExporter(t, binary, transport, source, workload, serverCertificate, trustedPEM, unrelatedPEM)
					})
				}
			})
		}
	}
}

func runOtelTrustExporter(t *testing.T, binary, transport, source string, workload otelTrustWorkload, certificate tls.Certificate, caPEM, unrelatedPEM []byte) {
	t.Helper()
	var exported atomic.Int64
	destination := logsImageTLSServer(t, certificate, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "SUSEObservability foobar", r.Header.Get("Authorization"))
		data, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		exported.Add(1)
		if transport == "grpc" {
			assert.Equal(t, "/opentelemetry.proto.collector.metrics.v1.MetricsService/Export", r.URL.Path)
			assert.Equal(t, 2, r.ProtoMajor)
			w.Header().Set("Content-Type", "application/grpc")
			w.Header().Set("Trailer", "Grpc-Status")
			_, _ = w.Write([]byte{0, 0, 0, 0, 0})
			w.Header().Set("Grpc-Status", "0")
		} else {
			assert.Equal(t, "/v1/metrics", r.URL.Path)
			w.Header().Set("Content-Type", "application/x-protobuf")
		}
	}))
	metrics := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "# TYPE chart_tls_test gauge\nchart_tls_test 1\n")
	}))
	t.Cleanup(metrics.Close)
	exporter := logsConfigMap(t, workload.config, "exporters", workload.exporter)
	endpoint := destination.URL
	if transport == "grpc" {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}
	exporter["endpoint"] = endpoint
	exporter["retry_on_failure"] = map[string]interface{}{"enabled": false}
	exporter["sending_queue"] = map[string]interface{}{"enabled": false}
	auth := logsConfigMap(t, exporter, "auth")["authenticator"].(string)
	// Replace Kubernetes receivers while retaining the chart's exporter and authentication.
	config := map[string]interface{}{
		"receivers": map[string]interface{}{"prometheus": map[string]interface{}{
			"config": map[string]interface{}{"scrape_configs": []interface{}{map[string]interface{}{
				"job_name": "chart-tls-test", "scrape_interval": "200ms",
				"static_configs": []interface{}{map[string]interface{}{"targets": []string{strings.TrimPrefix(metrics.URL, "http://")}}},
			}}},
		}},
		"exporters":  map[string]interface{}{workload.exporter: exporter},
		"extensions": map[string]interface{}{auth: logsConfigMap(t, workload.config, "extensions", auth)},
		"service": map[string]interface{}{
			"extensions": []string{auth},
			"telemetry":  map[string]interface{}{"metrics": map[string]interface{}{"level": "none"}},
			"pipelines": map[string]interface{}{"metrics": map[string]interface{}{
				"receivers": []string{"prometheus"}, "exporters": []string{workload.exporter},
			}},
		},
	}
	dir := t.TempDir()
	customDir := filepath.Join(dir, "custom")
	require.NoError(t, os.Mkdir(customDir, 0755))
	filename := "tls.pem"
	if source == "external" {
		filename = "private-ca.crt"
	}
	customPEM := caPEM
	if source == "untrusted" {
		customPEM = unrelatedPEM
	}
	require.NoError(t, os.WriteFile(filepath.Join(customDir, filename), customPEM, 0600))
	if source == "external" {
		require.NoError(t, os.WriteFile(filepath.Join(customDir, "other-ca.pem"), unrelatedPEM, 0600))
	}
	// Relocate the rendered mount into a local process fixture; keep normal root directories.
	env := []string{"PATH=/usr/bin:/bin", "STS_API_KEY=foobar", "SKIP_SSL_VALIDATION=false", "PROXY_URL="}
	for _, entry := range workload.pod.Containers[0].Env {
		if entry.Name == "SSL_CERT_DIR" {
			env = append(env, "SSL_CERT_DIR="+strings.ReplaceAll(entry.Value, workload.mountPath, customDir))
		}
	}
	systemPEM := unrelatedPEM
	if source == "disabled" || source == "system" {
		systemPEM = caPEM
		require.NoError(t, os.WriteFile(filepath.Join(customDir, filename), unrelatedPEM, 0600))
	}
	systemBundle := filepath.Join(dir, "system.pem")
	require.NoError(t, os.WriteFile(systemBundle, systemPEM, 0600))
	env = append(env, "SSL_CERT_FILE="+systemBundle)
	data, err := yaml.Marshal(config)
	require.NoError(t, err)
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, data, 0600))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--config="+configPath)
	cmd.Env = env
	logPath := filepath.Join(dir, "collector.log")
	logFile, err := os.Create(logPath)
	require.NoError(t, err)
	defer logFile.Close()
	cmd.Stdout, cmd.Stderr = logFile, logFile
	require.NoError(t, cmd.Start())
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	defer func() {
		cancel()
		<-done
		if t.Failed() {
			output, _ := os.ReadFile(logPath)
			t.Logf("collector: %s", output)
		}
	}()
	if source == "untrusted" || source == "hostname" {
		expectedError := "certificate signed by unknown authority"
		if source == "hostname" {
			expectedError = "certificate is valid for 192.0.2.1, not 127.0.0.1"
		}
		require.Eventually(t, func() bool {
			output, _ := os.ReadFile(logPath)
			select {
			case <-done:
				return true
			default:
				return strings.Contains(string(output), expectedError)
			}
		}, 10*time.Second, 100*time.Millisecond)
		output, err := os.ReadFile(logPath)
		require.NoError(t, err)
		assert.Contains(t, string(output), expectedError)
		assert.Zero(t, exported.Load(), "failed certificate verification must prevent export")
	} else {
		require.Eventually(t, func() bool {
			select {
			case <-done:
				return true
			default:
				return exported.Load() > 0
			}
		}, 10*time.Second, 100*time.Millisecond)
		assert.Positive(t, exported.Load(), "collector must start and export using the chart's trust settings")
	}
}

func otelTrustCertificate(t *testing.T, serverIP string) (tls.Certificate, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "chart-test-ca"},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	require.NoError(t, err)
	ca, err = x509.ParseCertificate(caDER)
	require.NoError(t, err)
	leaf := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "chart-test-endpoint"},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP(serverIP)},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leaf, ca, &key.PublicKey, key)
	require.NoError(t, err)
	return tls.Certificate{Certificate: [][]byte{leafDER}, PrivateKey: key},
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
}
