package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

type emailTestCase struct {
	name           string
	enabled        bool
	externalSecret string
	username       string
	password       string
	expectedSecret string
	expectSecret   bool
	expectEmailRef bool
}

func TestEmailSecretConfiguration(t *testing.T) {
	testCases := []emailTestCase{
		{
			name:           "enabled internal credentials",
			enabled:        true,
			username:       "fixture-user",
			password:       "fixture-password",
			expectedSecret: "suse-observability-email",
			expectSecret:   true,
			expectEmailRef: true,
		},
		{
			name:           "enabled external secret",
			enabled:        true,
			externalSecret: "email-credentials",
			expectedSecret: "email-credentials",
			expectEmailRef: true,
		},
		{
			name:           "disabled email",
			expectedSecret: "suse-observability-email",
		},
		{
			name:           "external secret takes precedence over inline credentials",
			enabled:        true,
			externalSecret: "email-credentials",
			username:       "inline-user",
			password:       "inline-password",
			expectedSecret: "email-credentials",
			expectEmailRef: true,
		},
	}

	modes := []struct {
		name        string
		splitValue  string
		deployments []string
	}{
		{
			name:        "split",
			splitValue:  "true",
			deployments: []string{"suse-observability-api", "suse-observability-notification"},
		},
		{
			name:        "monolithic",
			splitValue:  "false",
			deployments: []string{"suse-observability-server"},
		},
	}

	for _, mode := range modes {
		for _, testCase := range testCases {
			t.Run(mode.name+"/"+testCase.name, func(t *testing.T) {
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
					ValuesFiles: []string{"values/full.yaml"},
					SetValues: map[string]string{
						"stackstate.features.server.split":                mode.splitValue,
						"stackstate.email.enabled":                        boolValue(testCase.enabled),
						"stackstate.email.server.auth.fromExternalSecret": testCase.externalSecret,
						"stackstate.email.server.auth.username":           testCase.username,
						"stackstate.email.server.auth.password":           testCase.password,
					},
				})
				resources := helmtestutil.NewKubernetesResources(t, output)

				if testCase.expectSecret {
					secret, ok := resources.Secrets[testCase.expectedSecret]
					require.True(t, ok, "email secret should exist")
					assert.Equal(t, testCase.username, string(secret.Data["SMTP_USER_NAME"]))
					assert.Equal(t, testCase.password, string(secret.Data["SMTP_PASSWORD"]))
				} else {
					_, ok := resources.Secrets["suse-observability-email"]
					assert.False(t, ok, "email secret should not be generated")
				}

				for _, deploymentName := range mode.deployments {
					deployment, ok := resources.Deployments[deploymentName]
					require.True(t, ok, "deployment should exist")
					container := deployment.Spec.Template.Spec.Containers[0]
					emailRefs := 0
					for _, envFrom := range container.EnvFrom {
						if envFrom.SecretRef != nil && envFrom.SecretRef.Name == testCase.expectedSecret {
							emailRefs++
						}
					}
					if testCase.expectEmailRef {
						assert.Equal(t, 1, emailRefs, "deployment should reference the email secret")
					} else {
						assert.Zero(t, emailRefs, "deployment should not reference the email secret")
					}
				}
			})
		}
	}
}

func boolValue(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
