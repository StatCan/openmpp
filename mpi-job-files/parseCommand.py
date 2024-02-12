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

# note, absolute path, no dot
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

############################
# KLW 2024-02-07 replace command line argument parsing
#input string (command line arguments) handling
# merge argv array into a single string
inputArgsString = ' '.join(sys.argv)
# split on ' -' instead of just space
theArgs = inputArgsString.split(" -")

argDict = {}

#arg[0] is the command itself, so ignore
for i in range(1, len(theArgs)):
  argItem = theArgs[i].split() # defaults to spliting on space
  argKey = "-" + argItem[0] # replace the - at the beginning
  argVal = ""
  for j in range(1,len(argItem)):
    argVal = argVal + argItem[j] # combine the rest of the argItem, removing spaces
    
  #finally, add to the Dict
  argDict[argKey] = argVal.strip()
  #print(argKey, " ", argVal)

# add the command for completeness
argDict["command"] = theArgs[0]

### Modify the manifest template
mv = argDict["-modelName"]
modelExecutable = '_'.join([os.path.join(modelBinsDir, argDict["-modelName"]), "mpi"])

# KLW 16-01-2024 https://github.com/StatCan/openmpp/issues/51
if not os.path.isfile(modelExecutable):
  print("ERROR. Model executable not found: ", modelExecutable)
  exit(1)

mpiJobName = f"{mv}-{str(time()).replace('.', '-')}".lower()

manifest = manifest.replace("#<mpiJobName>", mpiJobName)

mv = argDict["-mpiNp"]
manifest = manifest.replace("#<numberOfReplicas>", f"{mv}") #argDict["-mpiNp"]) 

manifest = manifest.replace("#<mpirunOption>", f"- -n\n{12*' '}- '{mv}'\n{12*' '}#<mpirunOption>")

if "--bind-to" in argDict:
  mv = argDict["--bind-to"]
  manifest = manifest.replace("#<mpirunOption>", f"- --bind-to\n{12*' '}- {mv}\n{12*' '}#<mpirunOption>")

if "-x" in argDict:
  mv = argDict["-x"]
  manifest = manifest.replace("#<mpirunOption>", f"- -x\n{12*' '}- {mv}\n{12*' '}#<mpirunOption>")

# loop through the -OpenM. arguments and add them to the openmOptions string
for key in argDict.keys():
  if re.match("^-OpenM\.", key):
    openmOptions.append(key)
    if len(argDict[key]) > 0:
      openmOptions.append(argDict[key])
  if re.match("^-ini$", key):
    openmOptions.append(key)
    openmOptions.append(argDict[key])


# Compose bash arguments and replace placeholder in mpijob template:
bashArguments = f"ulimit -s 63356 && {modelExecutable} {' '.join(openmOptions)}"
manifest = manifest.replace("#<bashArguments>", f'- "{bashArguments}"')

# Print manifest to standard output:
print(manifest)
