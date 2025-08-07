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

package parameter

import (
	"fmt"
)

// Condition is the json-encoded value of the Condition parameter.
type Condition struct {
	// OTTL is the OTTL condition string either specified raw or generated from the UI.
	OTTL string `mapstructure:"ottl"`

	// OTTLContext is the OTTL context for the condition statement.
	OTTLContext string `mapstructure:"ottlContext"`

	// UI is the UI representation of the condition and is used to repopulate the condition
	// UI.
	UI OTTLConditionStatement `mapstructure:"ui"`
}

// OTTLConditionStatement represents either a single OTTL statement or a group of
// statements joined with a boolean operator.
type OTTLConditionStatement struct {
	// Operator is the operator in this statement. If there are multiple statements and the
	// operator is empty, it is assumed to be "and".
	Operator string `mapstructure:"operator"`

	// Match is the type of the field in this statement
	Match string `mapstructure:"match"`

	// Key is the key of the field in this statement
	Key string `mapstructure:"key"`

	// Value is the value to compare against using the operator in this statement
	Value string `mapstructure:"value"`

	// Statements contains sub-statements if the Operator is "and" or "or".
	Statements []OTTLConditionStatement `mapstructure:"statements"`
}

// ValidateOTTLConditionStatement validates that a condition UI block is properly formed and rejects malformed conditions
func ValidateOTTLConditionStatement(ui OTTLConditionStatement) error {
	if ui.Operator == "OR" || ui.Operator == "AND" {
		if len(ui.Statements) < 2 {
			return fmt.Errorf("parent operator '%s' must have at least 2 child statements, found %d", ui.Operator, len(ui.Statements))
		}
		// Validate each child statement
		for i, statement := range ui.Statements {
			if err := ValidateOTTLConditionStatement(statement); err != nil {
				return fmt.Errorf("child statement %d: %w", i, err)
			}
		}
		// If operator exists, it must have a key and match
	} else if ui.Operator != "" {
		if ui.Key == "" {
			return fmt.Errorf("statement with operator '%s' must have a key", ui.Operator)
		}
		if ui.Match == "" {
			return fmt.Errorf("statement with operator '%s' must have a match", ui.Operator)
		}
	}

	return nil
}
