// Copyright 2020-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"github.com/onosproject/helm-go/pkg/helm/chart"
	helmconfig "github.com/onosproject/helm-go/pkg/helm/config"
	"github.com/onosproject/helm-go/pkg/helm/release"
	"github.com/onosproject/helm-go/pkg/helm/repo"
	"github.com/onosproject/helm-go/pkg/kubernetes/config"
)

// DefaultNamespace is the default Helm namespace
var DefaultNamespace = config.GetNamespaceFromEnv()

// New creates a new Helm client for the given namespace
func New(namespace ...string) (Helm, error) {
	ns := DefaultNamespace
	if len(namespace) > 0 {
		ns = namespace[0]
	}
	config, err := helmconfig.GetConfig(ns)
	if err != nil {
		return nil, err
	}
	return &helmClient{
		namespace: ns,
		repos:     repo.NewClient(config),
		charts:    chart.NewClient(config),
		releases:  release.NewClient(config),
	}, nil
}

// Helm is a Helm client
type Helm interface {
	// Namespace returns the Helm namespace
	Namespace() string
	// Repos returns the repository client
	Repos() repo.Client
	// Charts returns the chart client
	Charts() chart.Client
	// Releases returns the release client
	Releases() release.Client
	// Install installs a chart
	Install(name string, chart string) *release.InstallRequest
	// Uninstall uninstalls a chart
	Uninstall(name string) *release.UninstallRequest
	// Upgrade upgrades a release
	Upgrade(name string, chart string) *release.UpgradeRequest
	// Rollback rolls back a release
	Rollback(name string) *release.RollbackRequest
}

// helmClient is the default implementation of the Helm Client
type helmClient struct {
	namespace string
	repos     repo.Client
	charts    chart.Client
	releases  release.Client
}

func (c *helmClient) Namespace() string {
	return c.namespace
}

func (c *helmClient) Repos() repo.Client {
	return c.repos
}

func (c *helmClient) Charts() chart.Client {
	return c.charts
}

func (c *helmClient) Releases() release.Client {
	return c.releases
}

func (c *helmClient) Install(name string, chart string) *release.InstallRequest {
	return c.releases.Install(name, chart)
}

func (c *helmClient) Uninstall(name string) *release.UninstallRequest {
	return c.releases.Uninstall(name)
}

func (c *helmClient) Upgrade(name string, chart string) *release.UpgradeRequest {
	return c.releases.Upgrade(name, chart)
}

func (c *helmClient) Rollback(name string) *release.RollbackRequest {
	return c.releases.Rollback(name)
}

var _ Helm = &helmClient{}
