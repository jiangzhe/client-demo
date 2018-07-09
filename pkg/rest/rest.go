package rest

import "github.com/emicklei/go-restful"

type Controller interface {
	WebService() *restful.WebService
}
