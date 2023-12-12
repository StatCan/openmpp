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

# Now we'd like to run some kind of job monitoring command that returns success when MPIJob is done.
# Will try a polling loop with kubectl describe and grepping for relevant parts.
# The terminating condition will be status succeeded or failed.
mpiJobName=$(<./etc/mpiJobName)

touch ./etc/mpiJobStatusA
touch ./etc/mpiJobStatusB

cat "Pending" > ./etc/mpiJobStatusA
mpiJobStatus="pending"

while [[ "$mpiJobStatus" != "done" ]]; do
  # Grep for status messages regarding current mpijob displayed by kubectl: 
  kubectl describe mpijobs/"$mpiJobName" | grep "Message:" > ./etc/mpiJobStatusB 

  # Output only newest messages:
  comm -13 ./etc/mpiJobStatusA ./etc/mpiJobStatusB
 
  # Copy over newest changes to old file:
  cp -f ./etc/mpiJobStatusB ./etc/mpiJobStatusA

  if [[ $(cat ./etc/mpiJobStatusB | grep "JobSucceeded") != "" || $(cat ./etc/mpiJobStatusB | grep "JobFailed" != "" ]]; then
    mpiJobStatus="done"
  else
    # Wait a short interval between updates to UI:
    echo ". . ."
    sleep 5
  fi
done
