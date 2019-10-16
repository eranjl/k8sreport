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
	var ns string

	flag.StringVar(&ns, "namespace", "", "namespace")

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

	// getting pods
	log.Printf("Getting Pods")
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get pods:", err)
	}

	// print pods
	for i, pod := range pods.Items {

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
			//TODO: add the pod to an array of all accounts using default service accounts
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

		// is not set - will populate manually.
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
		*/

		// iterating over the pod's volumes
		// TODO - find out how to iterate and get only those with Write access
		for i, vol := range pod.Spec.Volumes {
			fmt.Printf("Volume[%d] - Name [%s]\n", i, vol.Name)
		}
	}
	/*
		serviceAccounts, err := clientset.CoreV1().ServiceAccounts("").List(metav1.ListOptions{})

		// print service accounts
		for i, sa := range serviceAccounts.Items {
			fmt.Printf("[%d] %s\n", i, sa.GetName())
		}
	*/
}
