#!/bin/bash
# update_yaml_paths.sh
# Usage: ./update_yaml_paths.sh <logPath> <topoPath> <dataPath> <yaml_files...>
# Example: ./update_yaml_paths.sh ./log ./bgp_features/topo.yaml ./bgp_features/bgp.data example/*.yaml

set -e

if [ $# -lt 4 ]; then
    echo "Usage: $0 <logPath> <topoPath> <dataPath> <yaml_files...>"
    echo "Example: $0 ./log ./bgp_features/topo.yaml ./bgp_features/bgp.data example/*.yaml"
    exit 1
fi

LOG_PATH="$1"
TOPO_PATH="$2"
DATA_PATH="$3"
shift 3

for file in "$@"; do
    if [ -f "$file" ]; then
        echo "Updating: $file"
        sed -i "s|logPath:.*|logPath: \"${LOG_PATH}\"|g" "$file"
        sed -i "s|topo:.*|topo: \"${TOPO_PATH}\"|g" "$file"
        sed -i "s|data:.*|data: \"${DATA_PATH}\"|g" "$file"
    else
        echo "Warning: $file not found, skipping"
    fi
done

echo "Done."
