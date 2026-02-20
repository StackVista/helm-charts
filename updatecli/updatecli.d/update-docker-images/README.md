# Docker Image Update Pipeline

Single pipeline with 21 targets chained via `dependson` so they run sequentially and avoid overwriting each other (split pipelines ran in parallel and only the last write survived).
