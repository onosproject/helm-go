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

package release

import (
	"errors"
	"github.com/onosproject/helm-go/pkg/helm/config"
	"github.com/onosproject/helm-go/pkg/helm/values"
	"helm.sh/helm/v3/pkg/release"
)

// NewClient returns a new release client
func NewClient(config *config.Config) Client {
	return &releaseClient{
		config: config,
	}
}

// Client is a Helm release client
type Client interface {
	// Namespace returns the release client namespace
	Namespace() string
	// Get gets a release
	Get(name string) (*Release, error)
	// List lists releases
	List() ([]*Release, error)
	// Status gets the status of a release
	Status(name string) (StatusReport, error)
	// Install installs a release
	Install(release string, chart string) *InstallRequest
	// Uninstall uninstalls a release
	Uninstall(release string) *UninstallRequest
	// Upgrade upgrades a release
	Upgrade(release string, chart string) *UpgradeRequest
	// Rollback rolls back a release
	Rollback(release string) *RollbackRequest
}

// releaseClient is the Helm release client
type releaseClient struct {
	config *config.Config
}

func (c *releaseClient) Namespace() string {
	return c.config.Namespace()
}

// Get gets a release
func (c *releaseClient) Get(name string) (*Release, error) {
	list, err := c.config.Releases.List(func(r *release.Release) bool {
		return r.Namespace == c.config.Namespace() && r.Name == name
	})
	if err != nil {
		return nil, err
	} else if len(list) == 0 {
		return nil, errors.New("release not found")
	} else if len(list) > 1 {
		return nil, errors.New("release is ambiguous")
	}
	return getRelease(c.config, list[0])
}

// List lists releases
func (c *releaseClient) List() ([]*Release, error) {
	list, err := c.config.Releases.List(func(r *release.Release) bool {
		return r.Namespace == c.config.Namespace()
	})
	if err != nil {
		return nil, err
	}

	releases := make([]*Release, len(list))
	for i, release := range list {
		r, err := getRelease(c.config, release)
		if err != nil {
			return nil, err
		}
		releases[i] = r
	}
	return releases, nil
}

// Status gets the status of a release
func (c *releaseClient) Status(name string) (StatusReport, error) {
	release, err := c.Get(name)
	if err != nil {
		return StatusReport{}, err
	}
	return release.StatusReport, nil
}

// Install installs a release
func (c *releaseClient) Install(release string, chart string) *InstallRequest {
	return &InstallRequest{
		client: c,
		config: c.config,
		name:   release,
		chart:  chart,
		values: values.New(),
	}
}

// Uninstall uninstalls a release
func (c *releaseClient) Uninstall(release string) *UninstallRequest {
	return &UninstallRequest{
		client: c,
		config: c.config,
		name:   release,
	}
}

// Upgrade upgrades a release
func (c *releaseClient) Upgrade(release string, chart string) *UpgradeRequest {
	return &UpgradeRequest{
		client: c,
		config: c.config,
		name:   release,
		chart:  chart,
		values: values.New(),
	}
}

// Rollback rolls back a release
func (c *releaseClient) Rollback(release string) *RollbackRequest {
	return &RollbackRequest{
		client: c,
		config: c.config,
		name:   release,
	}
}

var _ Client = &releaseClient{}
