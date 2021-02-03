#!/bin/bash

values_yaml=$1

yq r -j  "${values_yaml}" | jq -r '[leaf_paths as $path | { "key": $path | join("."), "value": getpath($path) } | select( .key | endswith("requests.cpu")) | "\(.key),\(.value)" ]' > requests_cpu.csv

yq r -j  "${values_yaml}" | jq -r '[leaf_paths as $path | { "key": $path | join("."), "value": getpath($path) } | select( .key | endswith("limits.cpu")) | "\(.key),\(.value)" ]' > limits_cpu.csv

yq r -j  "${values_yaml}" | jq -r '[leaf_paths as $path | { "key": $path | join("."), "value": getpath($path) } | select( .key | endswith("requests.memory")) | "\(.key),\(.value)" ]' > requests_mem.csv

yq r -j  "${values_yaml}" | jq -r '[leaf_paths as $path | { "key": $path | join("."), "value": getpath($path) } | select( .key | endswith("limits.memory")) | "\(.key),\(.value)" ]' > limits_mem.csv
