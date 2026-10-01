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
	"fmt"
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// This opt-in test runs the supplied image; Kubernetes metadata is replaced by
// Filelog's filename metadata so transport checks need no cluster credentials.
func TestLogsAgentOtelImageTrust(t *testing.T) {
	image := os.Getenv("OTEL_AGENT_IMAGE")
	if image == "" {
		t.Skip("set OTEL_AGENT_IMAGE to test rendered transports inside the collector image")
	}
	bundleCtx, bundleCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer bundleCancel()
	bundleContainer := fmt.Sprintf("otel-chart-roots-%d", time.Now().UnixNano())
	t.Cleanup(func() { logsImageRemoveContainer(t, bundleContainer) })
	systemBundle, err := exec.CommandContext(bundleCtx, "docker", "run", "--rm", "--pull=never",
		"--name", bundleContainer, "--entrypoint=cat", image, "/etc/ssl/ca-bundle.pem").Output()
	require.NoError(t, err)
	for _, transport := range []string{"promtail", "http", "grpc"} {
		for _, layout := range []string{"inline", "external", "system-discovery", "system-export", "proxy"} {
			t.Run(transport+"/"+layout, func(t *testing.T) {
				var discovered, exported, discoveryProxied, exportProxied atomic.Int64
				discoveryCert, discoveryPEM := logsImageCertificate(t, "discovery")
				exportCert, exportPEM := logsImageCertificate(t, "export")
				discovery := logsImageTLSServer(t, discoveryCert, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/stsAgent/features", r.URL.Path)
					assert.Equal(t, "ApiKey synthetic-chart-trust", r.Header.Get("Authorization"))
					discovered.Add(1)
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprintf(w, `{"otel-logs":%t}`, transport != "promtail")
				}))
				destination := logsImageTLSServer(t, exportCert, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if transport == "promtail" {
						assert.Equal(t, "synthetic-chart-trust", r.Header.Get("sts-api-key"))
					} else {
						assert.Equal(t, "SUSEObservability synthetic-chart-trust", r.Header.Get("Authorization"))
					}
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.NotEmpty(t, body)
					exported.Add(1)
					switch transport {
					case "grpc":
						assert.Equal(t, "/opentelemetry.proto.collector.logs.v1.LogsService/Export", r.URL.Path)
						assert.Equal(t, 2, r.ProtoMajor)
						w.Header().Set("Content-Type", "application/grpc")
						w.Header().Set("Trailer", "Grpc-Status")
						_, _ = w.Write([]byte{0, 0, 0, 0, 0})
						w.Header().Set("Grpc-Status", "0")
					case "http":
						assert.Equal(t, "/v1/logs", r.URL.Path)
						w.Header().Set("Content-Type", "application/x-protobuf")
					default:
						assert.Equal(t, "/stsAgent/logs/k8s", r.URL.Path)
						w.WriteHeader(http.StatusNoContent)
					}
				}))
				values := map[string]string{
					"stackstate.url":                    discovery.URL + "/stsAgent",
					"global.customCertificates.enabled": "true",
				}
				if transport == "grpc" {
					endpoint := strings.TrimPrefix(destination.URL, "https://")
					if layout == "proxy" {
						_, port, err := net.SplitHostPort(endpoint)
						require.NoError(t, err)
						endpoint = net.JoinHostPort("192.0.2.1", port)
					}
					values["otel.platformGrpcOtlpEndpoint"] = endpoint
				} else {
					values["otel.platformHttpOtlpEndpoint"] = destination.URL
				}
				if layout == "proxy" {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodConnect {
							w.WriteHeader(http.StatusMethodNotAllowed)
							return
						}
						targetAddress := r.Host
						if transport == "grpc" && r.Host == values["otel.platformGrpcOtlpEndpoint"] {
							targetAddress = strings.TrimPrefix(destination.URL, "https://")
						}
						target, err := net.DialTimeout("tcp", targetAddress, 5*time.Second)
						if err != nil {
							w.WriteHeader(http.StatusBadGateway)
							return
						}
						client, _, err := w.(http.Hijacker).Hijack()
						if err != nil {
							_ = target.Close()
							return
						}
						if r.Host == strings.TrimPrefix(discovery.URL, "https://") {
							discoveryProxied.Add(1)
						} else {
							exportProxied.Add(1)
						}
						_, _ = client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
						go func() { _, _ = io.Copy(target, client); _ = target.Close() }()
						_, _ = io.Copy(client, target)
						_ = client.Close()
					}))
					t.Cleanup(server.Close)
					values["global.proxy.url"] = server.URL
				}
				if layout == "inline" {
					values["global.customCertificates.pemData"] = string(discoveryPEM) + string(exportPEM)
				} else {
					values["global.customCertificates.configMapName"] = "transport-certificates"
				}
				resources := renderLogsOtel(t, values)
				config := logsOtelConfig(t, resources)
				dir := t.TempDir()
				certDir := filepath.Join(dir, "certs")
				require.NoError(t, os.Mkdir(certDir, 0755))
				if layout == "inline" {
					found := false
					for _, cm := range resources.ConfigMaps {
						for name, content := range cm.Data {
							if strings.Contains(content, string(discoveryPEM)) && strings.Contains(content, string(exportPEM)) {
								require.NoError(t, os.WriteFile(filepath.Join(certDir, name), []byte(content), 0644))
								found = true
							}
						}
					}
					require.True(t, found, "rendered inline certificate ConfigMap")
				} else {
					if layout == "system-discovery" {
						require.NoError(t, os.WriteFile(filepath.Join(dir, "system.pem"), append(append([]byte{}, systemBundle...), discoveryPEM...), 0644))
					} else {
						require.NoError(t, os.WriteFile(filepath.Join(certDir, "receiver.crt"), discoveryPEM, 0644))
					}
					if layout == "system-export" {
						require.NoError(t, os.WriteFile(filepath.Join(dir, "system.pem"), append(append([]byte{}, systemBundle...), exportPEM...), 0644))
					} else {
						require.NoError(t, os.WriteFile(filepath.Join(certDir, "destination.pem"), exportPEM, 0644))
					}
				}
				require.NotContains(t, resources.ConfigMaps[otelLogsAgentName].Data["otel-logs.yaml"], "ca_file:")
				container := logsContainer(t, resources, otelLogsAgentName)
				mounted := false
				for _, mount := range container.VolumeMounts {
					if mount.Name == "custom-certificates" {
						require.Equal(t, "/etc/pki/tls/certs", mount.MountPath)
						require.True(t, mount.ReadOnly)
						mounted = true
					}
				}
				require.True(t, mounted)
				logsConfigMap(t, config, "extensions", "file_storage/logs")["directory"] = "/fixture/checkpoints"
				capability := logsConfigMap(t, config, "extensions", "stslogsagent/logs")
				capability["state_directory"] = "/fixture/controller"
				capability["termination_message_path"] = "/fixture/termination.log"
				capability["health_endpoint"] = "127.0.0.1:0"
				capability["receiver_url"] = discovery.URL + "/stsAgent"
				filelog := logsConfigMap(t, config, "receivers", "file_log/pods")
				filelog["include"] = []string{"/fixture/pods/*/*/*.log"}
				filelog["exclude"] = []string{}
				delete(logsConfigMap(t, config, "processors"), "k8s_attributes")
				logsConfigMap(t, config, "service", "pipelines", "logs/input")["processors"] =
					[]string{"memory_limiter", "transform/static_pod", "transform/cluster"}
				logsConfigMap(t, config, "service", "telemetry")["metrics"] = map[string]interface{}{"level": "none"}
				logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["endpoint"] = destination.URL + "/stsAgent/logs/k8s"
				data, err := yaml.Marshal(config)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filepath.Join(dir, "agent.yaml"), data, 0644))
				podDir := filepath.Join(dir, "pods", "fixture_test_12345678-1234-1234-1234-123456789abc", "app")
				require.NoError(t, os.MkdirAll(podDir, 0755))
				record := time.Now().UTC().Format(time.RFC3339Nano) + " stdout F synthetic-chart-trust-record\n"
				require.NoError(t, os.WriteFile(filepath.Join(podDir, "0.log"), []byte(record), 0644))
				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				t.Cleanup(cancel)
				name := fmt.Sprintf("otel-chart-trust-%d", time.Now().UnixNano())
				args := []string{"run", "--rm", "--pull=never", "--name", name, "--network=host",
					"--user=" + fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), "--security-opt=label=disable",
					"-v", dir + ":/fixture", "-v", certDir + ":/etc/pki/tls/certs:ro"}
				if strings.HasPrefix(layout, "system-") {
					args = append(args, "-v", filepath.Join(dir, "system.pem")+":/etc/ssl/ca-bundle.pem:ro")
				}
				for _, variable := range container.Env {
					value := variable.Value
					if variable.ValueFrom != nil {
						value = "synthetic-chart-trust"
					}
					args = append(args, "-e", variable.Name+"="+value)
				}
				args = append(args, image, "--config=/fixture/agent.yaml", "--feature-gates=stanza.synchronousLogEmitter")
				logFile, err := os.Create(filepath.Join(dir, "collector.log"))
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, logFile.Close()) })
				cmd := exec.CommandContext(ctx, "docker", args...)
				cmd.Stdout, cmd.Stderr = logFile, logFile
				require.NoError(t, cmd.Start())
				t.Cleanup(func() {
					stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer stopCancel()
					if err := exec.CommandContext(stopCtx, "docker", "stop", "--time=5", name).Run(); err != nil {
						logsImageRemoveContainer(t, name)
					}
					cancel()
					_ = cmd.Wait()
					if t.Failed() {
						output, _ := os.ReadFile(filepath.Join(dir, "collector.log"))
						t.Logf("collector: %s", output)
					}
				})
				require.Eventually(t, func() bool {
					return discovered.Load() > 0 && exported.Load() > 0
				}, 30*time.Second, 100*time.Millisecond, "TLS discovery and selected export must both succeed")
				if layout == "proxy" {
					assert.Positive(t, discoveryProxied.Load(), "proxy must carry discovery")
					assert.Positive(t, exportProxied.Load(), "proxy must carry selected export")
				}
				// The shared mount must not hide the image's system trust bundle.
				bundle, err := exec.CommandContext(ctx, "docker", "exec", name, "cat", "/etc/ssl/ca-bundle.pem").Output()
				require.NoError(t, err)
				require.True(t, strings.HasPrefix(string(bundle), string(systemBundle)))
				pool := x509.NewCertPool()
				require.True(t, pool.AppendCertsFromPEM(bundle))
				require.Greater(t, len(pool.Subjects()), 50)
			})
		}
	}
}

func logsImageRemoveContainer(t *testing.T, name string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "docker", "rm", "--force", name).CombinedOutput()
	if err != nil && !strings.Contains(string(output), "No such container") {
		t.Errorf("remove test container %s: %v: %s", name, err, output)
	}
}

func logsImageCertificate(t *testing.T, name string) (tls.Certificate, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:        true, BasicConstraintsValid: true,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("192.0.2.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key},
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func logsImageTLSServer(t *testing.T, certificate tls.Certificate, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewUnstartedServer(handler)
	server.EnableHTTP2 = true
	server.TLS = &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12}
	server.StartTLS()
	t.Cleanup(server.Close)
	return server
}
