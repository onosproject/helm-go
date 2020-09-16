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

package values

import (
	"encoding/csv"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

// NewImmutable creates a new immutable Values object
func NewImmutable(values map[string]interface{}) *ImmutableValues {
	return &ImmutableValues{
		values: values,
	}
}

// New creates a new Values object
func New(values ...map[string]interface{}) *Values {
	if len(values) > 1 {
		panic("too many values")
	}
	if len(values) == 1 {
		return &Values{
			values: values[0],
		}
	}
	return &Values{
		values: make(map[string]interface{}),
	}
}

// ImmutableValues is a helper for reading values
type ImmutableValues struct {
	values map[string]interface{}
}

func (v *ImmutableValues) Get(path string) interface{} {
	keys := splitKeys(path)
	parentKeys, childKey := keys[:len(keys)-1], keys[len(keys)-1]
	parent := getParent(v.values, parentKeys)
	return copyValue(parent[childKey])
}

func (v *ImmutableValues) Values() map[string]interface{} {
	return copy(v.values)
}

// Values is a utility for managing values
type Values struct {
	values map[string]interface{}
}

func (v *Values) Values() map[string]interface{} {
	return copy(v.values)
}

func (v *Values) Set(path string, value interface{}) *Values {
	keys := splitKeys(path)
	parentKeys, childKey := keys[:len(keys)-1], keys[len(keys)-1]
	parent := getParent(v.values, parentKeys)
	parent[childKey] = value
	return v
}

func (v *Values) Get(path string) interface{} {
	keys := splitKeys(path)
	parentKeys, childKey := keys[:len(keys)-1], keys[len(keys)-1]
	parent := getParent(v.values, parentKeys)
	return parent[childKey]
}

func (v *Values) Normalize() *Values {
	v.values = normalize(v.values)
	return v
}

func (v *Values) Override(overrides *Values) *Values {
	v.values = override(v.values, overrides.values)
	return v
}

func (v *Values) Immutable() *ImmutableValues {
	return NewImmutable(copy(v.values))
}

func splitKeys(path string) []string {
	r := csv.NewReader(strings.NewReader(path))
	r.Comma = '.'
	names, err := r.Read()
	if err != nil {
		panic(err)
	}
	return names
}

func getParent(parent map[string]interface{}, path []string) map[string]interface{} {
	if len(path) == 0 {
		return parent
	}
	child, ok := parent[path[0]]
	if !ok {
		child = make(map[string]interface{})
		parent[path[0]] = child
	}
	return getParent(child.(map[string]interface{}), path[1:])
}

// override recursively merges values 'b' into values 'a', returning a new merged values map
func override(values, overrides map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(overrides))
	for k, v := range overrides {
		out[k] = v
	}
	for k, v := range values {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = override(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// normalize normalizes the given values map, converting structs into maps
func normalize(values map[string]interface{}) map[string]interface{} {
	return copy(values)
}

// copy copies the given values map
func copy(values map[string]interface{}) map[string]interface{} {
	return copyValue(values).(map[string]interface{})
}

// copyValue copies the given value
func copyValue(value interface{}) interface{} {
	kind := reflect.ValueOf(value).Kind()
	if kind == reflect.Struct {
		return copyStruct(value.(struct{}))
	} else if kind == reflect.Map {
		return copyMap(value.(map[string]interface{}))
	} else if kind == reflect.Slice {
		return copySlice(value.([]interface{}))
	}
	return value
}

// copyStruct copies the given struct
func copyStruct(value struct{}) map[string]interface{} {
	elem := reflect.ValueOf(value).Elem()
	elemType := elem.Type()
	normalized := make(map[string]interface{})
	for i := 0; i < elem.NumField(); i++ {
		key := getFieldKey(elemType.Field(i))
		value := copyValue(elem.Field(i).Interface())
		normalized[key] = value
	}
	return normalized
}

// copyMap copies the given map
func copyMap(values map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	for key, value := range values {
		normalized[key] = copyValue(value)
	}
	return normalized
}

// copySlice copies the given slice
func copySlice(values []interface{}) interface{} {
	normalized := make([]interface{}, len(values))
	for i, value := range values {
		normalized[i] = copyValue(value)
	}
	return normalized
}

// getFieldKey returns the map key name for the given struct field
func getFieldKey(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag != "" {
		return strings.Split(tag, ",")[0]
	}
	return strcase.ToLowerCamel(field.Name)
}
