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
	"helm.sh/helm/v3/pkg/action"
)

// UninstallRequest is a release uninstall request
type UninstallRequest struct {
	client Client
	config *config.Config
	name   string
}

func (r *UninstallRequest) Do() error {
	uninstall := action.NewUninstall(r.config.Configuration)
	_, err := uninstall.Run(r.name)
	return err
}
