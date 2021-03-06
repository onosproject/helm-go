// Code generated by generate-client. DO NOT EDIT.

package v2alpha1

import (
	"github.com/onosproject/helm-go/pkg/kubernetes/resource"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetes "k8s.io/client-go/kubernetes"
	"time"
)

type CronJobsReader interface {
	Get(name string) (*CronJob, error)
	List() ([]*CronJob, error)
}

func NewCronJobsReader(client resource.Client, filter resource.Filter) CronJobsReader {
	return &cronJobsReader{
		Client: client,
		filter: filter,
	}
}

type cronJobsReader struct {
	resource.Client
	filter resource.Filter
}

func (c *cronJobsReader) Get(name string) (*CronJob, error) {
	cronJob := &batchv2alpha1.CronJob{}
	client, err := kubernetes.NewForConfig(c.Config())
	if err != nil {
		return nil, err
	}
	err = client.BatchV2alpha1().
		RESTClient().
		Get().
		NamespaceIfScoped(c.Namespace(), CronJobKind.Scoped).
		Resource(CronJobResource.Name).
		Name(name).
		VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec).
		Timeout(time.Minute).
		Do().
		Into(cronJob)
	if err != nil {
		return nil, err
	} else {
		ok, err := c.filter(metav1.GroupVersionKind{
			Group:   CronJobKind.Group,
			Version: CronJobKind.Version,
			Kind:    CronJobKind.Kind,
		}, cronJob.ObjectMeta)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.NewNotFound(schema.GroupResource{
				Group:    CronJobKind.Group,
				Resource: CronJobResource.Name,
			}, name)
		}
	}
	return NewCronJob(cronJob, c.Client), nil
}

func (c *cronJobsReader) List() ([]*CronJob, error) {
	list := &batchv2alpha1.CronJobList{}
	client, err := kubernetes.NewForConfig(c.Config())
	if err != nil {
		return nil, err
	}
	err = client.BatchV2alpha1().
		RESTClient().
		Get().
		NamespaceIfScoped(c.Namespace(), CronJobKind.Scoped).
		Resource(CronJobResource.Name).
		VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec).
		Timeout(time.Minute).
		Do().
		Into(list)
	if err != nil {
		return nil, err
	}

	results := make([]*CronJob, 0, len(list.Items))
	for _, cronJob := range list.Items {
		ok, err := c.filter(metav1.GroupVersionKind{
			Group:   CronJobKind.Group,
			Version: CronJobKind.Version,
			Kind:    CronJobKind.Kind,
		}, cronJob.ObjectMeta)
		if err != nil {
			return nil, err
		} else if ok {
			copy := cronJob
			results = append(results, NewCronJob(&copy, c.Client))
		}
	}
	return results, nil
}
