// A handler for mpi-based run requests to be called by the openm web service application.
// Accesses the kubernetes api endpoint for mpijobs, requests new mpijob, relays log information from the
// launcher pod as the job is running, and responds to an abort request if it is relayed by the web service.

package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	kubeAPI "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
    kubeClient "github.com/kubeflow/training-operator/pkg/client/clientset/versioned/typed/kubeflow.org/v1"
)

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

func main() {
	// Set some sample argument values here.
	modelName := "RiskPaths"
	exeStem := "RiskPaths"
	dir := "/home/jovyan/buckets/aaw-unclassified/microsim/models/bin"
	binDir := "."
	dbPath := "/home/jovyan/buckets/aaw-unclassified/microsim/models/bin/RiskPaths.sqlite"
	var mpiNp int32 = 4
	args := []string{
        "-OpenM.RunStamp",
        "2024_02_09_04_08_45_834",
        "-OpenM.LogToConsole",
        "true",
        "-OpenM.LogRank",
        "true",
        "-OpenM.MessageLanguage",
        "en-US",
        "-OpenM.RunName",
        "RiskPaths_Default_2024_02_08_23_08_10_047",
        "-OpenM.SetName"
        "Default"
		"-OpenM.SubValues",
		"32",
		"-OpenM.Threads",
		"16",
	}
	env := map[string]string{
		"SAMPLE_ENV": "VALUE",
	}

	// Create an MPIJobSpec using the sample arguments defined above.
	jobSpec := mpiJobSpec(modelName, exeStem, dir, binDir, dbPath, mpiNp, args, env)

    // Validate spec using validation function.
    err := kubeAPI.ValidateV1MpiJobSpec(jobSpec)
    if err != nil {
        panic(err.Error())
    }

    // Create in-cluster configuration object.
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

    // Obtain clientset from cluster.
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

    // Obtain client subset containing just the kubeflow based resources.
    kubeClientSubset, err := kubeClient.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    // Obtain an interface to the MPIJobs collection for given namespace.
    mpiJobs := kubeClientSubset.MPIJobs("jacek-dev")

    // Now attempt to submit sample MPIJob:
    func (c *mPIJobs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.MPIJob, err error) 

    func (c *mPIJobs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.MPIJobList, err error) {

	// Use os.Exit to terminate the program and return a status code.
	// In cases of error we will be able to send an error status code that the web service should pick up.
	os.Exit(0)
}


// Might want to define a structure for the arguments coming from openm web service as that will
// let us provide default values for some fields, and refactor mpiJobSpec below to accept the
// structure instead.
type mpiJobArgs struct {
	modelName string
	exeStem   string
	dir       string
	binDir    string
	dbPath    string
	mpiNp     int32
	args      []string
	env       map[string]string
}

// Generate mpijob spec based on arguments coming from openm web service and cluster configuration:
func mpiJobSpec(modelName, exeStem, dir, binDir, dbPath string, mpiNp int32, args []string, env map[string]string) kubeAPI.MPIJobSpec {

	// Start off by constructing constituent parts from the bottom up:
	timeStamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	containerImage := "k8scc01covidacr.azurecr.io/ompp-run-ubuntu:0f80cf47bb5d6f6e025990e511b72be614bbcf7c"

	//Need to join modelName with binDir and append _mpi:
	modelExecutable := strings.Join([]string{path.Join(binDir, modelName), "mpi"}, "_")

	// Entrypoint command for containers:
	containerCommand := []string{"mpirun", "/bin/bash"}

	// Append the ulimit setting, fully qualified model exec filename, and all OpenM options:
	bashArguments := append([]string{"ulimit -s 63356 &&", modelExecutable}, args...)

	// I think any occurrences of environment variables (entries prefixed by $) in the container
	// arguments field are automatically resolved to the values these environment variables hold
	// in the given execution context. That's why there are separate fields for command and
	// arguments in the container type.
	containerArguments := append([]string{"-c"}, strings.Join(bashArguments, " "))

	// CPU resource limits and requests:
	cpuResourceLimit := resource.NewMilliQuantity(2000, resource.BinarySI)
	cpuLauncherRequest := resource.NewMilliQuantity(250, resource.BinarySI)
	cpuWorkerRequest := resource.NewMilliQuantity(2000, resource.BinarySI)

	// Memory resource limits and requests:
	memoryLimit := resource.NewQuantity(2*1024*1024*1024, resource.BinarySI)
	memoryLauncherRequest := resource.NewQuantity(250*1024*1024, resource.BinarySI)
	memoryWorkerRequest := resource.NewQuantity(1024*1024*1024, resource.BinarySI)

	// Node resource limits:
	resourceLimits := core.ResourceList{
		core.ResourceCPU:    *cpuResourceLimit,
		core.ResourceMemory: *memoryLimit,
	}

	// Launcher resource requests:
	launcherResourceRequests := core.ResourceList{
		core.ResourceCPU:    *cpuLauncherRequest,
		core.ResourceMemory: *memoryLauncherRequest,
	}

	// Worker core requests:
	workerResourceRequests := core.ResourceList{
		core.ResourceCPU:    *cpuWorkerRequest,
		core.ResourceMemory: *memoryWorkerRequest,
	}

	// Launcher resource requirements:
	launcherResourceRequirements := core.ResourceRequirements{
		Limits:   resourceLimits,
		Requests: launcherResourceRequests,
	}

	// Worker resource requirements:
	workerResourceRequirements := core.ResourceRequirements{
		Limits:   resourceLimits,
		Requests: workerResourceRequests,
	}

	// Launcher container spec:
	mainContainerName := strings.Join([]string{modelName, timeStamp, "launcher"}, "-")
	launcherContainer := core.Container{
		Name:      mainContainerName,
		Image:     containerImage,
		Command:   containerCommand,
		Args:      containerArguments,
		Resources: launcherResourceRequirements,
	}

	// Worker container spec:
	workerContainer := core.Container{
		Name:      strings.Join([]string{modelName, timeStamp, "worker"}, "-"),
		Image:     containerImage,
		Resources: workerResourceRequirements,
	}

	// Next is pod specs:
	launcherPodSpec := core.PodSpec{
		Containers: []core.Container{launcherContainer},
	}

	workerPodSpec := core.PodSpec{
		Containers: []core.Container{workerContainer},
	}

	// Labels for worker and launcher pods:
	labels := map[string]string{
		"data.statcan.gc.ca/inject-blob-volumes": "true",
		"sidecar.istio.io/inject":                "false",
	}

	podObjectMetadata := meta.ObjectMeta{
		Labels: labels,
	}

	//podTypeMetadata := meta.TypeMeta{
	//	Kind:       "PodTemplate",
	//	APIVersion: "v1",
	//}

	// Pod template specifications:
	launcherPodTemplateSpec := core.PodTemplateSpec{
		ObjectMeta: podObjectMetadata,
		Spec:       launcherPodSpec,
	}

	workerPodTemplateSpec := core.PodTemplateSpec{
		ObjectMeta: podObjectMetadata,
		Spec:       workerPodSpec,
	}

	var one int32 = 1
	// Then replica specs:
	launcherReplicaSpec := kubeAPI.ReplicaSpec{
		Replicas: &one,
		Template: launcherPodTemplateSpec,
	}

	workerReplicaSpec := kubeAPI.ReplicaSpec{
		Replicas: &mpiNp,
		Template: workerPodTemplateSpec,
	}

	var two int32 = 2
	var cleanPodPolicy kubeAPI.CleanPodPolicy = kubeAPI.CleanPodPolicyRunning
	// Finally the mpijob spec:
	mpiJobSpec := kubeAPI.MPIJobSpec{
		SlotsPerWorker: &two, // probably 2 given Pat's comments about typical nodes used in the cluster.

		// This one is a map from the set of replica types {Launcher, Worker, ... } to *ReplicaSpecs.
		// We don't need to provide ReplicaSpecs for any replica types other than Launcher and Worker.
		// The reason they have these other ones is for specifying different types of distributed workloads.
		MPIReplicaSpecs: map[kubeAPI.ReplicaType]*kubeAPI.ReplicaSpec{
			"Launcher": &launcherReplicaSpec,
			"Worker":   &workerReplicaSpec,
		},

		MainContainer:  mainContainerName,
		CleanPodPolicy: &cleanPodPolicy,
	}
	return mpiJobSpec
}

// Function based on go-client example code that queries the kubernets api for pod info:
func showPods(cs *kubernetes.Clientset, namespace string) {
	// Specify namespace to get pods in particular namespace
	// or omit parameter to search all namespaces.
	pods, err := cs.CoreV1().Pods(namespace).List(context.TODO(), meta.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Use some reflection functions to explore the pods variable:

	// Examples for error handling:
	// - Use helper functions e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	_, err = cs.CoreV1().Pods("default").Get(context.TODO(), "example-xxxxx", meta.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod example-xxxxx not found in default namespace\n")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found example-xxxxx pod in default namespace\n")
	}
}

// Modify this to conform to our use case. We'll have already created a clientset.
// And we will want to use a label selector or name to locate the launcher pod for a given MPIjob.
func getPodLogs(pod core.Pod) string {
    podLogOpts := core.PodLogOptions{}
    config, err := rest.InClusterConfig()
    if err != nil {
        return "error in getting config"
    }
    // creates the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return "error in getting access to K8S"
    }
    req := clientset.Core().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
    podLogs, err := req.Stream()
    if err != nil {
        return "error in opening stream"
    }
    defer podLogs.Close()

    buf := new(bytes.Buffer)
    _, err = io.Copy(buf, podLogs)
    if err != nil {
        return "error in copy information from podLogs to buf"
    }
}

