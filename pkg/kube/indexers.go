package kube

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
)
const (
	IndexerOwner = "owner"
	IndexerNamespace = "namespace"
)
func DefaultIndexers() cache.Indexers {
	return cache.Indexers{
		IndexerOwner: IndexerOwnerFunc,
		IndexerNamespace: IndexerNamespaceFunc,
	}
}

func IndexerOwnerFunc(obj interface{}) ([]string, error) {
	if str, ok := obj.(string); ok {
		return []string{str}, nil
	}
	metadata, err := meta.Accessor(obj)
	if err != nil {
		//return nil, err
		return []string{}, nil
	}
	namespace := metadata.GetNamespace()

	owners := metadata.GetOwnerReferences()
	if owners == nil || len(owners) == 0 {
		//return nil, fmt.Errorf("empty OwnerReferences")
		return []string{}, nil
	}
	var ownerNames []string
	for _, owner := range owners {
		ownerNames = append(ownerNames, namespace + "/" + owner.Name)
	}
	return ownerNames, nil
}

func IndexerNamespaceFunc(obj interface{}) ([]string, error) {
	if str, ok := obj.(string); ok {
		return []string{str}, nil
	}
	metadata, err := meta.Accessor(obj)
	if err != nil {
		//return nil, err
		return []string{}, nil
	}
	return []string{metadata.GetNamespace()}, nil
}