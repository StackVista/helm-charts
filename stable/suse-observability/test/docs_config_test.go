package test

import (
	"testing"
)

const expectedDocsConfig = `stackstate.webUIConfig.docLinkUrlPrefix = "https://docs.stackstate.io"`

func TestDocsConfigRenderingApi(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/docs_config.yaml"}, expectedDocsConfig)
}

func TestDocsConfigRenderingServer(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/docs_config.yaml", "values/split_disabled.yaml"}, expectedDocsConfig)
}
