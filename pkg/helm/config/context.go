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

package config

import (
	"github.com/onosproject/helm-go/pkg/helm/values"
)

var context = &Context{}

// SetContext sets the Helm context
func SetContext(ctx *Context) {
	context = ctx
}

// GetContext gets the Helm context
func GetContext() *Context {
	if context == nil {
		context = &Context{
			Values: map[string]*values.ImmutableValues{},
		}
	}
	return context
}

// New creates a new Helm context
func New(values map[string]*values.ImmutableValues) *Context {
	return &Context{
		Values: values,
	}
}

// Context is a Helm context
type Context struct {
	// Values is a mapping of release values
	Values map[string]*values.ImmutableValues
}

// Release returns the context for the given release
func (c *Context) Release(name string) *ReleaseContext {
	v, ok := c.Values[name]
	if !ok {
		v = values.New().Immutable()
	}
	return &ReleaseContext{
		Values: v,
	}
}

// ReleaseContext is a Helm release context
type ReleaseContext struct {
	// Values is the release values
	Values *values.ImmutableValues
}
