# Docker Image Update Pipeline

Single pipeline so updatecli owns one checkout while updating all image tags. Each target uses its own `sourceid`; there is no artificial target-to-target dependency chain.
