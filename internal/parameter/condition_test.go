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
	"testing"
)

func TestValidateConditionStatement(t *testing.T) {
	tests := []struct {
		name    string
		ui      OTTLConditionStatement
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid single statement",
			ui: OTTLConditionStatement{
				Operator: "and",
				Match:    "body",
				Key:      "severity",
				Value:    "ERROR",
			},
			wantErr: false,
		},
		{
			name: "valid OR with two statements",
			ui: OTTLConditionStatement{
				Operator: "OR",
				Statements: []OTTLConditionStatement{
					{
						Operator: "or",
						Match:    "body",
						Key:      "severity",
						Value:    "ERROR",
					},
					{
						Operator: "or",
						Match:    "body",
						Key:      "severity",
						Value:    "INFO",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "malformed OR with only one statement - should fail",
			ui: OTTLConditionStatement{
				Operator: "OR",
				Statements: []OTTLConditionStatement{
					{
						Operator: "or",
						Match:    "body",
						Key:      "severity",
						Value:    "ERROR",
					},
				},
			},
			wantErr: true,
			errMsg:  "parent operator 'OR' must have at least 2 child statements, found 1",
		},
		{
			name: "nested malformed OR - should fail",
			ui: OTTLConditionStatement{
				Operator: "OR",
				Statements: []OTTLConditionStatement{
					{
						Operator: "Equals",
						Key:      "service",
						Match:    "resource",
						Value:    "example.link",
					},
					{
						Operator: "OR",
						Statements: []OTTLConditionStatement{
							{
								Operator: "Equals",
								Key:      "telemetry.type",
								Match:    "resource",
								Value:    "metric",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "child statement 1: parent operator 'OR' must have at least 2 child statements, found 1",
		},
		{
			name: "statement with operator but no key - should fail",
			ui: OTTLConditionStatement{
				Operator: "Equals",
				Match:    "resource",
				Value:    "prod",
			},
			wantErr: true,
			errMsg:  "statement with operator 'Equals' must have a key",
		},
		{
			name: "statement with operator but no match - should fail",
			ui: OTTLConditionStatement{
				Operator: "Equals",
				Key:      "env",
				Value:    "prod",
			},
			wantErr: true,
			errMsg:  "statement with operator 'Equals' must have a match",
		},

		{
			name:    "completely empty - should pass",
			ui:      OTTLConditionStatement{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOTTLConditionStatement(tt.ui)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateOTTLConditionStatement() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ValidateOTTLConditionStatement() error = %v, want error message %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateOTTLConditionStatement() unexpected error = %v", err)
				}
			}
		})
	}
}
