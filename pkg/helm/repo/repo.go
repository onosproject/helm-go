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

package repo

import (
	"context"
	"fmt"
	"github.com/gofrs/flock"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var settings = cli.New()

// NewClient creates a new Helm repository client
func NewClient() Client {
	return &repoClient{}
}

// Client is a Helm repository client
type Client interface {
	// Add adds a repository
	Add(name string) *AddRequest
	// Remove removes a repository
	Remove(name string) *RemoveRequest
}

// repoClient is the Helm repository client
type repoClient struct{}

// Add adds a repository
func (c *repoClient) Add(name string) *AddRequest {
	return &AddRequest{
		repo: &Repository{
			Name:      name,
			repoFile:  settings.RepositoryConfig,
			cacheFile: settings.RepositoryCache,
		},
	}
}

// Remove removes a repository
func (c *repoClient) Remove(name string) *RemoveRequest {
	return &RemoveRequest{
		repo: &Repository{
			Name:      name,
			repoFile:  settings.RepositoryConfig,
			cacheFile: settings.RepositoryCache,
		},
	}
}

// Repository is a Helm chart repository
type Repository struct {
	Name      string
	repoFile  string
	cacheFile string
}

// AddRequest is a Helm chart repository add request
type AddRequest struct {
	repo     *Repository
	url      string
	caFile   string
	keyFile  string
	certFile string
	username string
	password string
}

func (r *AddRequest) URL(url string) *AddRequest {
	r.url = url
	return r
}

func (r *AddRequest) CaFile(caFile string) *AddRequest {
	r.caFile = caFile
	return r
}

func (r *AddRequest) KeyFile(keyFile string) *AddRequest {
	r.keyFile = keyFile
	return r
}

func (r *AddRequest) CertFile(certFile string) *AddRequest {
	r.certFile = certFile
	return r
}

func (r *AddRequest) Username(username string) *AddRequest {
	r.username = username
	return r
}

func (r *AddRequest) Password(password string) *AddRequest {
	r.password = password
	return r
}

func (r *AddRequest) Do() (*Repository, error) {
	err := os.MkdirAll(filepath.Dir(r.repo.repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(r.repo.repoFile, filepath.Ext(r.repo.repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer func() {
			_ = fileLock.Unlock()
		}()
	}
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(r.repo.repoFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}

	if f.Has(r.repo.Name) {
		return nil, fmt.Errorf("repository %q already exists", r.repo.Name)
	}

	e := repo.Entry{
		Name:     r.repo.Name,
		URL:      r.url,
		Username: r.username,
		Password: r.password,
		CertFile: r.certFile,
		KeyFile:  r.keyFile,
		CAFile:   r.caFile,
	}

	cr, err := repo.NewChartRepository(&e, getter.All(settings))
	if err != nil {
		return nil, err
	}

	if _, err := cr.DownloadIndexFile(); err != nil {
		return nil, err
	}

	f.Update(&e)

	if err := f.WriteFile(r.repo.repoFile, 0644); err != nil {
		return nil, err
	}
	return r.repo, nil
}

// RemoveRequest is a Helm chart repository remove request
type RemoveRequest struct {
	repo *Repository
}

func (r *RemoveRequest) Do() error {
	cr, err := repo.LoadFile(r.repo.repoFile)
	if err != nil {
		return err
	}

	if !cr.Remove(r.repo.Name) {
		return fmt.Errorf("no repo named %q found", r.repo.Name)
	}
	if err := cr.WriteFile(r.repo.repoFile, 0644); err != nil {
		return err
	}

	idx := filepath.Join(r.repo.cacheFile, helmpath.CacheChartsFile(r.repo.Name))
	if _, err := os.Stat(idx); err == nil {
		os.Remove(idx)
	}

	idx = filepath.Join(r.repo.cacheFile, helmpath.CacheIndexFile(r.repo.Name))
	if _, err := os.Stat(idx); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(idx)
}
