# sandbox-operator

![Version: 0.1.6](https://img.shields.io/badge/Version-0.1.6-informational?style=flat-square) ![AppVersion: 0.5.0](https://img.shields.io/badge/AppVersion-0.5.0-informational?style=flat-square)
The StackState Sandboxer
**Homepage:** <https://github.com/StackVista/sandbox-operator>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| reaper.defaultTTL | string | `"168h"` | Default TTL for a Sandbox (default: 168 hours = 1 week) |
| reaper.firstExpirationWarning | string | `"72h"` | How long in advance to warn the user that his sandbox will expire (default: 72 hours = 3 days) |
| reaper.image | string | `"ghcr.io/stackvista/sandboxer:0.5.0"` | Image for the reaper job |
| reaper.messages.expirationOverdue | string | `"<@{{ .Sandbox.Spec.SlackID }}>: Your sandbox `{{ .Sandbox.Name }}` cannot be automatically expired and has been running a long time, consider cleaning it up."` |  |
| reaper.messages.expirationWarning | string | `"<@{{ .Sandbox.Spec.SlackID }}>: Your sandbox `{{ .Sandbox.Name }}` will expire on {{ .ExpirationDate }}"` |  |
| reaper.messages.reapMessage | string | `"<@{{ .Sandbox.Spec.SlackID }}>: Your sandbox `\\{\\{ .Sandbox.Name }}` has been reaped..."` |  |
| reaper.resources.limits.cpu | string | `"50m"` |  |
| reaper.resources.limits.memory | string | `"64Mi"` |  |
| reaper.resources.requests.cpu | string | `"25m"` |  |
| reaper.resources.requests.memory | string | `"32Mi"` |  |
| reaper.schedule | string | `"0 0 * * *"` | Cron schedule for the reaper, once per day at midnight |
| reaper.slack.apiKey | string | `""` | Slack API token |
| reaper.slack.channelId | string | `""` | Slack Channel ID to post in (can be an ID or the channel name prefixed with a '#') |
| reaper.warningInterval | string | `"24h"` | Interval between 2 warnings that the sandbox will expire (default: 24 hours = 1 day) |
| sandboxer.image | string | `"ghcr.io/stackvista/sandboxer:0.5.0"` | Image for the sandbox operator |
| sandboxer.resources.limits.cpu | string | `"50m"` |  |
| sandboxer.resources.limits.memory | string | `"64Mi"` |  |
| sandboxer.resources.requests.cpu | string | `"25m"` |  |
| sandboxer.resources.requests.memory | string | `"32Mi"` |  |
