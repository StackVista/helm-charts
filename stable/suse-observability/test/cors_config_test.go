package test

import (
	"testing"
)

const expectedCorsConfig = `stackstate.web.origins = [ "https://rancher.localhost" ]`

func TestCorsConfigRenderingApi(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/full.yaml"}, expectedCorsConfig)
}

func TestCorsConfigRenderingServer(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/full.yaml", "values/split_disabled.yaml"}, expectedCorsConfig)
}
