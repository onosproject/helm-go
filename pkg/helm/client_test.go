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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalInstall(t *testing.T) {
	client := New("default")
	release, err := client.Releases().
		Install("atomix-controller", "../atomix-helm-charts/atomix-controller").
		Set("scope", "Namespace").
		Wait().
		Do()
	assert.NoError(t, err)
	assert.NotNil(t, release)

	err = client.Releases().
		Uninstall("atomix-controller").
		Do()
	assert.NoError(t, err)
}

func TestRemoteInstall(t *testing.T) {
	client := New("default")
	release, err := client.Releases().
		Install("atomix-controller", "atomix-controller").
		Repo("https://charts.atomix.io").
		Set("scope", "Namespace").
		Wait().
		Do()
	assert.NoError(t, err)
	assert.NotNil(t, release)

	err = client.Releases().
		Uninstall("atomix-controller").
		Do()
	assert.NoError(t, err)
}

func TestRemoteInstallFromRepo(t *testing.T) {
	client := New("default")
	repo, err := client.Repos().
		Add("atomix-test").
		URL("https://charts.atomix.io").
		Do()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	release, err := client.Releases().
		Install("atomix-controller", "atomix-test/atomix-controller").
		Set("scope", "Namespace").
		Wait().
		Do()
	assert.NoError(t, err)
	assert.NotNil(t, release)

	err = client.Releases().
		Uninstall("atomix-controller").
		Do()
	assert.NoError(t, err)

	err = client.Repos().
		Remove("atomix-test").
		Do()
	assert.NoError(t, err)
}
