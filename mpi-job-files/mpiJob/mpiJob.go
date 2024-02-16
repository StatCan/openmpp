// A handler for mpi-based run requests to be called by the openm web service application.
// Accesses the kubernetes api endpoint for mpijobs, requests new mpijob, relays log information from the
// launcher pod as the job is running, and responds to an abort request if it is relayed by the web service.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
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
	// Set hardcoded namespace for now:
	namespace := "jacek-dev"
	// Set some sample argument values.
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
		"-OpenM.SetName",
		"Default",
		"-OpenM.SubValues",
		"32",
		"-OpenM.Threads",
		"16",
	}
	env := map[string]string{
		"SAMPLE_ENV": "VALUE",
	}

	// Create an MPIJob object using the sample arguments defined above.
	job := mpiJob(modelName, exeStem, dir, binDir, dbPath, mpiNp, args, env)

	// Validate spec using validation function.
	err := kubeAPI.ValidateV1MpiJobSpec(&job.Spec)
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("MPIJobSpec passed validation.")
	}

	// Create in-cluster configuration object.
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Cluster config object created.")
	}

	// Obtain client subset containing just the kubeflow based resources.
	kubeClientSubset, err := kubeClient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Kubeflow client subset obtained from config.")
	}

	// Obtain an interface to the MPIJobs collection for specified namespace.
	mpiJobs := kubeClientSubset.MPIJobs(namespace)

	// Submit request to create MPIJob.
	// It's confusing because an MPIJob is passed as an argument as well.
	// Presumably, the MPIJob being returned should have an active status,
	// whereas the one being submitted will not.
	_, err = mpiJobs.Create(context.TODO(), &job, meta.CreateOptions{})
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("MPIJob was successfully submitted.")
	}

	results, err := mpiJobs.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("MPIJobs collection obtained.")
	}

	// Display status of mpi jobs in collection.
	for _, r := range results.Items {
		fmt.Println("Job name: ", r.ObjectMeta.Name)
		for _, c := range r.Status.Conditions {
			fmt.Println("Condition type:", c.Type, " ... ", "Status", c.Status)
			fmt.Println("")
		}
	}

	// Obtain clientset with core resources. We will need it to access launcher pod logs.
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Clientset obtained from config.")
	}

	// Get name of launcher pod. It defaults to name of main container in launcher pod.
	name := job.Spec.MPIReplicaSpecs["Launcher"].Template.Spec.Containers[0].Name
	podLogOptions := core.PodLogOptions{}

	req := clientSet.CoreV1().Pods(namespace).GetLogs(name, &podLogOptions)

	// I want to check what type req is:
	tp := reflect.TypeOf(req)
	fmt.Println("Type of req: ", tp, " Name: ", tp.Name())

	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		panic(err.Error())
	}
	defer podLogs.Close()

	// Let's see if this passes the launcher logs stream to standard output:
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		panic(err.Error())
	}

	// Use os.Exit to terminate the program and return a status code.
	// In cases of error we will be able to send an error status code that the web service should pick up.
	os.Exit(0)
}

// Generate mpijob object based on arguments coming from openm web service and cluster configuration:
func mpiJob(modelName, exeStem, dir, binDir, dbPath string, mpiNp int32, args []string, env map[string]string) kubeAPI.MPIJob {

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
	mainContainerName := strings.Join([]string{strings.ToLower(modelName), timeStamp, "launcher"}, "-")
	launcherContainer := core.Container{
		Name:      mainContainerName,
		Image:     containerImage,
		Command:   containerCommand,
		Args:      containerArguments,
		Resources: launcherResourceRequirements,
	}

	// Worker container spec:
	workerContainer := core.Container{
		Name:      strings.Join([]string{strings.ToLower(modelName), timeStamp, "worker"}, "-"),
		Image:     containerImage,
		Resources: workerResourceRequirements,
	}

	// Launcher pod specs:
	launcherPodSpec := core.PodSpec{
		Containers: []core.Container{launcherContainer},
	}

	// Worker pod specs:
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

	// Pod template specifications:
	launcherPodTemplateSpec := core.PodTemplateSpec{
		ObjectMeta: podObjectMetadata,
		Spec:       launcherPodSpec,
	}

	workerPodTemplateSpec := core.PodTemplateSpec{
		ObjectMeta: podObjectMetadata,
		Spec:       workerPodSpec,
	}

	// Launcher and worker replica specs:
	var one int32 = 1
	launcherReplicaSpec := kubeAPI.ReplicaSpec{
		Replicas: &one,
		Template: launcherPodTemplateSpec,
	}

	workerReplicaSpec := kubeAPI.ReplicaSpec{
		Replicas: &mpiNp,
		Template: workerPodTemplateSpec,
	}

	// MPIJobSpec:
	var two int32 = 2
	var cleanPodPolicy kubeAPI.CleanPodPolicy = kubeAPI.CleanPodPolicyRunning
	mpiJobSpec := kubeAPI.MPIJobSpec{
		// 2 slots conforms to typical worker nodes used in the aaw cluster.
		SlotsPerWorker: &two,

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

	// Get hostname to annotate to MPIJob object.
	data, err := os.ReadFile("/etc/hostname")
	if err != nil {
		panic(err.Error())
	}
	hostname := strings.Replace(string(data), "\n", "", -1)

	// Type and object metadata:
	tm := meta.TypeMeta{
		Kind:       "MPIJob",
		APIVersion: "kubeflow.org/v1",
	}
	om := meta.ObjectMeta{
		Name: strings.Join([]string{strings.ToLower(modelName), timeStamp}, "-"),
		Labels: map[string]string{
			"notebook-name": hostname,
		},
	}

	// Construct MPIJob object:
	job := kubeAPI.MPIJob{
		TypeMeta:   tm,
		ObjectMeta: om,
		Spec:       mpiJobSpec,
	}
	return job
}
