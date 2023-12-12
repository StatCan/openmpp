#!/bin/bash

# Working directory is assumed to be: /opt/openmpp/<openmpp-root-dir>/

# Parse openm web service arguments and create manifest instance:
manifest=$(python3 ./bin/parseCommand.py "$@")

# echo "Input Arguments:"
# cat ./etc/inputArguments

# Create a copy for trouble-shooting:
echo "$manifest" > ./etc/temp.yaml

# Send manifest to standard input of kubectl:
kubectl apply -f - <<< "$manifest"

# Get generated name of mpijob:
mpiJobName=$(<./etc/mpiJobName)

podStatus="Pending"
while [[ podStatus != *"Running"* ]]; do
  podStatus=$(kubectl get pods | grep "$mpiJobName-launcher")
  sleep 2
done

kubectl logs -f "$mpiJobName-launcher"

# Enable this after testing works out.
# kubectl delete "mpijobs/$mpiJobName"
