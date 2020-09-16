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

package codegen

import "path"

// ResourceReferenceOptions contains options for generating a resource reference
type ResourceReferenceOptions struct {
	Location Location
	Package  Package
	Types    ResourceReaderTypes
}

// ResourceReferenceTypes contains types for generating a resource reference
type ResourceReferenceTypes struct {
	Interface string
	Struct    string
}

func generateResourceReference(options ResourceOptions) error {
	if options.Reference == nil {
		return nil
	}
	return generateTemplate(getTemplate("resourcereference.tpl"), path.Join(options.Reference.Location.Path, options.Reference.Location.File), options)
}
