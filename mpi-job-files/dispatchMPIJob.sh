#!/bin/bash

# Working directory is assumed to be: /opt/openmpp/<openmpp-root-dir>/

# Parse openm web service arguments and create manifest instance:
manifest=$(python3 ./bin/parseCommand.py "$@")

# Create a copy for trouble-shooting:
echo "$manifest" > ./etc/temp.yaml

# Send manifest to standard input of kubectl:
kubectl apply -f - <<< "$manifest"

# Now we'd like to run some kind of job monitoring command that returns success when MPIJob is done.
#kubectl wait --for=condition=complete job/MPIJob

