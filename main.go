package main

import (
	"context"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig string
	fmt.Println("Start Controller")
	fmt.Println("Start Controller again")
	fmt.Println("third Controller again")
	fmt.Println("forth Controller again")
	
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		fmt.Println("Falling back to in-cluster config")

		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	thefoothebar := schema.GroupVersionResource{Group: "myk8s.io", Version: "v1", Resource: "thefoosthebars"}

	//setup informer

	informer := cache.NewSharedIndexInformer(

		// callbacks defining how to list and watch the resources, respectively
		&cache.ListWatch{
			// ListFunc initially populates informer's cache
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return dynClient.Resource(thefoothebar).Namespace("").List(context.TODO(), options)
			},
			// WatchFunc keeps cache updated with any changes
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return dynClient.Resource(thefoothebar).Namespace("").Watch(context.TODO(), options)
			},
		},

		// specifies that informer is for unstructured data
		// unstructured data represent any K8s resource without needing a predefined struct
		&unstructured.Unstructured{},

		// resync period of 0 means that informer will not resync the resources unless explicitly triggered
		0,
		cache.Indexers{},
	)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// called when new resource is created
		AddFunc: func(obj interface{}) {
			fmt.Println("Add event detected:", obj)
		},
		// called when existing resource is updated
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("Update event detected:", newObj)
		},
		// called when existing resource is deleted
		DeleteFunc: func(obj interface{}) {
			fmt.Println("Delete event detected:", obj)
		},
	})
	//  signal informer to stop watching for resource changes
	stop := make(chan struct{})
	defer close(stop)
	// starts informer in separate goroutine, allowing it to begin processing events
	go informer.Run(stop)

	// blocks until the informer's local cache is initially synced with current state of cluster
	// if it times out, program panics
	if !cache.WaitForCacheSync(stop, informer.HasSynced) {
		panic("Timeout waiting for cache sync")
	}

	fmt.Println("Custom Resource Controller started successfully")

}
