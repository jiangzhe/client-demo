package main

import (
	"client-demo/pkg/rest"
	"client-demo/pkg/kube"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	"net/http"
	"flag"
	"fmt"
	"context"
	"k8s.io/client-go/kubernetes"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func main() {
	portPtr := flag.Int("port", 8080, "Set the server port")
	kubeconfigPtr := flag.String("kubeconfig", "/srv/kubernetes/kubeconfig", "Set path to kubeconfig")
	flag.Parse()

	clientset, err := kube.NewClient(*kubeconfigPtr)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	mirror.StartUntilSynced(ctx)

	dc := rest.NewDeploymentController(mirror)

	restful.DefaultContainer.Add(dc.WebService())

	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath: "/apidocs.json",
		APIVersion: "1.0.0",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	glog.Infoln("Start listening on localhost:8080")
	err = http.ListenAndServe(fmt.Sprintf(":%v", *portPtr), nil)
	glog.Fatal(err)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title: "Client Demo Service",
			Description: "Manage kubernetes resources",
			Contact: &spec.ContactInfo{
				Name: "jiangzhe",
				Email: "zhe.jiang@transwarp.io",
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{
		spec.Tag{TagProps: spec.TagProps{
			Name: "deployments",
			Description: "Manage deployments",
		}},
	}
}