package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	grafana "gitlab.com/StackVista/DevOps/helm-charts/util/dashboards/limits-and-usage/grafana"
	agent "gitlab.com/StackVista/DevOps/helm-charts/util/dashboards/limits-and-usage/stackstate-k8s-agent-dashboard"
)

// Global incremental counter to never let panel ids overlap
var panelIdCounts = 0

// LimitOrRequestType Types of request of limit mappings
type LimitOrRequestType int

const (
	CPU LimitOrRequestType = iota
	Memory
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphaNumericRunes[rand.Intn(len(alphaNumericRunes))]
	}
	return string(b)
}

var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// Base panel when adding a low level graph
var basePanel = grafana.Panel{
	Datasource: grafana.Datasource{
		Type: "prometheus",
		UID:  "PBFA97CFB590B2093",
	},
	Description: "",
	FieldConfig: grafana.FieldConfig{
		Defaults: grafana.Defaults{
			Color: grafana.Color{
				Mode: "thresholds",
			},
			Mappings: []string{},
			Thresholds: grafana.Thresholds{
				Mode: "absolute",
			},
		},
		Overrides: []string{},
	},
	Options: grafana.Options{
		ColorMode:   "value",
		GraphMode:   "area",
		JustifyMode: "auto",
		Orientation: "auto",
		ReduceOptions: grafana.ReduceOptions{
			Fields: "",
			Values: false,
		},
		Text:     grafana.Text{},
		TextMode: "auto",
	},
	PluginVersion: "8.5.13",
	Targets: []grafana.TargetAlt{
		{
			Datasource: grafana.Datasource{Type: "prometheus", UID: "PBFA97CFB590B2093"},
			EditorMode: "code",
			Exemplar:   false,
			Range:      true,
			RefID:      "A",
		},
	},
	Type: "stat",
}

func getRequestUsageAverage(id int, requestValue string, targetType LimitOrRequestType, namespace string, containerName string, gridPosition grafana.GridPos) grafana.Panel {
	var displayName string
	var steps []grafana.Step
	var expression string
	var measurementUnit string
	var measurementUnitShort string
	var requestValueFloat float64

	title := " / Average Usage - Requests"

	if targetType == CPU {
		title = "CPU" + title
		requestValueFloat, _ = strconv.ParseFloat(strings.Replace(requestValue, "m", "", -1), 32)
		expression = fmt.Sprintf("avg(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{namespace=\"%s\", container=\"%s\"}) * 1000", namespace, containerName)
		measurementUnit = "MilliCPU"
		measurementUnitShort = "Mi"
	} else if targetType == Memory {
		title = "Memory" + title
		requestValueFloat, _ = strconv.ParseFloat(strings.Replace(requestValue, "Mi", "", -1), 32)
		expression = fmt.Sprintf("avg(container_memory_working_set_bytes{namespace=\"%s\", container=\"%s\"}) / 1000000", namespace, containerName)
		measurementUnit = "mbytes"
		measurementUnitShort = "m"
	}

	// Make sure the limit value is not blank
	if requestValue == "" {
		displayName = "!!! NO REQUEST LIMIT SET !!!"
		steps = []grafana.Step{
			{Color: "dark-red"},
		}
	} else {
		stableAverage := requestValueFloat - (requestValueFloat * 0.25)
		displayName = fmt.Sprintf("Request %v%s. Recommended between %v%s and %v%s based on values file.", requestValueFloat, measurementUnitShort, stableAverage, measurementUnitShort, requestValueFloat, measurementUnitShort)
		steps = []grafana.Step{
			{Color: "dark-orange"},
			{Color: "green", Value: int(stableAverage)},
			// We add one to prevent it showing as bad if the requested amount is equal to the actual average
			{Color: "dark-red", Value: int(requestValueFloat) + 1},
		}
	}

	newPanel := basePanel
	newPanel.ID = id
	newPanel.FieldConfig.Defaults.DisplayName = displayName
	newPanel.FieldConfig.Defaults.Thresholds.Steps = steps
	newPanel.FieldConfig.Defaults.Unit = measurementUnit
	newPanel.GridPos = gridPosition
	newPanel.Targets = []grafana.TargetAlt{
		{
			Datasource: grafana.Datasource{Type: "prometheus", UID: "PBFA97CFB590B2093"},
			EditorMode: "code",
			Exemplar:   false,
			Expr:       expression,
			Range:      true,
			RefID:      "A",
		},
	}
	newPanel.Title = title
	newPanel.Options.ReduceOptions.Calcs = []string{
		"mean",
	}
	return newPanel
}

