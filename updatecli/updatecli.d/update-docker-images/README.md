# Docker Image Update Pipeline

Single pipeline with 28 targets so updatecli owns one checkout while updating all image tags. Each target uses its own `sourceid`; there is no artificial target-to-target dependency chain.
