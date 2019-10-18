package main

// imports now include
import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

// This program lists the pods in a cluster equivalent to
//
// kubectl get pods
//
func main() {

	// FFU - if we want to limit the report to a specific namespace
	var fNamespace string

	flag.StringVar(&fNamespace, "namespace", "", "namespace")

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// all namespaces
	log.Printf("Getting Namespaces")
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get namespaces:", err)
	}

	for iNs, ns := range namespaces.Items {
		fmt.Printf("Working on Namespace [%d] - [%s]\n", iNs, ns.GetName())

		// Get all deployments
		deployments, err := clientset.AppsV1().Deployments(ns.GetNamespace()).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get deployments:", err)
		}

		for iDeployment, deployment := range deployments.Items {
			fmt.Printf("Working on deployment [%d] - [%s]\n", iDeployment, deployment.GetName())

			// every deploytment is a set of Pods.
			// Getting the current deployment pod.

			//deploymentSpec := deployment.Spec

			podSpec := deployment.Spec.Template.Spec

			// get pod spec host network access
			podHNA := podSpec.HostNetwork
			if podHNA {
				fmt.Printf("***Pod Spec in deployment [%s] default HostNetwork is set to true\n", deployment.GetName())
			}

			// get pod spec Run As User
			podRAU := podSpec.SecurityContext.RunAsUser
			if podRAU == nil {
				fmt.Printf("***Pod Spec in deployment [%s] default RunAsUser is not set\n", deployment.GetName())
			}

			// get pod spec Run As Group
			podRAG := podSpec.SecurityContext.RunAsGroup
			if podRAG == nil {
				fmt.Printf("***Pod Spec in deployment [%s] default RunAsGroup is not set\n", deployment.GetName())
			}

			// get pod spec Run As Non Root User
			podRANonRoot := podSpec.SecurityContext.RunAsNonRoot
			if podRANonRoot == nil {
				fmt.Printf("***Pod Spec in deployment [%s] default RunAsNonRoot is not set\n", deployment.GetName())
			}

			podSA := podSpec.ServiceAccountName
			if podSA == "" {
				fmt.Printf("***Pod Spec in deployment [%s] default ServiceAccount is not set\n", deployment.GetName())
			}
			/* FFU
			podSpec.HostPID
			podSpec.HostIPC
			*/

			for iContainer, container := range podSpec.Containers {
				fmt.Printf("[%d] - [%s]\n", iContainer, container.Name)
			}
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
	}
	/*
		// getting pods
		log.Printf("Getting Pods")
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get pods:", err)
		}

		// print pods
		for i, pod := range pods.Items {

			// check the service account running the pod
			if pod.Spec.ServiceAccountName == "default" {
				fmt.Printf(
					"%d,%s,%s,%s,%s,%s\n",
					i,
					pod.ObjectMeta.OwnerReferences[0].UID,
					pod.GetName(),
					pod.GetNamespace(),
					pod.ObjectMeta.OwnerReferences[0].Kind,
					"Pod running with [default] ServiceAccount",
				)
				// TODO: how to get the UID of the parent
				// TODO: add the pod to an array of all accounts using default service accounts
			}

			if pod.GetNamespace() == "default" {
				fmt.Printf(
					"%d,%s,%s,%s,%s,%s\n",
					i,
					pod.ObjectMeta.OwnerReferences[0].UID,
					pod.GetName(),
					pod.GetNamespace(),
					pod.ObjectMeta.OwnerReferences[0].Kind,
					"Pod running in [default] namespace",
				)
				//TODO: add the pod to an array of all accounts using default service accounts
			}

			// TODO: is not set - will populate manually.
			fmt.Printf("Cluster Name: %s\n", pod.ObjectMeta.ClusterName)
			// uncomment to print all labels and annotations - FFU
			/*
				fmt.Println("---Labels---")
				for k, v := range pod.ObjectMeta.Labels {
					fmt.Printf("Key [%s] value [%s]\n", k, v)
				}

				fmt.Println("---Annotations---")
				for k, v := range pod.ObjectMeta.Annotations {
					fmt.Printf("Key [%s] value [%s]\n", k, v)
				}


			// iterating over the pod's volumes
			// TODO - find out how to iterate and get only those with Write access
			for i, vol := range pod.Spec.Volumes {
				fmt.Printf("Volume[%d] - Name [%s]\n", i, vol.Name)

				// TODO: need to identify the VolSource
			}
		}

			serviceAccounts, err := clientset.CoreV1().ServiceAccounts("").List(metav1.ListOptions{})

			// print service accounts
			for i, sa := range serviceAccounts.Items {
				fmt.Printf("[%d] %s\n", i, sa.GetName())
			}


		// all services
		log.Printf("Getting Services")
		services, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get services:", err)
		}

		for i, service := range services.Items {
			fmt.Printf("[%d] - [%s]\n", i, service.GetName())
		}
	*/
}
