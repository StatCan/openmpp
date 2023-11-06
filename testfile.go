 package main

import (
     "log"
     "os"
     "text/template"
)

type User struct {
     Name       string
     Occupation string
}

type MPITemplateInfo struct {
    JobName          string
    Spw              string
    TTL              string
    SSHAMP           string
    LauncherReplicas string
    LauncherImageName string
    ContainerLauncher string
    LauncherUserID   string
    ContainerPath    string
    LauncherCPUs     string
    LauncherMemSize  string
    WorkerReplicas   string
    WorkerImageName  string
    WorkerLauncher   string
    WorkerUserID     string
    WorkerCPUs       string
    WorkerMemSize    string
}

func NewTempl(JobName string) MPITemplateInfo {
    return MPITemplateInfo{
        JobName: JobName,
        Spw: "1", //Slots per worker
        TTL: "90", // TTL seconds after finish
        SSHAMP: "", //   sshAuthMountPath
        LauncherReplicas: "1", 
        LauncherImageName: "",
        ContainerLauncher: "", 
        LauncherUserID: "", 
        ContainerPath: "", 
        LauncherCPUs: "2",  
        LauncherMemSize: "2Gi",
        WorkerReplicas: "1", 
        WorkerImageName: "", 
        WorkerLauncher: "", 
        WorkerUserID: "", 
        WorkerCPUs: "2",  
        WorkerMemSize: "2Gi", 
    }
} // end MPITemplateInfo struct constructor

func main() {

     //user := User{"John Doe", "gardener"}
     jobID := "Job123"
     templ := NewTempl(jobID)

     //tmp, err := template.ParseFiles("message.txt") //mpi-template-go.yaml
     tmp, err := template.ParseFiles("mpi-template-go.yaml") //

     if err != nil {
         log.Fatal(err)
     }

     //err2 := tmp.Execute(os.Stdout, user)

     //if err2 != nil {
     //    log.Fatal(err2)
     //}


     // create a new file
     jobName := jobID+".yaml"
     file, _ := os.Create(jobName)
     defer file.Close()

     // apply the template to the vars map and write the result to file.
     tmp.Execute(file, templ)

}
