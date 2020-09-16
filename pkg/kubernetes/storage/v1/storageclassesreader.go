// Code generated by generate-client. DO NOT EDIT.

package v1

import (
	"github.com/onosproject/helm-client/pkg/kubernetes/resource"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetes "k8s.io/client-go/kubernetes"
	"time"
)

type StorageClassesReader interface {
	Get(name string) (*StorageClass, error)
	List() ([]*StorageClass, error)
}

func NewStorageClassesReader(client resource.Client, filter resource.Filter) StorageClassesReader {
	return &storageClassesReader{
		Client: client,
		filter: filter,
	}
}

type storageClassesReader struct {
	resource.Client
	filter resource.Filter
}

func (c *storageClassesReader) Get(name string) (*StorageClass, error) {
	storageClass := &storagev1.StorageClass{}
	client, err := kubernetes.NewForConfig(c.Config())
	if err != nil {
		return nil, err
	}
	err = client.StorageV1().
		RESTClient().
		Get().
		NamespaceIfScoped(c.Namespace(), StorageClassKind.Scoped).
		Resource(StorageClassResource.Name).
		Name(name).
		VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec).
		Timeout(time.Minute).
		Do().
		Into(storageClass)
	if err != nil {
		return nil, err
	} else {
		ok, err := c.filter(metav1.GroupVersionKind{
			Group:   StorageClassKind.Group,
			Version: StorageClassKind.Version,
			Kind:    StorageClassKind.Kind,
		}, storageClass.ObjectMeta)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.NewNotFound(schema.GroupResource{
				Group:    StorageClassKind.Group,
				Resource: StorageClassResource.Name,
			}, name)
		}
	}
	return NewStorageClass(storageClass, c.Client), nil
}

func (c *storageClassesReader) List() ([]*StorageClass, error) {
	list := &storagev1.StorageClassList{}
	client, err := kubernetes.NewForConfig(c.Config())
	if err != nil {
		return nil, err
	}
	err = client.StorageV1().
		RESTClient().
		Get().
		NamespaceIfScoped(c.Namespace(), StorageClassKind.Scoped).
		Resource(StorageClassResource.Name).
		VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec).
		Timeout(time.Minute).
		Do().
		Into(list)
	if err != nil {
		return nil, err
	}

	results := make([]*StorageClass, 0, len(list.Items))
	for _, storageClass := range list.Items {
		ok, err := c.filter(metav1.GroupVersionKind{
			Group:   StorageClassKind.Group,
			Version: StorageClassKind.Version,
			Kind:    StorageClassKind.Kind,
		}, storageClass.ObjectMeta)
		if err != nil {
			return nil, err
		} else if ok {
			copy := storageClass
			results = append(results, NewStorageClass(&copy, c.Client))
		}
	}
	return results, nil
}