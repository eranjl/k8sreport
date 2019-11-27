package main

// imports now include
import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

// report record

// Attribute represent a k8s attribute
type Attribute struct {
	Name      string
	Value     interface{}
	Flagged   bool
	Default   bool
	Set       bool
	Reason    string
	Container string
}

// Record represents a K8s object and all it's attributes
type Record struct {
	Ns         string
	Kind       string
	KindName   string
	ObjFlagged bool
	Attributes []Attribute
}

// Global Variables
var clientset *kubernetes.Clientset
var report []Record

// main function
func main() {

	// FFU - if we want to limit the report to a specific namespace
	var fNamespace string

	// connect to k8s
	connect(fNamespace)

	// all namespaces
	log.Printf("Getting Namespaces")
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get namespaces:", err)
	}

	for iNs, ns := range namespaces.Items {
		fmt.Printf("Working on Namespace [%d] - [%s]\n", iNs, ns.GetName())

		analyzeNS(ns)
	}
}

// connect to K8s
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
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Failed to connect to kubernetes. See ERROR")
		log.Fatal(err)
		os.Exit(1)
	}
}

// analyze namespace details
func analyzeNS(ns v1.Namespace) {
	// Get all deployments
	deployments, err := clientset.AppsV1().Deployments(ns.GetNamespace()).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	/*
		// Get all DaemonSets
		daemonSets, err := clientset.AppsV1().DaemonSets(ns.GetNamespace()).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get daemonsets:", err)
		}
		fmt.Println(daemonSets)

		// get all replicaSets
		replicaSets, err := clientset.AppsV1().ReplicaSets(ns.GetNamespace()).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get replicasets:", err)
		}
		fmt.Println(replicaSets)

		statefulSets, err := clientset.AppsV1().StatefulSets(ns.GetNamespace()).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get statefulsets:", err)
		}
		fmt.Println(statefulSets)
	*/

	for iDeployment, deployment := range deployments.Items {
		fmt.Printf("Working on deployment [%d] - [%s]\n", iDeployment, deployment.GetName())

		analyzeDeployment(deployment)
	}
}

