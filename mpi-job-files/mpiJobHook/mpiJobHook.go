// A handler for mpi-based run requests to be called by the openm web service application. 
// Accesses the kubernetes api endpoint for mpijobs, requests new mpijob, relays log information from the
// launcher pod as the job is running, and responds to an abort request if it is relayed by the web service.

// Rename package name to mpijob after we test the handler functionality, and remove the test harness code.
package main //mpiJob

import (
    "os",
    "fmt",
    "path",
    "time",
    "context",
    "strings",
	"k8s.io/client-go/rest",
	"k8s.io/client-go/kubernetes",
    "k8s.io/apimachinery/pkg/api/resource",
    core "k8s.io/api/core/v1",
    meta "k8s.io/apimachinery/pkg/apis/meta/v1",
    kubeflow "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
)

// Some other imports from go-client examples. 
// We'll add to imports block if/when need anything from these.
// "k8s.io/apimachinery/pkg/api/errors"
// _ "k8s.io/client-go/plugin/pkg/client/auth"
// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"


// Arguments that are currently passed via the kubeflow template:
//  ModelName string            // model name
//  ExeStem   string            // base part of model exe name, usually modelName
//  Dir       string            // work directory to run the model
//  BinDir    string            // bin directory where model exe is located
//  DbPath    string            // absolute path to sqlite database file: models/bin/model.sqlite
//  MpiNp     int               // number of MPI processes
//  HostFile  string            // if not empty then path to hostfile
//  Args      []string          // model command line arguments
//  Env       map[string]string // environment variables to run the model  


// Will make a simple test harness for the handler and exploration of the rest apis.
// Check if there are any mandatory arguments for main functions.
func main () {
    // Create in-cluster configuration object:
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// Obtain the clientset from cluster:
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

    // First we just want to figure out how to submit a single mpi job.
    // So let's set some sample values here:
    modelName :=
    dir :=
    binDir :=
    dbPath :=
    hostFile :=
    mpiNp := 4
    args := string[]{ // Find some samples, I don't remember which options were key.
        "-OpenM. ": ,
        "-OpenM. ": ,
        "-OpenM. ":
    }

    // Invoke Handler in the same way that oms invokes the goroutine for the ExecCommand thing.
    //go Handler(/* Could use hard-coded arguments for now */)
    Handler(modelName, exeStem, dir, binDir, dbPath, hostFile, mpiNp, args, env)

    // Will need to formulate something analogous to this but for the mpijobs api endpoint: 
	// pods, err := clientset.CoreV1().Pods("").List(context.TODO(), meta.ListOptions{})

    // clientset.KubeflowV1().mpiJobs("").List(context.TODO(), meta.ListOptions{})

    // We could pass a unidirectional signal argument as an additional argument and use it to
    // signal an abort command if/when required.
    // Use standard input to send signal to terminate.

    // Use os.Exit to terminate the program and return a status code.
    // In cases of error we will be able to send an error status code that the web service should pick up.
    os.Exit(0)
}


