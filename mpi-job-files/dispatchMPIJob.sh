#!/bin/bash

# Working directory is assumed to be: /opt/openmpp/<openmpp-root-dir>/

# Parse openm web service arguments and create manifest instance:
manifest=$(python3 ./bin/parseCommand.py "$@")

echo "Input Arguments:"
cat ./etc/inputArguments

# Create a copy for trouble-shooting:
echo "$manifest" > ./etc/temp.yaml

# Send manifest to standard input of kubectl:
kubectl apply -f - <<< "$manifest"

# Now we'd like to run some kind of job monitoring command that returns success when MPIJob is done.
# Will try a polling loop with kubectl describe and grepping for relevant parts.
# The terminating condition will be status succeeded or failed.
mpiJobName=$(<./etc/mpiJobName)
cat "Pending" > ./etc/mpiJobStatusA
touch ./etc/mpiJobStatusB
mpiJobStatus="pending"

while [[ "$mpiJobStatus" != "done" ]]; do
  # Echo a simple status update. We can elaborate on this later.
  kubectl describe mpijobs/"$mpiJobName" | grep "Message" > ./etc/mpiJobStatusB 

  # Output only newest messages:
  comm -13 ./etc/mpiJobStatusA ./etc/mpiJobStatusB
 
  # Copy over newest changes to old file:
  cp -f ./etc/mpiJobStatusB ./etc/mpiJobStatusA

  if [[ $(cat ./etc/mpiJobStatusB | grep "JobSucceeded") != "" ]]
    mpiJobStatus="done"
  fi

  if [[ $(cat ./etc/mpiJobStatusB | grep "JobSucceeded") == "" ]]
    # Wait a short interval between updates to UI.
    echo ". . ."
    sleep 5
  fi
done
