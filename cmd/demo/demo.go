package main

import (
	"client-demo/pkg/kube"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
)

func main() {
	flag.Parse()

	clientset, err := kube.NewClient("/srv/kubernetes/kubeconfig")
	if err != nil {
		panic(err)
	}

	mirror := kube.NewMirror(clientset)
	mirror.AddResource(
		kube.ResourceDeployment,
		&extensionsv1beta1.Deployment{},
		func(clientset *kubernetes.Clientset, options metav1.ListOptions) (runtime.Object, error) {
			return clientset.ExtensionsV1beta1().Deployments(metav1.NamespaceAll).List(options)
		},
		func(clientset *kubernetes.Clientset, options metav1.ListOptions) (watch.Interface, error) {
			return clientset.ExtensionsV1beta1().Deployments(metav1.NamespaceAll).Watch(options)
		},
		0,
		kube.DefaultLoggingHandler,
		kube.DefaultIndexers())

	stopCh := mirror.Start()

	fmt.Printf("Deployments: %v\n", len(mirror.List(kube.ResourceDeployment)))
	fmt.Println("Deployments in tdcsys:")
	for _, deployment := range mirror.ListNamespaced(kube.ResourceDeployment, "tdcsys") {
		metadata, _ := meta.Accessor(deployment)
		fmt.Printf("%v\n", metadata.GetName())
	}

	fmt.Printf("Find indexed by namespace=tdcsys:\n")

	//if indexer, exists := mirror.GetIndexer(kube.ResourceDeployment); exists {
	//	//if keys, err := indexer.ByIndex(kube.IndexerNamespace, "tdcsys"); err != nil {
	//	//	for _, key := range keys {
	//	//		fmt.Printf("%v\n", key)
	//	//	}
	//	//}
	//
	//	for _, k := range indexer.ListIndexFuncValues(kube.IndexerNamespace) {
	//		fmt.Printf("%v\n", k)
	//		if k == "tdcsys" {
	//			data, _ := indexer.ByIndex(kube.IndexerNamespace, k)
	//			for _, d := range data {
	//				fmt.Printf("data=%v\n", d)
	//			}
	//		}
	//	}
	//}

	if indexer, exists := mirror.GetIndexer(kube.ResourceDeployment); exists {
		data, err := indexer.ByIndex(kube.IndexerOwner, "")
		if err != nil {
			panic(err)
		}

		for _, d := range data {
			metadata, _ := meta.Accessor(d)
			fmt.Printf("%v\n", metadata.GetName())
		}
	}


	stopCh<-struct{}{}
}
