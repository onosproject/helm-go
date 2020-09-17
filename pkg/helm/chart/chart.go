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

package chart

import (
	"errors"
	"fmt"
	"github.com/onosproject/helm-go/pkg/helm/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

var settings = cli.New()

const chartsDir = "charts"

// NewClient creates a new Helm chart client
func NewClient() Client {
	return &chartClient{}
}

// Client is a Helm chart client
type Client interface {
	// Get gets a chart by name
	Get(name string) (*Chart, error)
}

// chartClient is a Helm chart client
type chartClient struct{}

// Get gets a chart
func (c *chartClient) Get(name string) (*Chart, error) {
	opts := action.ChartPathOptions{}
	path, err := opts.LocateChart(name, settings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return newChart(chart)
}

var _ Client = &chartClient{}

func newChart(chart *chart.Chart) (*Chart, error) {
	vals := values.New()
	for _, f := range chart.Raw {
		if f.Name == chartutil.ValuesfileName {
			val := make(map[string]interface{})
			if err := yaml.Unmarshal(f.Data, val); err != nil {
				return nil, err
			}
			vals.Override(values.New(val))
		}
	}

	return &Chart{
		chart:  chart,
		Name:   chart.Name(),
		values: vals.Immutable(),
	}, nil
}

// Chart is a Helm chart
type Chart struct {
	chart  *chart.Chart
	Name   string
	values *values.ImmutableValues
}

// Values returns the chart values
func (c *Chart) Values() *values.ImmutableValues {
	return c.values
}

// SubChart returns a sub-chart by name
func (c *Chart) SubChart(name string) (*Chart, error) {
	deps := c.chart.Metadata.Dependencies
	if deps == nil {
		return nil, errors.New("chart not found")
	}

	var subDep *chart.Dependency
	for _, dep := range deps {
		if dep.Name == name {
			subDep = dep
			break
		}
	}

	if subDep == nil {
		return nil, errors.New("chart not found")
	}

	// If CheckDependencies returns an error, we have unfulfilled dependencies.
	// As of Helm 2.4.0, this is treated as a stopping condition:
	// https://github.com/helm/helm/issues/2209
	if err := action.CheckDependencies(c.chart, deps); err != nil {
		man := &downloader.Manager{
			Out:              os.Stdout,
			ChartPath:        c.chart.ChartPath(),
			SkipUpdate:       false,
			Getters:          getter.All(cli.New()),
			RepositoryConfig: settings.RepositoryConfig,
			RepositoryCache:  settings.RepositoryCache,
		}
		if err := man.Update(); err != nil {
			return nil, err
		}
	}

	subPath := filepath.Join(c.chart.ChartPath(), chartsDir, fmt.Sprintf("%s-%s.tgz", subDep.Name, subDep.Version))
	subChart, err := loader.Load(subPath)
	if err != nil {
		return nil, err
	}
	return newChart(subChart)
}

// SubCharts returns the chart's sub-charts
func (c *Chart) SubCharts() ([]*Chart, error) {
	deps := c.chart.Metadata.Dependencies
	if deps == nil {
		return nil, errors.New("chart not found")
	}

	// If CheckDependencies returns an error, we have unfulfilled dependencies.
	// As of Helm 2.4.0, this is treated as a stopping condition:
	// https://github.com/helm/helm/issues/2209
	if err := action.CheckDependencies(c.chart, deps); err != nil {
		man := &downloader.Manager{
			Out:              os.Stdout,
			ChartPath:        c.chart.ChartPath(),
			SkipUpdate:       false,
			Getters:          getter.All(cli.New()),
			RepositoryConfig: settings.RepositoryConfig,
			RepositoryCache:  settings.RepositoryCache,
		}
		if err := man.Update(); err != nil {
			return nil, err
		}
	}

	charts := make([]*Chart, 0, len(deps))
	for _, dep := range deps {
		subPath := filepath.Join(c.chart.ChartPath(), chartsDir, fmt.Sprintf("%s-%s.tgz", dep.Name, dep.Version))
		subChart, err := loader.Load(subPath)
		if err != nil {
			return nil, err
		}
		chart, err := newChart(subChart)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}
	return charts, nil
}
