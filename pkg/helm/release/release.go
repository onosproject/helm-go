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
	"errors"
	"github.com/onosproject/helm-go/pkg/helm/context"
	"github.com/onosproject/helm-go/pkg/helm/values"
	"github.com/onosproject/helm-go/pkg/kubernetes"
	"github.com/onosproject/helm-go/pkg/kubernetes/filter"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"log"
	"os"
	"sync"
	"time"
)

var settings = cli.New()

// NewClient returns a new release client
func NewClient(namespace string) Client {
	config, err := conf.get(namespace)
	if err != nil {
		panic(err)
	}
	return &releaseClient{
		namespace: namespace,
		config:    config,
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
	namespace string
	config    *action.Configuration
}

func (c *releaseClient) Namespace() string {
	return c.namespace
}

// Get gets a release
func (c *releaseClient) Get(name string) (*Release, error) {
	list, err := c.config.Releases.List(func(r *release.Release) bool {
		return r.Namespace == c.namespace && r.Name == name
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
		return r.Namespace == c.namespace
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

// UninstallRequest is a release uninstall request
type UninstallRequest struct {
	client Client
	config *action.Configuration
	name   string
}

func (r *UninstallRequest) Do() error {
	uninstall := action.NewUninstall(r.config)
	_, err := uninstall.Run(r.name)
	return err
}

// UpgradeRequest is a release upgrade request
type UpgradeRequest struct {
	client       Client
	config       *action.Configuration
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
	upgrade := action.NewUpgrade(r.config)

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
	path, err := upgrade.ChartPathOptions.LocateChart(r.chart, settings)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chart, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	ctx := context.GetContext().Release(r.name)
	values := r.values.Normalize().Override(values.New(ctx.Values.Values()))

	if upgrade.Install {
		// If a release does not exist, install it. If another error occurs during
		// the check, ignore the error and continue with the upgrade.
		histClient := action.NewHistory(r.config)
		histClient.Max = 1
		if _, err := histClient.Run(r.name); err == driver.ErrReleaseNotFound {
			install := action.NewInstall(r.config)
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

// RollbackRequest is a release rollback request
type RollbackRequest struct {
	client Client
	config *action.Configuration
	name   string
}

func (r *RollbackRequest) Do() error {
	rollback := action.NewRollback(r.config)
	return rollback.Run(r.name)
}

var conf = &configs{
	configs: make(map[string]*action.Configuration),
}

type configs struct {
	configs map[string]*action.Configuration
	mu      sync.Mutex
}

func (c *configs) get(namespace string) (*action.Configuration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	config, ok := c.configs[namespace]
	if !ok {
		config = &action.Configuration{}
		if err := config.Init(settings.RESTClientGetter(), namespace, "memory", log.Printf); err != nil {
			return nil, err
		}
		c.configs[namespace] = config
	}
	return config, nil
}

func getRelease(config *action.Configuration, release *release.Release) (*Release, error) {
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
