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

mpiJobName=$(<./etc/mpiJobName)
podStatus=""

while [[ podStatus != *"Running"* && podStatus != *"Completed"* ]]; do
  podStatus=$(kubectl get pods | grep "$mpiJobName-launcher")
  echo "$podStatus"
  sleep 1
done

echo kubectl get pods | grep "$mpiJobName-launcher"

echo "Before executing kubectl logs ..."

kubectl logs -f "$mpiJobName-launcher"

echo "After executing kubectl logs."

# Enable this after testing works out.
# kubectl delete "mpijobs/$mpiJobName"