func getLimitUsageMax(id int, limitValue string, targetType LimitOrRequestType, namespace string, containerName string, gridPosition grafana.GridPos) grafana.Panel {
	var displayName string
	var steps []grafana.Step
	var expression string
	var measurementUnit string
	var measurementUnitShort string
	var limitValueFloat float64

	title := " / Max Usage - Limits"

	if targetType == CPU {
		title = "CPU" + title
		limitValueFloat, _ = strconv.ParseFloat(strings.Replace(limitValue, "m", "", -1), 32)
		expression = fmt.Sprintf("max(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{namespace=\"%s\", container=\"%s\"}) * 1000", namespace, containerName)
		measurementUnit = "MilliCPU"
		measurementUnitShort = "Mi"
	} else if targetType == Memory {
		title = "Memory" + title
		limitValueFloat, _ = strconv.ParseFloat(strings.Replace(limitValue, "Mi", "", -1), 32)
		expression = fmt.Sprintf("max(container_memory_working_set_bytes{namespace=\"%s\", container=\"%s\"}) / 1000000", namespace, containerName)
		measurementUnit = "mbytes"
		measurementUnitShort = "M"
	}

	stableAverage := limitValueFloat - (limitValueFloat * 0.25)
	stableAverageMax := limitValueFloat - 5

	// Make sure the limit value is not blank
	if limitValue == "" {
		displayName = "!!! NO LIMIT SET !!!"
		steps = []grafana.Step{
			{Color: "dark-red"},
		}
	} else {
		displayName = fmt.Sprintf("Limit %v%s. Recommended between %v%s and %v%s based on values file. Max limit should not be reached", limitValueFloat, measurementUnitShort, stableAverage, measurementUnitShort, stableAverageMax, measurementUnitShort)
		steps = []grafana.Step{
			{Color: "dark-orange"},
			{Color: "green", Value: int(stableAverage)},
			{Color: "dark-red", Value: int(stableAverageMax)},
		}
	}

	newPanel := basePanel
	newPanel.ID = id
	newPanel.FieldConfig.Defaults.DisplayName = displayName
	newPanel.FieldConfig.Defaults.Thresholds.Steps = steps
	newPanel.FieldConfig.Defaults.Unit = measurementUnit
	newPanel.GridPos = gridPosition
	newPanel.Targets = []grafana.TargetAlt{
		{
			Datasource: grafana.Datasource{Type: "prometheus", UID: "PBFA97CFB590B2093"},
			EditorMode: "code",
			Exemplar:   false,
			Expr:       expression,
			Range:      true,
			RefID:      "A",
		},
	}
	newPanel.Title = title
	newPanel.Options.ReduceOptions.Calcs = []string{
		"max",
	}
	return newPanel
}