// Get the deployment attributes and analyze them
func analyzeDeployment(deployment apps.Deployment) {
	log.Printf("Analyzing deployment [%s]", deployment.GetName())

	dRecord := Record{
		Ns:         deployment.Namespace,
		Kind:       "Deployment",
		KindName:   deployment.Name,
		ObjFlagged: false,
	}

	pSpec := deployment.Spec.Template.Spec

	{ // get pod template host network access
		pHostNetwork := NewAttribute("Host Network", pSpec.HostNetwork, "")

		if pHostNetwork.Value == (*bool)(nil) {
			dRecord.ObjFlagged = true
			pHostNetwork.Flagged = true
			pHostNetwork.Reason = "Host (node) network access is [NOT SET]"
		} else if pHostNetwork.Value == true {
			dRecord.ObjFlagged = true
			pHostNetwork.Flagged = true
			pHostNetwork.Set = true
			pHostNetwork.Default = false
			pHostNetwork.Reason = "Host (node) network access is set to [TRUE]"
		}

		// add attribute to record
		dRecord.Attributes = append(dRecord.Attributes, *pHostNetwork)
	}

	// if this is later overriden by the container level specific configuration than it's ok.
	{ // get pod template Run As User
		pRunAsUser := NewAttribute("Run As User", pSpec.SecurityContext.RunAsUser, "")

		// checking if the Run As User is defined at all
		if pRunAsUser.Value == (*int64)(nil) {
			dRecord.ObjFlagged = true
			pRunAsUser.Flagged = true
			pRunAsUser.Reason = "Run as user is not set"
		} else if pRunAsUser.Value == 0 { // checking if the defined user is root
			dRecord.ObjFlagged = true
			pRunAsUser.Flagged = true
			pRunAsUser.Set = true
			pRunAsUser.Default = false
			pRunAsUser.Reason = "Run as user is set to ROOT"
		}

		// add attribute to record
		dRecord.Attributes = append(dRecord.Attributes, *pRunAsUser)
	}

	{ // get pod template Must Run As NON Root
		pNonRoot := NewAttribute("Run As Non Root", pSpec.SecurityContext.RunAsNonRoot, "")

		if pNonRoot.Value == (*bool)(nil) {
			dRecord.ObjFlagged = true
			pNonRoot.Flagged = true
			pNonRoot.Reason = "Run as non root is not set"
		} else if pNonRoot.Value == false {
			dRecord.ObjFlagged = true
			pNonRoot.Flagged = true
			pNonRoot.Set = true
			pNonRoot.Default = false
			pNonRoot.Reason = "Run as non root is set to FALSE"
		}

		// add attribute to recorda
		dRecord.Attributes = append(dRecord.Attributes, *pNonRoot)
	}

	{ // get pod spec Run As Group
		pRunAsGroup := NewAttribute("Run As Group", pSpec.SecurityContext.RunAsGroup, "")

		if pRunAsGroup.Value == (*int64)(nil) {
			dRecord.ObjFlagged = true
			pRunAsGroup.Flagged = true
			pRunAsGroup.Reason = "Run as group is not set"
		} else if pRunAsGroup.Value == 0 {
			dRecord.ObjFlagged = true
			pRunAsGroup.Flagged = true
			pRunAsGroup.Set = true
			pRunAsGroup.Default = false
			pRunAsGroup.Reason = "Run as group is set to ROOT group"
		}

		// add attribute to record
		dRecord.Attributes = append(dRecord.Attributes, *pRunAsGroup)
	}

	{ // get pod spec Run As Group
		pSA := NewAttribute("POD ServiceAccount", pSpec.ServiceAccountName, "")

		if pSA.Value == (*string)(nil) || pSA.Value == "" {
			dRecord.ObjFlagged = true
			pSA.Flagged = true
			pSA.Reason = "Run as group is not set"
		} else if pSA.Value == "default" {
			dRecord.ObjFlagged = true
			pSA.Flagged = true
			pSA.Set = true
			pSA.Default = false
			pSA.Reason = "POD default service account is set to [default]"
		}

		// add attribute to record
		dRecord.Attributes = append(dRecord.Attributes, *pSA)
	}

	for _, container := range pSpec.Containers {
		fmt.Printf("Working on container [%s]\n", container.Name)

		{ // Image pull policy
			cPullPolicy := NewAttribute("Image Pull Policy", container.ImagePullPolicy, container.Name)

			// [TODO] we should use constants here instead
			if cPullPolicy.Value == (*string)(nil) || cPullPolicy.Value == "" {
				dRecord.ObjFlagged = true
				cPullPolicy.Flagged = true
				cPullPolicy.Reason = "ImagePullPolicy is not set"
			} else if cPullPolicy.Value == "Never" {
				dRecord.ObjFlagged = true
				cPullPolicy.Flagged = true
				cPullPolicy.Set = true
				cPullPolicy.Default = false
				cPullPolicy.Reason = "ImagePullPolicy is set to [Never]"
			} else if cPullPolicy.Value == "IfNotPresent" {
				dRecord.ObjFlagged = true
				cPullPolicy.Flagged = true
				cPullPolicy.Set = true
				cPullPolicy.Default = false
				cPullPolicy.Reason = "ImagePullPolicy is set to [IfNotPresent]"
			}

			// add attribute to record
			dRecord.Attributes = append(dRecord.Attributes, *cPullPolicy)
		}

		{ // Image name check
			cImage := NewAttribute("Image Name", container.Image, container.Name)

			// [TODO] check the image?
			// This is an impossible edge case (doing this just to cover it)
			if cImage.Value == (*string)(nil) || cImage.Value == "" {
				dRecord.ObjFlagged = true
				cImage.Flagged = true
				cImage.Reason = "Image is not set [?!]"
			} else if strings.HasSuffix(strings.ToLower(cImage.Value.(string)), ":latest") {
				dRecord.ObjFlagged = true
				cImage.Flagged = true
				cImage.Set = true
				cImage.Default = false
				cImage.Reason = "Image with [LATEST] tag is not recommended"
			}

			// add attribute to record
			dRecord.Attributes = append(dRecord.Attributes, *cImage)
		}

		{ // Container privileged attribute
			cPriv := NewAttribute("Privileged", container.SecurityContext.Privileged, container.Name)

			if cPriv.Value == (*bool)(nil) {
				dRecord.ObjFlagged = true
				cPriv.Flagged = true
				cPriv.Reason = "Container privileged is not defined & is not limited"
			} else if cPriv.Value == true {
				dRecord.ObjFlagged = true
				cPriv.Flagged = true
				cPriv.Set = true
				cPriv.Default = false
				cPriv.Reason = "Container set to allow [PRIVILEGED]"
			}

			// add attribute to record
			dRecord.Attributes = append(dRecord.Attributes, *cPriv)
		}

		{ // privileged escalation
			cPrivEsc := NewAttribute("Privilege Escalation", container.SecurityContext.AllowPrivilegeEscalation, container.Name)

			if cPrivEsc.Value == (*bool)(nil) {
				dRecord.ObjFlagged = true
				cPrivEsc.Flagged = true
				cPrivEsc.Reason = "Container privileged escalation is not defined & is not limited"
			} else if cPrivEsc.Value == true {
				dRecord.ObjFlagged = true
				cPrivEsc.Flagged = true
				cPrivEsc.Set = true
				cPrivEsc.Default = false
				cPrivEsc.Reason = "Container set to allow [PRIVILEGED ESCALATION]"
			}

			// add attribute to record
			dRecord.Attributes = append(dRecord.Attributes, *cPrivEsc)
		}

		{ // Read only root file system
			cROFS := NewAttribute("Read Only File System", container.SecurityContext.ReadOnlyRootFilesystem, container.Name)

			if cROFS.Value == (*bool)(nil) {
				dRecord.ObjFlagged = true
				cROFS.Flagged = true
				cROFS.Reason = "Container access to root file system is not defined & is not limited"
			} else if cROFS.Value == false {
				dRecord.ObjFlagged = true
				cROFS.Flagged = true
				cROFS.Set = true
				cROFS.Default = false
				cROFS.Reason = "Container access to root file system is [NOT READ ONLY]"
			}

			// add attribute to record
			dRecord.Attributes = append(dRecord.Attributes, *cROFS)
		}

		// ********************************************
		// TAKE CARE OF THE UID vs the DEPLOYMENT LEVEL
		// ********************************************
		{ // Allow privileged escalation
			cRunAsUser := NewAttribute("Container Run As User ", container.SecurityContext.RunAsUser, container.Name)

			if cRunAsUser.Value == (*int64)(nil) {
				dRecord.ObjFlagged = true
				cRunAsUser.Flagged = true
				cRunAsUser.Reason = "Container Run As User is not defined & is not limited"
			} else if cRunAsUser.Value == false {
				dRecord.ObjFlagged = true
				cRunAsUser.Flagged = true
				cRunAsUser.Set = true
				cRunAsUser.Default = false
				cRunAsUser.Reason = "Container access to root file system is [NOT READ ONLY]"
			}

			// add attribute to record
			dRecord.Attributes = append(dRecord.Attributes, *cRunAsUser)
		}

		/*
			if container.SecurityContext != nil {
				containerRAU := container.SecurityContext.RunAsUser

				if containerRAU == nil {
					// set the container RunAsUser to the Pod
					containerRAU = podRAU

				}
			}
		*/
	}

	/* FFU
	podSpec.HostPID
	podSpec.HostIPC
	*/
	/*
	 */
	/*
		if podSpec.ServiceAccountName == "default" {
							fmt.Printf(
							"%s,%s,%s,%s,%s\n",
							deployment.UID,
							podSpec.,
							ns.Name,
							"Deployment",
							"Pod running with [default] ServiceAccount",
						)
		}
	*/

	// TODO: get all matching labels and expressions - not sure how to use it in our analysis currently.

}

func createCSVReport(reportName string, records [][]string) {
	csvFile, err := os.Create(reportName + ".csv")

	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
		log.Println("File cannot be created for report")
		os.Exit(2)
	}

	// write headers to the report
	csvWriter := csv.NewWriter(csvFile)

	for _, record := range records {
		csvWriter.Write(record)
	}
	csvWriter.Flush()
}

// NewAttribute creates new attribute and set defaults
func NewAttribute(name string, value interface{}, container string) *Attribute {
	a := Attribute{
		Name:      name,
		Value:     value,
		Default:   true,
		Flagged:   false,
		Set:       false,
		Container: container,
	}

	return &a
}

// deployment/statefulset/daemonset podspec - all attributes
// podspec - container attributes
// pods all attributes (from runtime)
// containers - all attributes from running
