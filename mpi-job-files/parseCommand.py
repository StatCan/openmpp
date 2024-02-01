import sys
import re
import os
from time import time
from pathlib import Path
import subprocess

# Working directory is assumed to be: 
# /opt/openmpp/<openmpp-root-dir>/

# Get directory where model executables are stored:
with open("./oms_model_dir") as mD:
  modelBinsDir = os.path.join(mD.read().strip("\n"), "bin".strip("\n"))

#KLW model isolation test
#print("##############################")
#print(modelBinsDir)

# Load manifest template contents:
with open("./MPIJobTemplate.yaml") as template:
  manifest = template.read()

# Save input arguments to file for debugging:
with open("./inputArguments", "w") as inputArgs:
  inputArgs.write(' '.join(sys.argv))

#no dot!
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
  #rint(sys.argv[i])
  if (i + 1 < len(sys.argv) and re.match("^-modelName$", sys.argv[i])):
    # Construct fully qualified model executable name:
    modelExecutable = '_'.join([os.path.join(modelBinsDir, sys.argv[i+1]), "mpi"])
    modelName = sys.argv[i+1]
    # KLW 16-01-2024 https://github.com/StatCan/openmpp/issues/51
    if not os.path.isfile(modelExecutable):
        exit()
      
    # Enter unique mpijob name:
    mpiJobName = f"{sys.argv[i+1]}-{str(time()).replace('.', '-')}".lower()
    manifest = manifest.replace("#<mpiJobName>", mpiJobName)

    # Pass name to file for monitoring mpijob later:
    with open("./mpiJobName", "w") as jN:
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
    openmOptions.append(sys.argv[i])
    openmOptions.append(sys.argv[i+1])

    if re.match("^-OpenM.RunName$", sys.argv[i]):
      #print(sys.argv[i+1])
      runName = sys.argv[i+1]

    i += 2
    
  # KLW 2024-01-22 #get dbPath
  elif (i + 1 < len(sys.argv) and re.match("^-dbPath$", sys.argv[i])): 
  #and re.match("[a-zA-Z0-9_/\.-]+", sys.argv[i+1])):
    dbPath = sys.argv[i+1]
    if not os.path.isfile(dbPath):
        exit()
    i += 2
      
  # Unrecognized command line options:
  else:
    unrecognized += f"{sys.argv[i]}\n"
    i += 1

# Write any unrecognized options to file for debugging:
with open("./unrecognizedCmdLineOptions", "w") as u:
  u.write(unrecognized)



#KLW 22-01-2024 Isolate sqlite dbfile by making a copy 
#https://github.com/StatCan/openmpp/issues/31

#copy database
dbDir  = os.path.dirname(dbPath)
dbName = os.path.basename(dbPath)
fext = os.path.splitext(dbPath)[1]

ndbName = runName+fext
ndirName = dbDir+"/"+runName
ndirName = os.path.normpath(ndirName)

#print(ndirName)

mdcmd = f'mkdir "{ndirName}"'
os.system(mdcmd)
#print(mdcmd)

ndbPath = os.path.join(dbDir, ndbName)
      
cfcmd = f'cp "{dbPath}" "{ndirName}"'
os.system(cfcmd)
#print(cfcmd)

cecmd = f'cp "{modelExecutable}" "{ndirName}"'
os.system(cecmd)
#print(cecmd)

# switch modelExecutable to point to one in new dir
modelExecutable = '_'.join([os.path.join(ndirName, modelName), "mpi"])

#print(modelExecutable)

#openmOptions.append(sys.argv[i+1]) # add OpenM.DbPath 
#openmOptions.append("-OpenM.DbPath")
#openmOptions.append(ndbPath)
#print(openmOptions)

# Compose bash arguments and replace placeholder in mpijob template:
bashArguments = f"ulimit -s 63356 && {modelExecutable} {' '.join(openmOptions)}"
manifest = manifest.replace("#<bashArguments>", f'- "{bashArguments}"')

# Print manifest to standard output:
#print("@@@@@@@@@@@@@@@@@@@@@@@@@@")
print(manifest)
