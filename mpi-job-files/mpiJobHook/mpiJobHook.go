// A handler for mpi-based run requests to be called by the openm web service application. 
// Accesses the kubernetes api endpoint for mpijobs, requests new mpijob, relays log information from the
// launcher pod as the job is running, and responds to an abort request if it is relayed by the web service.

// Rename package name to mpijob after we test the handler functionality, and remove the test harness code.
package main //mpiJob

import (
    "fmt",
    "context",
	"k8s.io/client-go/rest",
	"k8s.io/client-go/kubernetes",
    core "k8s.io/api/core/v1",
    meta "k8s.io/apimachinery/pkg/apis/meta/v1",
    kubeflow "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
)

// Some other imports from go-client examples. 
// We'll add to imports block if/when need anything from these.
// "time"
// "k8s.io/apimachinery/pkg/api/errors"
// _ "k8s.io/client-go/plugin/pkg/client/auth"
// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

// Will make a simple test harness for the handler and exploration of the rest apis.
// Check if there are any mandatory arguments for main functions.
func main () {
    // Invoke Handler in the same way that oms invokes the goroutine for the ExecCommand thing.
    go Handler(/* Could use hard-coded arguments for now */)

    // We could pass a unidirectional signal argument as an additional argument and use it to
    // signal an abort command if/when required.
    // Use standard input to send signal to terminate.
}

// Arguments currently passed via the kubeflow template:
//  ModelName string            // model name
//  ExeStem   string            // base part of model exe name, usually modelName
//  Dir       string            // work directory to run the model
//  BinDir    string            // bin directory where model exe is located
//  DbPath    string            // absolute path to sqlite database file: models/bin/model.sqlite
//  MpiNp     int               // number of MPI processes
//  HostFile  string            // if not empty then path to hostfile
//  Args      []string          // model command line arguments
//  Env       map[string]string // environment variables to run the model  

func Handler (modelName, exeStem, dir, binDir, dbPath, hostFile string, mpiNp int, args []string, env map[string]string) (/*no obvious return value*/) {

    // Let's walk down the structure hierarchy for an mpijob specification: 
    var mpiJobSpec kubeflow.MPIJobSpec

    // probably 2 given Pat's comments about the typical nodes used in the cluster:
    var slotsPerWorker *int32

    var cleanPodPolicy *CleanPodPolicy

    // This one is a map from the set of replica types {Launcher, Worker, ... } to ReplicaSpecs.
    // We don't need to provide ReplicaSpecs for any replica types other than Launcher and Worker.
    // The reason they have these other ones is for specifying different types of distributed workloads.
    var MPIReplicaSpecs map[kubeflow.ReplicaType]kubeflow.ReplicaSpec

    var mainContainer string
    var runPolicy kubeflow.RunPolicy

    var launcherReplicaSpec, workerReplicaSpec kubeflow.ReplicaSpec

    var launcherReplicasNumber, workerReplicasNumber *int32
    var launcherPodTemplateSpec, workerPodTemplateSpec core.PodTemplateSpec

    var launcherMetadata, workerMetadata core.
    var launcherPodSpec, workerPodSpec core.PodSpec

    var mpiJobContainer core.Container
    var containerName string
    var containerImage string
    var containerCommand []string
    var containerArguments []string
    var workingDirectory

    var resourceRequirements core.ResourceRequirements
    var limits, requests core.ResourceList

    var RestartPolicy

    //Maybe these?
    var HostName
    var ServiceAccountName
    var NodeSelector


    // So start off by constructing the container parts.
    timeStamp := time.Now().UnixNano()
    containerImage := "k8scc01covidacr.azurecr.io/ompp-run-ubuntu:0f80cf47bb5d6f6e025990e511b72be614bbcf7c"

    containerCommand := string[]{

    mpiJobLauncherContainer = {
        Name: modelName + "-" + timeStamp + "launcher"
        Image: containerImage

    }

    mpiJobWorkerContainer = {
        Name: modelName + "-" timeStamp + "-" + "worker"
        Image: containerImage
    }

    
}

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
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "example-xxxxx", metav1.GetOptions{})
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
