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

# The remaining manifest values come from command line options passed by oms:
i = 1
while i < len(sys.argv):
  if (i + 1 < len(sys.argv) and re.match("^-modelName$", sys.argv[i])):
    # Enter model executable name to mpirun command:
    modelExecutable = '_'.join([os.path.join(modelBinsDir, sys.argv[i+1]), "mpi"])
    if os.path.isfile(modelExecutable):
      manifest = manifest.replace("#<modelExecutable>", f"- {modelExecutable}")

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

  # OpenM options and arguments:
  elif (i + 1 < len(sys.argv) and re.match("^-OpenM\.", sys.argv[i]) \
  and re.match("[a-zA-Z0-9_/\.-]+", sys.argv[i+1])):
    manifest = manifest.replace("#<OpenMOption>", \
      f"- {sys.argv[i]}\n{12*' '}- '{sys.argv[i+1]}'\n{12*' '}#<OpenMOption>") 
    i += 2

  # Unrecognized command line options:
  else:
    unrecognized += f"{sys.argv[i]}\n"
    i += 1

# Write any unrecognized options to file for debugging:
with open("./etc/unrecognized", "w") as u:
  u.write(unrecognized)

# Print manifest to standard output:
print(manifest)
