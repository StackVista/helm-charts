package stackstate_k8s_agent_dashboard

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func ParseValuesYaml() (StackStateK8sAgentValues, error) {
	filePath := "./stable/suse-observability-agent/values.yaml"
	valuesFileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Print("Error")
		return StackStateK8sAgentValues{}, err
	}

	var parsedYamlData StackStateK8sAgentValues
	err = yaml.Unmarshal(valuesFileContent, &parsedYamlData)
	if err != nil {
		fmt.Print("Error")
		return StackStateK8sAgentValues{}, err
	}

	return parsedYamlData, nil
}
