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
			name: "valid or with two statements",
			ui: OTTLConditionStatement{
				Operator: "or",
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
			name: "malformed or with only one statement - should fail",
			ui: OTTLConditionStatement{
				Operator: "or",
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
			errMsg:  "parent operator 'or' must not have only one child statement, found 1",
		},
		{
			name: "nested malformed or - should fail",
			ui: OTTLConditionStatement{
				Operator: "or",
				Statements: []OTTLConditionStatement{
					{
						Operator: "Equals",
						Key:      "service",
						Match:    "resource",
						Value:    "example.link",
					},
					{
						Operator: "or",
						Statements: []OTTLConditionStatement{
							{
								Operator: "and",
								Key:      "telemetry.type",
								Match:    "resource",
								Value:    "metric",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "child statement 1: parent operator 'or' must not have only one child statement, found 1",
		},
		{
			name: "statement with operator but no key - should fail",
			ui: OTTLConditionStatement{
				Operator: "or",
				Match:    "resource",
				Value:    "prod",
			},
			wantErr: true,
			errMsg:  "statement with operator 'or' must have a key",
		},
		{
			name: "statement with operator but no match - should fail",
			ui: OTTLConditionStatement{
				Operator: "and",
				Key:      "env",
				Value:    "prod",
			},
			wantErr: true,
			errMsg:  "statement with operator 'and' must have a match",
		},
		{
			name: "nested statement with operator but no match or - should fail",
			ui: OTTLConditionStatement{
				Operator: "or",
				Statements: []OTTLConditionStatement{
					{
						Operator: "Equals",
						Key:      "service",
						Match:    "resource",
						Value:    "example.link",
					},
					{
						Operator: "or",
						Statements: []OTTLConditionStatement{
							{
								Operator: "and",
								Key:      "telemetry.type",
								Value:    "metric",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "child statement 1: parent operator 'or' must not have only one child statement, found 1",
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
