package kube

import (
	"k8s.io/client-go/tools/cache"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/meta"
)

var DefaultLoggingHandler = &cache.ResourceEventHandlerFuncs{
	AddFunc: logObjFunc("added"),
	UpdateFunc: logObj2Func("updated"),
	DeleteFunc: logObjFunc("deleted"),
}

func logObjFunc(event string) func(obj interface{}) {
	return func(obj interface{}) {
		objtype, err := meta.TypeAccessor(obj)
		if err != nil {
			glog.Warningf("Unknown object type with event %v: %v", event, obj)
		}
		metadata, err := meta.Accessor(obj)
		if err != nil {
			glog.Warningf("Unknown object metadata with event %v: %v", event, obj)
			return
		}
		glog.V(6).Infof("%v %v/%v event", objtype.GetKind(),  metadata.GetNamespace(), metadata.GetName(), event)
	}
}

func logObj2Func(event string) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		objtype, err := meta.TypeAccessor(newObj)
		if err != nil {
			glog.Warningf("Unknown object type with event %v: %v", event, newObj)
		}
		metadata, err := meta.Accessor(newObj)
		if err != nil {
			glog.Warningf("Unknown object metadata with event %v: %v", event, newObj)
			return
		}
		glog.V(6).Infof("%v %v/%v event", objtype.GetKind(),  metadata.GetNamespace(), metadata.GetName(), event)
	}
}