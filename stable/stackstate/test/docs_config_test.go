package test

import (
	"testing"
)

const expectedDocsConfig = `stackstate.webUIConfig.docLinkUrlPrefix = "https://docs.stackstate.io"`

func TestDocsConfigRenderingApi(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{"values/docs_config.yaml"}, expectedDocsConfig)
}

func TestDocsConfigRenderingServer(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/docs_config.yaml", "values/split_disabled.yaml"}, expectedDocsConfig)
}
