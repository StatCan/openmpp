// Moved type and function declarations that are not being used into this separate file because
// go compiler throws errors when building a package with unused declarations.

// But we could try this suggestion also. This is a function with any empty body that accepts
// any number of arguments of any type. Invoking this function once in our main() and passing 
// it all the unused variables takes care of all the unused variable build errors.
func UNUSED(x ...interface{}) {}

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

type MPIJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MPIJob `json:"items"`

type MPIJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MPIJobSpec `json:"spec,omitempty"`
	Status            JobStatus  `json:"status,omitempty"`

type JobStatus struct {
	// Conditions is an array of current observed job conditions.
	Conditions []JobCondition `json:"conditions,omitempty"`

	// ReplicaStatuses is map of ReplicaType and ReplicaStatus,
	// specifies the status of each replica.
	ReplicaStatuses map[ReplicaType]*ReplicaStatus `json:"replicaStatuses,omitempty"`

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Represents last time when the job was reconciled. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

type JobCondition struct {
	// Type of job condition.
	Type JobConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

const (
	// JobCreated means the job has been accepted by the system,
	// but one or more of the pods/services has not been started.
	// This includes time before pods being scheduled and launched.
	JobCreated JobConditionType = "Created"

	// JobRunning means all sub-resources (e.g. services/pods) of this job
	// have been successfully scheduled and launched.
	// The training is running without error.
	JobRunning JobConditionType = "Running"

	// JobRestarting means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed but maybe restarted according to it's restart policy
	// which specified by user in v1.PodTemplateSpec.
	// The training is freezing/pending.
	JobRestarting JobConditionType = "Restarting"

	// JobSucceeded means all sub-resources (e.g. services/pods) of this job
	// reached phase have terminated in success.
	// The training is complete without error.
	JobSucceeded JobConditionType = "Succeeded"

	// JobSuspended means the job has been suspended.
	JobSuspended JobConditionType = "Suspended"

	// JobFailed means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed with no restarting.
	// The training has failed its execution.
	JobFailed JobConditionType = "Failed"
)
