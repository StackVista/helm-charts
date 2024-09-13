package test

import (
	"testing"
)

const expectedDocsConfig = `stackstate.webUIConfig.docLinkUrlPrefix = "https://docs.stackstate.io"`

func TestDocsConfigRenderingApi(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-k8s-api", []string{"values/docs_config.yaml"}, expectedDocsConfig)
}

func TestDocsConfigRenderingServer(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-k8s-server", []string{"values/docs_config.yaml", "values/split_disabled.yaml"}, expectedDocsConfig)
}
