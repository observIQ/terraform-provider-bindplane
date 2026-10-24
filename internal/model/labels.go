// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Label keys and values follow the Kubernetes label syntax, which Bindplane uses.
// See https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
const (
	maxLabelNameLength   = 63
	maxLabelPrefixLength = 253
	maxLabelValueLength  = 63
)

var (
	labelNameRegexp   = regexp.MustCompile(`^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$`)
	labelPrefixRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	labelValueRegexp  = regexp.MustCompile(`^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`)
)

// Labels are the labels of a resource. They are represented in JSON as an object.
type Labels struct {
	Set map[string]string
}

// LabelsFromMap returns Labels with the valid labels in labels, and an error
// describing any invalid labels.
func LabelsFromMap(labels map[string]string) (Labels, error) {
	var errs error
	valid := map[string]string{}
	for name, value := range labels {
		if err := validateLabelKey(name); err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s is not a valid label name: %w", name, err))
			continue
		}
		if err := validateLabelValue(value); err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s is not a valid label value: %w", value, err))
			continue
		}
		valid[name] = value
	}
	return Labels{Set: valid}, errs
}

// LabelsFromValidatedMap returns Labels without validating labels
func LabelsFromValidatedMap(labels map[string]string) Labels {
	return Labels{Set: labels}
}

// AsMap returns the labels as a map
func (l Labels) AsMap() map[string]string {
	return l.Set
}

// MarshalJSON marshals the labels as a JSON object. Nil labels are
// marshaled as an empty object.
func (l Labels) MarshalJSON() ([]byte, error) {
	if l.Set == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(l.Set)
}

// UnmarshalJSON unmarshals labels from a JSON object
func (l *Labels) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &l.Set)
}

// validateLabelKey validates a label key, which is a name with an
// optional DNS subdomain prefix, e.g. bindplane.com/name
func validateLabelKey(key string) error {
	name := key
	parts := strings.Split(key, "/")
	switch len(parts) {
	case 1:
	case 2:
		var prefix string
		prefix, name = parts[0], parts[1]
		switch {
		case prefix == "":
			return errors.New("prefix part must be non-empty")
		case len(prefix) > maxLabelPrefixLength:
			return fmt.Errorf("prefix part must be no more than %d characters", maxLabelPrefixLength)
		case !labelPrefixRegexp.MatchString(prefix):
			return errors.New("prefix part must be a lowercase DNS subdomain")
		}
	default:
		return errors.New("must be a name with an optional DNS subdomain prefix and '/' (e.g. 'example.com/MyName')")
	}

	switch {
	case name == "":
		return errors.New("name part must be non-empty")
	case len(name) > maxLabelNameLength:
		return fmt.Errorf("name part must be no more than %d characters", maxLabelNameLength)
	case !labelNameRegexp.MatchString(name):
		return errors.New("name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character")
	}
	return nil
}

// validateLabelValue validates a label value, which may be empty
func validateLabelValue(value string) error {
	switch {
	case len(value) > maxLabelValueLength:
		return fmt.Errorf("must be no more than %d characters", maxLabelValueLength)
	case !labelValueRegexp.MatchString(value):
		return errors.New("must be empty or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character")
	}
	return nil
}
