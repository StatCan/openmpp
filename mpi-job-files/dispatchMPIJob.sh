#!/bin/bash

# Working directory is assumed to be: /opt/openmpp/<openmpp-root-dir>/

# Parse openm web service arguments and create manifest instance:
manifest=$(python3 ./bin/parseCommand.py "$@")

# KLW 16-01-2024 https://github.com/StatCan/openmpp/issues/51
if [[ -z $manifest ]]; then
  echo "ERROR: path to MPI model executable does not exist"
  echo "Did you compile it?"
  exit 1
fi

# Create a copy of manifest for trouble-shooting:
echo "$manifest" > "./etc/MPIJob-$ManifestInstance.yaml"

# Send manifest to standard input of kubectl:
kubectl apply -f - <<< "$manifest"

# Set up variables to check mpijob launcher pod status:
mpiJobName=$(<./etc/mpiJobName)
podStatus=""

# Poll for status of mpijob launcher pod until it is running or stopped:
while [[ $podStatus != *"Running"* && $podStatus != *"Completed"* && $podStatus != *"Error"* ]]; do
  podStatus=$(kubectl get pods | grep "$mpiJobName-launcher")
  echo "$podStatus"
  sleep 1
done

# Show final pod status before going into logs:
kubectl get pods | grep "$mpiJobName-launcher"

# Forward logs command output from launcher pod:
kubectl logs -f "$mpiJobName-launcher"

# Enable this after testing works out.
# kubectl delete "mpijobs/$mpiJobName"
