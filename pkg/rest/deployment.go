package rest

import (
	"github.com/emicklei/go-restful"
	"client-demo/pkg/kube"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"github.com/emicklei/go-restful-openapi"
	"fmt"
)

type deploymentController struct {
	mirror kube.K8sMirror
}

func NewDeploymentController(mirror kube.K8sMirror) *deploymentController {
	return &deploymentController{
		mirror:mirror,
	}
}

func (dc *deploymentController) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/api/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/deployments").To(dc.List).
		Doc("List deployments").
		Metadata(restfulspec.KeyOpenAPITags, []string{"deployments"}).
		Writes([]extensionsv1beta1.Deployment{}))

	ws.Route(ws.GET("/namespaces/:namespace/deployments/:name").To(dc.Get).
		Doc("Find a deployment").
		Metadata(restfulspec.KeyOpenAPITags, []string{"deployments"}).
		Writes(extensionsv1beta1.Deployment{}))

	return ws
}

func (dc *deploymentController) List(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity(dc.mirror.List(kube.ResourceDeployment))
}

func (dc *deploymentController) Get(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")
	key := kube.NamespaceKeyFunc(namespace, name)
	item, exists, err := dc.mirror.GetByKey(kube.ResourceDeployment, key)
	if err != nil {
		resp.WriteServiceError(500, restful.ServiceError{500, err.Error()})
		return
	}
	if !exists {
		resp.WriteServiceError(404, restful.ServiceError{404, fmt.Sprintf("Deployment %v not found", key)})
		return
	}
	resp.WriteEntity(item)
}