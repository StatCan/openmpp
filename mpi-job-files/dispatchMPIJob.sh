#!/bin/bash

# Working directory is assumed to be: /opt/openmpp/<openmpp-root-dir>/

# Send input arguments to console for debugging purposes:
echo "$@"

# Parse openm web service arguments and create manifest instance:
manifest=$(python3 ./bin/parseCommand.py "$@")

# Create a copy for trouble-shooting:
echo "$manifest" > ./etc/temp.yaml

# Send manifest to standard input of kubectl:
kubectl apply -f - <<< "$manifest"

# Now we'd like to run some kind of job monitoring command that returns success when MPIJob is done.
#kubectl wait --for=condition=complete job/MPIJob

# Will try a polling loop with kubectl describe and grepping for relevant parts.
# The terminating condition will be status succeeded or failed.
mpijobstatus=""
mpijobname=$(<./etc/mpiJobName)

while [ "$mpijobstatus" != *"Succeeded"* && "$mpijobstatus" != *"Failed"* ];
do
  # Echo a simple status update. We can elaborate on this later.
  mpijobstatus=$(kubectl describe mpijobs/"$mpijobname" | grep status) 
  echo $mpijobstatus
  echo ". . ."

  # Wait a short interval
  wait 5
done
