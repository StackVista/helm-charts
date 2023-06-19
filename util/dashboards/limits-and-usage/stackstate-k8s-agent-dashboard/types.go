package stackstate_k8s_agent_dashboard

type StackStateK8sAgentValues struct {
	ChecksAgent  ChecksAgent  `yaml:"checksAgent"`
	ClusterAgent ClusterAgent `yaml:"clusterAgent"`
	LogsAgent    LogsAgent    `yaml:"logsAgent"`
	NodeAgent    NodeAgent    `yaml:"nodeAgent"`
}

type NodeAgent struct {
	Containers Containers `yaml:"containers"`
}

type Containers struct {
	Agent        Agent `yaml:"agent"`
	ProcessAgent Agent `yaml:"processAgent"`
}

type Agent struct {
	Resources Resources `yaml:"resources"`
}

type ProcessAgent struct {
	Resources Resources `yaml:"resources"`
}

type LogsAgent struct {
	Resources Resources `yaml:"resources"`
}

type ClusterAgent struct {
	Resources Resources `yaml:"resources"`
}

type ChecksAgent struct {
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	Limits   Limits `yaml:"limits"`
	Requests Limits `yaml:"requests"`
}

type Limits struct {
	Cpu    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}
