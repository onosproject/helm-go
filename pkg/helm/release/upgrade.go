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
	"github.com/onosproject/helm-go/pkg/helm/config"
	"github.com/onosproject/helm-go/pkg/helm/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/storage/driver"
	"os"
	"time"
)

// UpgradeRequest is a release upgrade request
type UpgradeRequest struct {
	client       Client
	config       *config.Config
	name         string
	chart        string
	repo         string
	caFile       string
	keyFile      string
	certFile     string
	username     string
	password     string
	version      string
	values       *values.Values
	disableHooks bool
	dryRun       bool
	atomic       bool
	wait         bool
	timeout      time.Duration
}

func (r *UpgradeRequest) CaFile(caFile string) *UpgradeRequest {
	r.caFile = caFile
	return r
}

func (r *UpgradeRequest) KeyFile(keyFile string) *UpgradeRequest {
	r.keyFile = keyFile
	return r
}

func (r *UpgradeRequest) CertFile(certFile string) *UpgradeRequest {
	r.certFile = certFile
	return r
}

func (r *UpgradeRequest) Username(username string) *UpgradeRequest {
	r.username = username
	return r
}

func (r *UpgradeRequest) Password(password string) *UpgradeRequest {
	r.password = password
	return r
}

func (r *UpgradeRequest) Repo(url string) *UpgradeRequest {
	r.repo = url
	return r
}

func (r *UpgradeRequest) Version(version string) *UpgradeRequest {
	r.version = version
	return r
}

func (r *UpgradeRequest) Set(path string, value interface{}) *UpgradeRequest {
	r.values.Set(path, value)
	return r
}

func (r *UpgradeRequest) DisableHooks() *UpgradeRequest {
	r.disableHooks = true
	return r
}

func (r *UpgradeRequest) DryRun() *UpgradeRequest {
	r.dryRun = true
	return r
}

func (r *UpgradeRequest) Atomic() *UpgradeRequest {
	r.atomic = true
	return r
}

func (r *UpgradeRequest) Wait() *UpgradeRequest {
	r.wait = true
	return r
}

func (r *UpgradeRequest) Timeout(timeout time.Duration) *UpgradeRequest {
	r.timeout = timeout
	return r
}

func (r *UpgradeRequest) Do() (*Release, error) {
	upgrade := action.NewUpgrade(r.config.Configuration)

	// Setup the repo options
	upgrade.RepoURL = r.repo
	upgrade.Username = r.username
	upgrade.Password = r.password
	upgrade.CaFile = r.caFile
	upgrade.KeyFile = r.keyFile
	upgrade.CertFile = r.certFile

	// Setup the chart options
	upgrade.Version = r.version

	// Setup the release options
	upgrade.Namespace = r.client.Namespace()
	upgrade.Atomic = r.atomic
	upgrade.DryRun = r.dryRun
	upgrade.DisableHooks = r.disableHooks
	upgrade.Wait = r.wait
	upgrade.Timeout = r.timeout

	// Locate the chart path
	path, err := upgrade.ChartPathOptions.LocateChart(r.chart, r.config.EnvSettings)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chart, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	ctx := config.GetContext().Release(r.name)
	values := r.values.Normalize().Override(values.New(ctx.Values.Values()))

	if upgrade.Install {
		// If a release does not exist, install it. If another error occurs during
		// the check, ignore the error and continue with the upgrade.
		histClient := action.NewHistory(r.config.Configuration)
		histClient.Max = 1
		if _, err := histClient.Run(r.name); err == driver.ErrReleaseNotFound {
			install := action.NewInstall(r.config.Configuration)
			install.ChartPathOptions = upgrade.ChartPathOptions
			install.DryRun = upgrade.DryRun
			install.DisableHooks = upgrade.DisableHooks
			install.Timeout = upgrade.Timeout
			install.Wait = upgrade.Wait
			install.Devel = upgrade.Devel
			install.Namespace = upgrade.Namespace
			install.Atomic = upgrade.Atomic
			install.PostRenderer = upgrade.PostRenderer

			if req := chart.Metadata.Dependencies; req != nil {
				// If CheckDependencies returns an error, we have unfulfilled dependencies.
				// As of Helm 2.4.0, this is treated as a stopping condition:
				// https://github.com/helm/helm/issues/2209
				if err := action.CheckDependencies(chart, req); err != nil {
					if install.DependencyUpdate {
						man := &downloader.Manager{
							Out:              os.Stdout,
							ChartPath:        path,
							Keyring:          install.ChartPathOptions.Keyring,
							SkipUpdate:       false,
							Getters:          getter.All(cli.New()),
							RepositoryConfig: r.config.EnvSettings.RepositoryConfig,
							RepositoryCache:  r.config.EnvSettings.RepositoryCache,
						}
						if err := man.Update(); err != nil {
							return nil, err
						}
					} else {
						return nil, err
					}
				}
			}

			_, err = install.Run(chart, values.Values())
			return nil, err
		}
	}

	release, err := upgrade.Run(r.name, chart, values.Values())
	if err != nil {
		return nil, err
	}
	return getRelease(r.config, release)
}
