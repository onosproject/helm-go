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
	"bytes"
	"github.com/onosproject/helm-go/pkg/helm/config"
	"github.com/onosproject/helm-go/pkg/helm/values"
	"github.com/onosproject/helm-go/pkg/kubernetes"
	"github.com/onosproject/helm-go/pkg/kubernetes/filter"
	"helm.sh/helm/v3/pkg/release"
	"time"
)

type Status string

const (
	// StatusUnknown indicates that a release is in an uncertain state.
	StatusUnknown Status = Status(release.StatusUnknown)
	// StatusDeployed indicates that the release has been pushed to Kubernetes.
	StatusDeployed Status = Status(release.StatusDeployed)
	// StatusUninstalled indicates that a release has been uninstalled from Kubernetes.
	StatusUninstalled Status = Status(release.StatusUninstalled)
	// StatusSuperseded indicates that this release object is outdated and a newer one exists.
	StatusSuperseded Status = Status(release.StatusSuperseded)
	// StatusFailed indicates that the release was not successfully deployed.
	StatusFailed Status = Status(release.StatusFailed)
	// StatusUninstalling indicates that a uninstall operation is underway.
	StatusUninstalling Status = Status(release.StatusUninstalling)
	// StatusPendingInstall indicates that an install operation is underway.
	StatusPendingInstall Status = Status(release.StatusPendingInstall)
	// StatusPendingUpgrade indicates that an upgrade operation is underway.
	StatusPendingUpgrade Status = Status(release.StatusPendingUpgrade)
	// StatusPendingRollback indicates that an rollback operation is underway.
	StatusPendingRollback Status = Status(release.StatusPendingRollback)
)

// StatusReport is Helm release status report
type StatusReport struct {
	Status        Status
	FirstDeployed time.Time
	LastDeployed  time.Time
}

// Release is a Helm release
type Release struct {
	StatusReport
	Namespace string
	Name      string
	values    *values.ImmutableValues
	client    kubernetes.Client
}

// Values returns the release values
func (r *Release) Values() *values.ImmutableValues {
	return r.values
}

// Client returns the release client
func (r *Release) Client() kubernetes.Client {
	return r.client
}

func getRelease(config *config.Config, release *release.Release) (*Release, error) {
	resources, err := config.KubeClient.Build(bytes.NewBufferString(release.Manifest), true)
	if err != nil {
		return nil, err
	}

	parent, err := kubernetes.NewForNamespace(release.Namespace)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewFiltered(release.Namespace, filter.Resources(parent, resources))
	if err != nil {
		return nil, err
	}

	values := values.New(release.Chart.Values).Override(values.New(release.Config))
	return &Release{
		StatusReport: StatusReport{
			Status:        Status(release.Info.Status),
			FirstDeployed: release.Info.FirstDeployed.Time,
			LastDeployed:  release.Info.LastDeployed.Time,
		},
		Namespace: release.Namespace,
		Name:      release.Name,
		values:    values.Immutable(),
		client:    client,
	}, nil
}
