#!/usr/bin/env bash

set -euo pipefail

# Make a registry
docker run --name registry --rm -it -d -p 80:5000 registry:2

# Backup cluster-agent chart images
./backup.sh

# After some time has passed, we import them to our local registry, could be days later
./import.sh -b stackstate.tar.gz -d localhost -p

# verify import success
docker images | grep localhost

# Clean up
docker stop registry
docker rmi registry:2
rm stackstate.tar.gz

echo "Goodbye"
