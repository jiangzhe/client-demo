package rest

import (
	"github.com/emicklei/go-restful"
	"client-demo/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

type KubeRest struct {
	mirror kube.K8sMirror
}

func (kr *KubeRest) ListNamespaces(req *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(kr.mirror.List(kube.ResourceNamespace))
}

func SetupRestful(mirror kube.K8sMirror) {
	kubeRest := &KubeRest{mirror}
	ws := new(restful.WebService)
	ws.Path("/api/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/namespaces").To(kubeRest.ListNamespaces).Doc("List all namespaces").Writes([]corev1.Namespace{}))
}