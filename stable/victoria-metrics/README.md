# Info

We use the single node deployment of Victoria Metrics - each node has own copy of all data.

## Backups

### Motivation

Victoria Metrics provides a tool named `vmbackup` to backup all data stored in the database. The tool requires to have
access to the same file system used by Victoria Metrics. So we have to options:

- use Volumes with mode `RWX`
- use `vmbackup` on the same pod as `Victoria Metrics`

We can't use RWX Volumes (at least not now), so we have to deploy everything to the same pod as Victoria Metrics.

Victoria Metrics has provided a helm chart to run the service it also allows to run a side car container with `vmbackup`
but the configuration isn't flexible so we have to start "coding" inside a `values.yaml`, also we have to repeat the
same configuration twice (we have to Pods with Victoria Metrics) and also we had to copy the save value into multiple
places (child helm chart can't use values from the parent chart). Because of these limitations and problems I have
decided to "fork" original helm chart and add some customization.
