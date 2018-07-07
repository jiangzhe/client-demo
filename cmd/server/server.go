package main

import (
	"flag"
	"github.com/golang/glog"
	"client-demo/pkg/kube"
	"client-demo/pkg/rest"
	"net/http"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Set path of kubeconfig")
	flag.Parse()
	glog.V(4).Infof("kubeconfig is set to: %v", *kubeconfig)
	clientset, err := kube.NewClient(*kubeconfig)
	if err != nil {
		panic(err)
	}
	mirror := kube.DefaultMirrorWithAllResources(clientset)
	stopCh := make(chan struct{})
	mirror.Start(stopCh)
	glog.V(4).Infof("Kubernetes mirror has been synchronized")
	rest.SetupRestful(mirror)
	err = http.ListenAndServe(":8080", nil)
	glog.Fatal(err)
}