func createGrafanaBlockPanel(id int, namespace string, containerTarget string, resources agent.Resources) grafana.Panel {
	panelIdCounts += 1
	cpuRequestPanel := getRequestUsageAverage(panelIdCounts, resources.Requests.Cpu, CPU, namespace, containerTarget, grafana.GridPos{
		H: 9,
		W: 12,
		X: 0,
		Y: 1,
	})

	panelIdCounts += 1
	cpuLimitPanel := getLimitUsageMax(panelIdCounts, resources.Limits.Cpu, CPU, namespace, containerTarget, grafana.GridPos{
		H: 9,
		W: 12,
		X: 12,
		Y: 1,
	})

	panelIdCounts += 1
	memoryRequestPanel := getRequestUsageAverage(panelIdCounts, resources.Requests.Memory, Memory, namespace, containerTarget, grafana.GridPos{
		H: 9,
		W: 12,
		X: 0,
		Y: 9,
	})

	panelIdCounts += 1
	memoryLimitPanel := getLimitUsageMax(panelIdCounts, resources.Limits.Memory, Memory, namespace, containerTarget, grafana.GridPos{
		H: 9,
		W: 12,
		X: 12,
		Y: 9,
	})

	return grafana.Panel{
		ID:        id,
		Collapsed: true,
		GridPos: grafana.GridPos{
			H: 1,
			W: 24,
			X: 0,
			Y: 0,
		},
		Panels: []grafana.Panel{
			cpuRequestPanel,
			cpuLimitPanel,
			memoryRequestPanel,
			memoryLimitPanel,
		},
		Title: fmt.Sprintf("%s - %s namespace", containerTarget, namespace),
		Type:  "row",
	}
}

func main() {
	fmt.Println("Enter the namespace you want to observe for the K8S Agent (default: monitoring): ")
	var namespace string
	unexNl := errors.New("unexpected newline")
	_, err := fmt.Scanln(&namespace)
	if err != nil {
		if err.Error() != unexNl.Error() {
			fmt.Println(err)
			fmt.Println(unexNl)
			fmt.Println("Unable to get the namespace from the user input.")
			return
		}
		fmt.Println("No value entered, using monitoring")
		namespace = "monitoring"
	}

	data, err := agent.ParseValuesYaml()
	if err != nil {
		fmt.Println("Error Occurred:", err)
		return
	}

	dashboard := grafana.Dashboard{
		Annotations: grafana.Annotations{
			List: []grafana.AnnotationItem{
				{
					BuiltIn:    1,
					Datasource: grafana.Datasource{Type: "grafana", UID: "-- Grafana --"},
					Enable:     true,
					Hide:       true,
					IconColor:  "rgba(0, 211, 255, 1)",
					Name:       "Annotations & Alerts",
					Target: grafana.Target{
						Limit:    100,
						MatchAny: false,
						Tags:     []string{},
						Type:     "dashboard",
					},
					Type: "dashboard",
				},
			},
		},
		Editable:             true,
		FiscalYearStartMonth: 0,
		GraphTooltip:         0,
		ID:                   78,
		Links:                []string{},
		LiveNow:              false,
		Panels: []grafana.Panel{
			createGrafanaBlockPanel(1, namespace, "suse-observability-agent", data.ChecksAgent.Resources),
			createGrafanaBlockPanel(2, namespace, "cluster-agent", data.ClusterAgent.Resources),
			createGrafanaBlockPanel(3, namespace, "logs-agent", data.LogsAgent.Resources),
			createGrafanaBlockPanel(4, namespace, "node-agent", data.NodeAgent.Containers.Agent.Resources),
			createGrafanaBlockPanel(5, namespace, "process-agent", data.NodeAgent.Containers.ProcessAgent.Resources),
		},
		SchemaVersion: 36,
		Style:         "dark",
		Tags:          []string{},
		Templating:    grafana.Templating{List: []string{}},
		Time: grafana.Time{
			From: "now-90d",
			To:   "now",
		},
		Timepicker: grafana.Timepicker{},
		Timezone:   "",
		Title:      "CPU & Mem (Request and Limits)",
		UID:        randSeq(9), //"Q4RgQel4k",
		Version:    9,
		WeekStart:  "",
	}

	indentedData, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = ioutil.WriteFile("util/dashboards/limits-and-usage/agent-usage-and-limits-dashboard.json", indentedData, 0644)
}
