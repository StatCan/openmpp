#!/bin/bash

# Working directory is assumed to be: /opt/openmpp/<openmpp-root-dir>/

# Parse openm web service arguments and create manifest instance:
manifest=$(python3 ./bin/parseCommand.py "$@")

cat ./etc/inputArguments

# Create a copy for trouble-shooting:
echo "$manifest" > ./etc/temp.yaml

# Send manifest to standard input of kubectl:
kubectl apply -f - <<< "$manifest"

# Now we'd like to run some kind of job monitoring command that returns success when MPIJob is done.
# Will try a polling loop with kubectl describe and grepping for relevant parts.
# The terminating condition will be status succeeded or failed.
mpiJobStatus=""
mpiJobName=$(<./etc/mpiJobName)

while [[ "$mpiJobStatus" != *"Succeeded"* && "$mpiJobStatus" != *"Failed"* ]]; do
  # Echo a simple status update. We can elaborate on this later.
  mpiJobStatus=$(kubectl describe mpijobs/"$mpiJobName" | grep status) 
  echo $mpiJobStatus
  echo ". . ."

  # Wait a short interval between updates to UI.
  sleep 2
done
