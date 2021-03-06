// Code generated by generate-client. DO NOT EDIT.

package v1

import (
	"github.com/onosproject/helm-go/pkg/kubernetes/resource"
)

type ClusterRolesClient interface {
	ClusterRoles() ClusterRolesReader
}

func NewClusterRolesClient(resources resource.Client, filter resource.Filter) ClusterRolesClient {
	return &clusterRolesClient{
		Client: resources,
		filter: filter,
	}
}

type clusterRolesClient struct {
	resource.Client
	filter resource.Filter
}

func (c *clusterRolesClient) ClusterRoles() ClusterRolesReader {
	return NewClusterRolesReader(c.Client, c.filter)
}