func Handler (modelName, exeStem, dir, binDir, dbPath, hostFile string, mpiNp int, args []string, env map[string]string ) (/*no obvious return value*/) {

    // Start off by constructing constituent parts from the bottom up:
    timeStamp := time.Now().UnixNano()
    containerImage := "k8scc01covidacr.azurecr.io/ompp-run-ubuntu:0f80cf47bb5d6f6e025990e511b72be614bbcf7c"

    //Need to join modelName with binDir and append _mpi:      
    modelExecutable := strings.Join(string[]{path.Join(binDir, modelName), "mpi"}, "_")

    // Entrypoint command for containers:
    containerCommand := string[]{"mpirun", "/bin/bash"}

    // Append the ulimit setting, fully qualified model exec filename, and all OpenM options:
    bashArguments := append(string[]{"ulimit -s 63356 &&", modelExecutable}, args...)

    // I think any occurrences of environment variables (entries prefixed by $) in the container
    // arguments field are automatically resolved to the values these environment variables hold
    // in the given execution context. That's why there are separate fields for command and 
    // arguments in the container type.
    containerArguments := append(string[]{"-c"}, strings.Join(bashArguments, " "))

    // CPU resource limits and requests:
    cpuResourceLimit := resource.NewMilliQuantity(2000, resource.BinarySI)
    cpuLauncherRequest := resource.NewMilliQuantity(250, resource.BinarySI)
    cpuWorkerRequest := resource.NewMilliQuantity(2000, resource.BinarySI)

    // Memory resource limits and requests:
    memoryLimit := resource.NewQuantity(2 * 1024 * 1024 * 1024, resource.BinarySI)
    memoryLauncherRequest := resource.NewQuantity(250 * 1024 * 1024, resource.BinarySI)
    memoryWorkerRequest := resource.NewQuantity(1024 * 1024 * 1024, resource.BinarySI)

    // Node resource limits:
    resourceLimits := core.ResourceList {
        ResourceCPU: cpuResourceLimit,
        ResourceMemory: memoryLimit
    }

    // Launcher resource requests:
    launcherResourceRequests := core.ResourceList {
        ResourceCPU: cpuLauncherRequest,
        ResourceMemory: memoryLauncherRequest
    }

    // Worker resource requests:
    workerResourceRequests := core.ResourceList {
        ResourceCPU: cpuWorkerReqest,
        ResourceMemory: memoryWorkerRequest
    }

    // Launcher resource requirements:
    launcherResoureRequirements := core.ResourceRequirements {
        Limits: resourceLimits,
        Requests: launcherResourceRequests
    }

    // Worker resource requirements:
    workerResourceRequirements := core.ResourceRequirements {
        Limits: resourceLimits,
        Requests: workerResourceRequests
    }

    // Launcher container spec:
    mainContainerName := strings.Join(string[]{modelName, timeStamp, "launcher"}, "-")
    launcherContainer := core.Container {
        Name: mainContainerName,
        Image: containerImage,
        Command: containerCommand,
        Resources: launcherResourceRequirements
    }

    // Worker container spec:
    workerContainer := core.Container {
        Name: strings.Join(string[]{modelName, timeStamp, "worker"}, "-"),
        Image: containerImage,
        Resources: workerResourceRequirements
    }

    // Next is pod specs:
    launcherPodSpec := core.PodSpec {
        Containers: core.Container[]{launcherContainer}
    }

    workerPodSpec := core.PodSpec {
        Containers: core.Container[]{workerContainer}
    }

    // Labels for worker and launcher pods:
    labels := map[string][string]{
        "data.statcan.gc.ca/inject-blob-volumes": "true",
        "sidecar.istio.io/inject": "false"
    }

    podObjectMetadata := meta.ObjectMeta {
        Labels: labels
    }

    podTypeMetadata := meta.TypeMeta {
        Kind: "PodTemplate",
        Version: "v1"
    }

    // Next is pod template specs:
    laucherPodTemplate := core.PodTemplateSpec {
        TypeMeta: podTypeMetadata,
        ObjectMeta: podObjectMetadata,
        Spec: launcherPodSpec
    }

    workerPodTemplate := core.PodTemplateSpec {
        TypeMeta: podTypeMetadata,
        ObjectMeta: podObjectMetadata,
        Spec: workerPodSpec
    }

    one := 1
    // Then replica specs:
    launcheReplicaSpec := kubeflow.ReplicaSpec {
        Replicas: &one
        Template: launcherPodTemplate
    }

    workerReplicaSpec := kubeflow.ReplicaSpec {
        Replicas: &mpiNp
        Template: workerPodTemplate
    }

    two := 2
    // Finally the mpijob spec:
    mpiJobSpec := kubeflow.MPIJobSpec {
        SlotsPerWorker: &two // probably 2 given Pat's comments about typical nodes used in the cluster.

        // This one is a map from the set of replica types {Launcher, Worker, ... } to ReplicaSpecs.
        // We don't need to provide ReplicaSpecs for any replica types other than Launcher and Worker.
        // The reason they have these other ones is for specifying different types of distributed workloads.
        MPIReplicaSpecs: map[kubeflow.ReplicaType]kubeflow.ReplicaSpec {
            "Launcher": launcherReplicaSpec,
            "Worker": workerReplicaSpec
        },

        MainContainer: mainContainerName,
        CleanPodPolicy: &kubeflow.CleanPodPolicyRunning
    }
}


// Leaving this here as reference for how the clientset object is interacted with.
func notMain () {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), meta.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "example-xxxxx", meta.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found example-xxxxx pod in default namespace\n")
		}

		time.Sleep(10 * time.Second)
	}
}
