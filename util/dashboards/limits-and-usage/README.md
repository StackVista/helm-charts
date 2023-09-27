# Limit and Usage Dashboards

The purpose of this script is to generate a grafana dashboard that shows the limits and usages of the stackstate k8s agent
based on the current values file. The dashboard will span over a total of 90 days and show one of the following
- green: Your set limit or request falls into the spectrum of the 90 days value (There is a 25% gap to test for the optimal amount)
- yellow: Your limit or request is far below the optimal 25% gap and should be modified to use less resources
- red: Your limit or request is either not set or is over the value you have defined in the values file.

The following script [create-dashboards.go](create-dashboards.go) will generate a `agent-usage-and-limits-dashboard.json` file that
can be imported into grafana as a new dashboard.

Note, When running the script you will be prompted for the namespace that the agent is running in.
