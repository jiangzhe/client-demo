package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"fmt"
	//extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	//"k8s.io/apimachinery/pkg/api/meta"
	"time"
	"github.com/golang/glog"
	"context"
)

//func main() {
//	mirror, err := NewMirror("/srv/kubernetes/kubeconfig")
//	if err != nil {
//		panic(err)
//	}
//	stopCh := mirror.Start()
//
//	for _, obj := range mirror.List(ResourceDeployment) {
//		metadata, err := meta.Accessor(obj)
//		if err != nil {
//			panic(err)
//		}
//		fmt.Printf("object name is: %v\n", metadata.GetName())
//	}
//	stopCh<-struct{}{}
//}

// mirror is local store of kubernetes resources,
// support list all resources
type K8sMirror interface {
	AddResource(resource ResourceType, objType runtime.Object, listFunc ListFunc, watchFunc WatchFunc,
		resyncPeriod time.Duration, handler cache.ResourceEventHandler, indexers cache.Indexers)

	Start(ctx context.Context)
	StartUntilSynced(ctx context.Context)
	List(resource ResourceType) []interface{}
	ListNamespaced(resource ResourceType, namespace string) []interface{}
	ListKeys(resource ResourceType) []string
	GetByKey(resource ResourceType, key string) (item interface{}, exists bool, err error)
	GetIndexer(resource ResourceType) (indexer cache.Indexer, exists bool)
}

type ResourceType string
const (
	ResourceNamespace = "namespace"
	ResourcePod = "pod"
	ResourceDeployment = "deployment"
	ResourceStatefulset = "statefulset"
	ResourceService = "service"
	ResourceConfigmap = "configmap"
)

type mirror struct {
	clientset *kubernetes.Clientset
	informers map[ResourceType]cache.SharedIndexInformer
}

// inject clientset to cache.ListFunc
type ListFunc func(clientset *kubernetes.Clientset, options metav1.ListOptions) (runtime.Object, error)
// inject clientset to cache.WatchFunc
type WatchFunc func(clientset *kubernetes.Clientset, options metav1.ListOptions) (watch.Interface, error)

func (m *mirror) AddResource(
	resource ResourceType,
	objType runtime.Object,
	listFunc ListFunc,
	watchFunc WatchFunc,
	resyncPeriod time.Duration,
	handler cache.ResourceEventHandler, // work with queue using the event handler
	indexers cache.Indexers) {
		indexInformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return listFunc(m.clientset, options)
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return watchFunc(m.clientset, options)
				},
			},
			objType,
			resyncPeriod,
			indexers)
		indexInformer.AddEventHandler(handler)
		m.informers[resource] = indexInformer
}

func (m *mirror) Start(ctx context.Context) {
	for _, informer := range m.informers {
		go informer.Run(ctx.Done())
	}

}

// block until all resources are synced
func (m *mirror) StartUntilSynced(ctx context.Context) {
	m.Start(ctx)
	for resource, informer := range m.informers {
		if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
			utilruntime.HandleError(fmt.Errorf("timeout waiting for %v to sync", resource))
		} else {
			glog.V(4).Infof("Resource %v synced\n", resource)
		}
	}
}


func (m *mirror) List(resource ResourceType) []interface{} {
	informer, exists := m.informers[resource]
	if !exists {
		return []interface{}{}
	}
	return informer.GetStore().List()
}

func (m *mirror) ListNamespaced(resource ResourceType, namespace string) []interface{} {
	informer, exists := m.informers[resource]
	if !exists {
		return []interface{}{}
	}
	results, err := informer.GetIndexer().ByIndex(IndexerNamespace, namespace)
	if err != nil {
		return []interface{}{}
	}
	return results
}

func (m *mirror) ListKeys(resource ResourceType) []string {
	informer, exists := m.informers[resource]
	if !exists {
		return []string{}
	}
	return informer.GetStore().ListKeys()
}

func (m *mirror) GetByKey(resource ResourceType, key string) (item interface{}, exists bool, err error) {
	informer, exists := m.informers[resource]
	if !exists {
		return
	}
	item, exists, err = informer.GetStore().GetByKey(key)
	return
}

func (m *mirror) GetIndexer(resource ResourceType) (indexer cache.Indexer, exists bool) {
	informer, exists := m.informers[resource]
	if !exists {
		return
	}
	return informer.GetIndexer(), true
}

func NewMirror(clientset *kubernetes.Clientset) K8sMirror {
	return &mirror{
		clientset: clientset,
		informers: map[ResourceType]cache.SharedIndexInformer{},
	}
}

func NewClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NamespaceKeyFunc(namespace string, name string) string {
	return namespace + "/" + name
}