// Keeping the reflection code here for now, but probably won't need it.
func discoverApis(cs *kubernetes.Clientset, namespace string) {
    // Create the meta-variable to examine the clientset variable.
	cst := reflect.TypeOf(cs)
    fmt.Println("The underlying type of cs: ", cst.Name())

    for i := 0; i < cst.NumMethod(); i++ {
		method := cst.Method(i)
		name := method.Name
		tp := method.Type
		fmt.Println("Name: ", name)
		fmt.Println("Type: ", tp)
		fmt.Println()
	}

    // Try to call several of the most promising ones and repeat the method discovery process:
    extensionsV1beta1 := cs.ExtensionsV1beta1()
    extensionsV1beta1Type := reflect.TypeOf(extensionsV1beta1)
    fmt.Println("Method set of the extensionsV1beta1 client subset:")
    for i := 0; i < extensionsV1beta1Type.NumMethod(); i++ {
        method := extensionsV1beta1Type.Method(i)
        name := method.Name
        tp := method.Type
        fmt.Println("Name: ", name)
        fmt.Println("Type: ", tp)
        fmt.Println()
	}

    restClient := cs.RESTClient()
    restClientType := reflect.TypeOf(restClient)
    fmt.Println("Method set of the RESTClient subset:")
    for i := 0; i < restClientType.NumMethod(); i++ {
        method := restClientType.Method(i)
        name := method.Name
        tp := method.Type
        fmt.Println("Name: ", name)
        fmt.Println("Type: ", tp)
        fmt.Println()
	}

    discoveryV1beta1 := cs.DiscoveryV1beta1()
    discoveryV1beta1Type := reflect.TypeOf(discoveryV1beta1)
    fmt.Println("Method set of the discoveryV1beta1 client subset:")
    for i := 0; i < discoveryV1beta1Type.NumMethod(); i++ {
        method := discoveryV1beta1Type.Method(i)
        name := method.Name
        tp := method.Type
        fmt.Println("Name: ", name)
        fmt.Println("Type: ", tp)
        fmt.Println()
	}
}
