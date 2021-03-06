// Code generated by generate-client. DO NOT EDIT.

package v1beta1

import (
	"github.com/onosproject/helm-go/pkg/kubernetes/resource"
)

type StatefulSetsClient interface {
	StatefulSets() StatefulSetsReader
}

func NewStatefulSetsClient(resources resource.Client, filter resource.Filter) StatefulSetsClient {
	return &statefulSetsClient{
		Client: resources,
		filter: filter,
	}
}

type statefulSetsClient struct {
	resource.Client
	filter resource.Filter
}

func (c *statefulSetsClient) StatefulSets() StatefulSetsReader {
	return NewStatefulSetsReader(c.Client, c.filter)
}
