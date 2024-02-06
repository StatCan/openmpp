import sys
import re
import os
from time import time
from pathlib import Path

# Working directory is assumed to be: 
# /opt/openmpp/<openmpp-root-dir>/

# Get directory where model executables are stored:
with open("./etc/oms_model_dir") as mD:
  modelBinsDir = os.path.join(mD.read().strip("\n"), "bin".strip("\n"))

# Load manifest template contents:
with open("./etc/MPIJobTemplate.yaml") as template:
  manifest = template.read()

# Save input arguments to file for debugging:
with open("./etc/inputArguments", "w") as inputArgs:
  inputArgs.write(' '.join(sys.argv))

with open("/etc/hostname") as nN:
  notebookName = nN.read()

# Save unrecognized command line options to file for debugging:
unrecognized = ""

# Some manifest values are based on system configuration:
# Set notebook name label so we can filter mpijobs by label for deletion:
manifest = manifest.replace("#<notebookName>", notebookName)

# Set working directory for mpirun to the ...models/bin directory:
manifest = manifest.replace("#<mpirunOption>", \
  f"- -wdir\n{12*' '}- {modelBinsDir}\n{12*' '}#<mpirunOption>")

#These variables need to be scoped outside of the while loop:
openmOptions = []
modelExecutable = ""

# The remaining manifest values come from command line options passed by oms:
i = 1
while i < len(sys.argv):
  if (i + 1 < len(sys.argv) and re.match("^-modelName$", sys.argv[i])):
    # Construct fully qualified model executable name:
    modelExecutable = '_'.join([os.path.join(modelBinsDir, sys.argv[i+1]), "mpi"])

    # KLW 16-01-2024 https://github.com/StatCan/openmpp/issues/51
    if not os.path.isfile(modelExecutable):
        exit()
      
    # Enter unique mpijob name:
    mpiJobName = f"{sys.argv[i+1]}-{str(time()).replace('.', '-')}".lower()
    manifest = manifest.replace("#<mpiJobName>", mpiJobName)

    # Pass name to file for monitoring mpijob later:
    with open("./etc/mpiJobName", "w") as jN:
      jN.write(mpiJobName)
    i += 2

  # Number of replicas to create:
  elif (i + 1 < len(sys.argv) and re.match("^-mpiNp$", sys.argv[i]) and re.match("^[0-9]+$", sys.argv[i+1])):
    manifest = manifest.replace("#<numberOfReplicas>", f"{sys.argv[i+1]}")
    manifest = manifest.replace("#<mpirunOption>", \
      f"- -n\n{12*' '}- '{sys.argv[i+1]}'\n{12*' '}#<mpirunOption>")
    i += 2

  # mpirun bind-to option:
  elif (i + 1 < len(sys.argv) and re.match("^--bind-to$", sys.argv[i]) \
  and re.match("^(core|socket|none)$", sys.argv[i+1])):
    manifest = manifest.replace("#<mpirunOption>", \
      f"- --bind-to\n{12*' '}- {sys.argv[i+1]}\n{12*' '}#<mpirunOption>")
    i += 2

  # mpirun environment variable options:
  elif (i + 1 < len(sys.argv) and re.match("^-x$", sys.argv[i]) \
  and re.match("[a-zA-Z_][a-zA-Z0-9_]*=[^=]+$", sys.argv[i+1])):
    manifest = manifest.replace("#<mpirunOption>", \
      f"- -x\n{12*' '}- {sys.argv[i+1]}\n{12*' '}#<mpirunOption>")
    i += 2

  # OpenM options and arguments exception:
  # KLW 2024-02-06  https://github.com/StatCan/openmpp/issues/60
  # There is a bug in OpenM in regard to the -OpenM.NotOnRoot argument.  
  # The Value is suppose to be true or false, but is missing when false
  elif (i + 1 < len(sys.argv) and re.match("^-OpenM.NotOnRoot$", sys.argv[i])):
    if (re.match("^true$", sys.argv[i+1]) or re.match("^false$", sys.argv[i+1]) ):
      openmOptions.append(sys.argv[i])
      openmOptions.append(sys.argv[i+1])
      i += 2
    else:
      openmOptions.append(sys.argv[i])
      openmOptions.append("true") #KLW 2024-02-06 OpenM docs say this value should be 'true'
      i += 1

  # ini options and arguments:
  # KLW 2024-02-06 Explicit handling of the -ini argument
  elif (i + 1 < len(sys.argv) and re.match("^-ini$", sys.argv[i]) \
  and re.match("[a-zA-Z0-9_/\.-]+", sys.argv[i+1])):
    openmOptions.append(sys.argv[i])
    openmOptions.append(sys.argv[i+1])
    i += 2
    
  
  # OpenM options and arguments:
  elif (i + 1 < len(sys.argv) and re.match("^-OpenM\.", sys.argv[i]) \
  and re.match("[a-zA-Z0-9_/\.-]+", sys.argv[i+1])):
    openmOptions.append(sys.argv[i])
    openmOptions.append(sys.argv[i+1])
    i += 2

  # Unrecognized command line options:
  else:
    unrecognized += f"{sys.argv[i]}\n"
    i += 1

# Write any unrecognized options to file for debugging:
with open("./etc/unrecognizedCmdLineOptions", "w") as u:
  u.write(unrecognized)

# Compose bash arguments and replace placeholder in mpijob template:
bashArguments = f"ulimit -s 63356 && {modelExecutable} {' '.join(openmOptions)}"
manifest = manifest.replace("#<bashArguments>", f'- "{bashArguments}"')

# Print manifest to standard output:
print(manifest)
