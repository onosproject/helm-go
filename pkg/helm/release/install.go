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
	"github.com/onosproject/helm-go/pkg/helm/context"
	"github.com/onosproject/helm-go/pkg/helm/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"os"
	"time"
)

// InstallRequest is a release install request
type InstallRequest struct {
	client                   Client
	config                   *action.Configuration
	name                     string
	chart                    string
	repo                     string
	caFile                   string
	keyFile                  string
	certFile                 string
	username                 string
	password                 string
	version                  string
	values                   *values.Values
	skipCRDs                 bool
	includeCRDs              bool
	disableHooks             bool
	disableOpenAPIValidation bool
	dryRun                   bool
	replace                  bool
	atomic                   bool
	wait                     bool
	timeout                  time.Duration
}

func (r *InstallRequest) CaFile(caFile string) *InstallRequest {
	r.caFile = caFile
	return r
}

func (r *InstallRequest) KeyFile(keyFile string) *InstallRequest {
	r.keyFile = keyFile
	return r
}

func (r *InstallRequest) CertFile(certFile string) *InstallRequest {
	r.certFile = certFile
	return r
}

func (r *InstallRequest) Username(username string) *InstallRequest {
	r.username = username
	return r
}

func (r *InstallRequest) Password(password string) *InstallRequest {
	r.password = password
	return r
}

func (r *InstallRequest) Repo(url string) *InstallRequest {
	r.repo = url
	return r
}

func (r *InstallRequest) Version(version string) *InstallRequest {
	r.version = version
	return r
}

func (r *InstallRequest) Set(path string, value interface{}) *InstallRequest {
	r.values.Set(path, value)
	return r
}

func (r *InstallRequest) SkipCRDs() *InstallRequest {
	r.skipCRDs = true
	return r
}

func (r *InstallRequest) IncludeCRDs() *InstallRequest {
	r.includeCRDs = true
	return r
}

func (r *InstallRequest) DisableHooks() *InstallRequest {
	r.disableHooks = true
	return r
}

func (r *InstallRequest) DisableOpenAPIValidation() *InstallRequest {
	r.disableOpenAPIValidation = true
	return r
}

func (r *InstallRequest) DryRun() *InstallRequest {
	r.dryRun = true
	return r
}

func (r *InstallRequest) Replace() *InstallRequest {
	r.replace = true
	return r
}

func (r *InstallRequest) Atomic() *InstallRequest {
	r.atomic = true
	return r
}

func (r *InstallRequest) Wait() *InstallRequest {
	r.wait = true
	return r
}

func (r *InstallRequest) Timeout(timeout time.Duration) *InstallRequest {
	r.timeout = timeout
	return r
}

func (r *InstallRequest) Do() (*Release, error) {
	install := action.NewInstall(r.config)

	// Setup the repo options
	install.RepoURL = r.repo
	install.Username = r.username
	install.Password = r.password
	install.CaFile = r.caFile
	install.KeyFile = r.keyFile
	install.CertFile = r.certFile

	// Setup the chart options
	install.Version = r.version

	// Setup the release options
	install.ReleaseName = r.name
	install.Namespace = r.client.Namespace()
	install.Atomic = r.atomic
	install.Replace = r.replace
	install.DryRun = r.dryRun
	install.DisableHooks = r.disableHooks
	install.DisableOpenAPIValidation = r.disableOpenAPIValidation
	install.SkipCRDs = r.skipCRDs
	install.IncludeCRDs = r.includeCRDs
	install.Wait = r.wait
	install.Timeout = r.timeout

	// Locate the chart path
	path, err := install.ChartPathOptions.LocateChart(r.chart, settings)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chart, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

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
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	ctx := context.GetContext().Release(r.name)
	values := r.values.Normalize().Override(values.New(ctx.Values.Values()))
	release, err := install.Run(chart, values.Values())
	if err != nil {
		return nil, err
	}
	return getRelease(r.config, release)
}
