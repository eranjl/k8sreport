package main

// imports now include
import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

// Global Variables
var cs *kubernetes.Clientset
var rawData []Record

// main function
func main() {

	// FFU - if we want to limit the report to a specific namespace
	var fNamespace string

	// connect to k8s
	connect(fNamespace)

	// all namespaces
	log.Printf("Getting Namespaces")
	namespaces, err := cs.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get namespaces:", err)
	}

	for i, ns := range namespaces.Items {
		fmt.Printf("Working on Namespace [%d] - [%s]\n", i, ns.GetName())

		analyzeNS(ns)
	}

	createCSVReport("apolicy.k8s.report")
}

// Connect to K8s
func connect(fNamespace string) {
	flag.StringVar(&fNamespace, "namespace", "", "namespace")

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	cs, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Failed to connect to kubernetes. See ERROR")
		log.Fatal(err)
		os.Exit(1)
	}
}

// analyze namespace details
func analyzeNS(ns v1.Namespace) {
	// Get all deployments
	deployments, err := cs.AppsV1().Deployments(ns.GetName()).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	for i, deployment := range deployments.Items {
		fmt.Printf("Working on Deployment [%d] - [%s]\n", i, deployment.GetName())

		analyzeDep(deployment)
	}

	// Get all DaemonSets
	daemonSets, err := cs.AppsV1().DaemonSets(ns.GetName()).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get daemonsets:", err)
	}

	for i, ds := range daemonSets.Items {
		fmt.Printf("Working on DaemonSet [%d] - [%s]\n", i, ds.GetName())

		analyzeDS(ds)
	}

	statefulSets, err := cs.AppsV1().StatefulSets(ns.GetName()).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get statefulsets:", err)
	}

	for i, ss := range statefulSets.Items {
		fmt.Printf("Working on StatefulSet [%d] - [%s]\n", i, ss.GetName())

		analyzeSS(ss)
	}

	/* Need to handle replicaSets differently
	replicaSets, err := cs.AppsV1().ReplicaSets(ns.GetName()).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get replicasets:", err)
	}

	for i, rs := range replicaSets.Items {
		fmt.Printf("Working on ReplicaSet [%d] - [%s]\n", i, rs.GetName())

		analyzeRS(rs)
	}
	*/

	jobs, err := cs.BatchV1().Jobs(ns.GetName()).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get jobs:", err)
	}

	for i, job := range jobs.Items {
		fmt.Printf("Working on ReplicaSet [%d] - [%s]\n", i, job.GetName())

		analyzeJob(job)
	}
}

func analyzeJob(j batch.Job) {
	log.Printf("Analyzing Job [%s]", j.GetName())

	jData := Record{
		Cluster:  j.GetClusterName(),
		Ns:       j.GetNamespace(),
		Kind:     "Job",
		KindName: j.GetName(),
	}

	podSpec := j.Spec.Template.Spec

	collectPodSpec(podSpec, &jData)

	for _, container := range podSpec.Containers {

		collectContainerPodSpec(container, &jData)
	}

	rawData = append(rawData, jData)

}

func analyzeRS(rs apps.ReplicaSet) {
	log.Printf("Analyzing ReplicaSet [%s]", rs.GetName())

	rsData := Record{
		Cluster:  rs.GetClusterName(),
		Ns:       rs.GetNamespace(),
		Kind:     "ReplicaSet",
		KindName: rs.GetName(),
	}

	podSpec := rs.Spec.Template.Spec

	collectPodSpec(podSpec, &rsData)

	for _, container := range podSpec.Containers {

		collectContainerPodSpec(container, &rsData)
	}

	rawData = append(rawData, rsData)

}

func analyzeSS(ss apps.StatefulSet) {
	log.Printf("Analyzing StatefulSet [%s]", ss.GetName())

	ssData := Record{
		Cluster:  ss.GetClusterName(),
		Ns:       ss.GetNamespace(),
		Kind:     "StatefulSet",
		KindName: ss.GetName(),
	}

	podSpec := ss.Spec.Template.Spec

	collectPodSpec(podSpec, &ssData)

	for _, container := range podSpec.Containers {

		collectContainerPodSpec(container, &ssData)
	}

	rawData = append(rawData, ssData)

}

func analyzeDS(ds apps.DaemonSet) {
	log.Printf("Analyzing DaemonSet [%s]", ds.GetName())

	dsData := Record{
		Cluster:  ds.GetClusterName(),
		Ns:       ds.GetNamespace(),
		Kind:     "DaemonSet",
		KindName: ds.GetName(),
	}

	podSpec := ds.Spec.Template.Spec

	collectPodSpec(podSpec, &dsData)

	for _, container := range podSpec.Containers {

		collectContainerPodSpec(container, &dsData)
	}

	rawData = append(rawData, dsData)

}

// Get the deployment attributes and analyze them
func analyzeDep(deployment apps.Deployment) {
	log.Printf("Analyzing Deployment [%s]", deployment.GetName())

	depData := Record{
		Cluster:  deployment.GetClusterName(),
		Ns:       deployment.GetNamespace(),
		Kind:     "Deployment",
		KindName: deployment.GetName(),
	}

	podSepc := deployment.Spec.Template.Spec

	collectPodSpec(podSepc, &depData)

	for _, container := range podSepc.Containers {

		collectContainerPodSpec(container, &depData)
	}

	// TODO: get all matching labels and expressions - not sure how to use it in our analysis currently.

	rawData = append(rawData, depData)
}
