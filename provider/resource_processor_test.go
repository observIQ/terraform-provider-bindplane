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

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bpclient "github.com/observiq/bindplane-op-enterprise/client"
	"github.com/observiq/bindplane-op-enterprise/client/mocks"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// testProcessor returns a batch processor as it would be returned
// by the Bindplane API.
func testProcessor(t *testing.T, id, recommendation string) *model.Processor {
	r, err := resource.AnyResourceV1(id, "my-batch", "batch:1", model.KindProcessor, nil, nil, resource.WithRecommendation(recommendation))
	require.NoError(t, err)

	p, err := model.AsKind[*model.Processor](&r)
	require.NoError(t, err)
	return p
}

func TestResourceProcessorCreate(t *testing.T) {
	cases := []struct {
		name           string
		recommendation string
	}{
		{
			name: "without recommendation",
		},
		{
			name:           "with recommendation",
			recommendation: "google_secops_batch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceProcessor().Schema, map[string]any{
				"name":           "my-batch",
				"type":           "batch",
				"rollout":        false,
				"recommendation": tc.recommendation,
			})

			m := mocks.NewMockBindplane(t)

			// The processor does not exist before it is applied.
			m.EXPECT().Processor(mock.Anything, "my-batch").Return(nil, nil).Once()

			var applied *model.AnyResource
			m.EXPECT().Apply(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, r []*model.AnyResource, _ ...bpclient.ApplyOption) ([]*model.AnyResourceStatus, error) {
					require.Len(t, r, 1)
					applied = r[0]
					return []*model.AnyResourceStatus{{Resource: *r[0], Status: model.StatusCreated}}, nil
				},
			).Once()

			// The read after apply returns the applied processor.
			m.EXPECT().Processor(mock.Anything, "my-batch").RunAndReturn(
				func(_ context.Context, _ string) (*model.Processor, error) {
					return model.AsKind[*model.Processor](applied)
				},
			).Once()

			require.NoError(t, resourceProcessorCreate(d, &client.BindPlane{Client: m}))

			rec, ok := applied.Spec["recommendation"]
			if tc.recommendation == "" {
				require.False(t, ok, "recommendation should not be sent when unset")
			} else {
				require.True(t, ok, "recommendation should be sent when set")
				require.Equal(t, tc.recommendation, rec)
			}

			require.Equal(t, applied.ID(), d.Id())
			require.Equal(t, tc.recommendation, d.Get("recommendation"))
		})
	}
}

func TestResourceProcessorRead(t *testing.T) {
	const id = "01JTEST0000000000000000000"

	cases := []struct {
		name                string
		stateRecommendation string
		processor           *model.Processor
		expectID            string
		expectRecommend     string
	}{
		{
			name:            "recommendation set by bindplane",
			processor:       testProcessor(t, id, "google_secops_batch"),
			expectID:        id,
			expectRecommend: "google_secops_batch",
		},
		{
			name:                "recommendation removed outside terraform",
			stateRecommendation: "google_secops_batch",
			processor:           testProcessor(t, id, ""),
			expectID:            id,
			expectRecommend:     "",
		},
		{
			name:                "recommendation changed outside terraform",
			stateRecommendation: "google_secops_batch",
			processor:           testProcessor(t, id, "azure_blob_batch"),
			expectID:            id,
			expectRecommend:     "azure_blob_batch",
		},
		{
			name:                "processor does not exist",
			stateRecommendation: "google_secops_batch",
			processor:           nil,
			expectID:            "",
			expectRecommend:     "google_secops_batch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceProcessor().Schema, map[string]any{
				"name":           "my-batch",
				"type":           "batch",
				"rollout":        false,
				"recommendation": tc.stateRecommendation,
			})
			d.SetId(id)

			m := mocks.NewMockBindplane(t)
			m.EXPECT().Processor(mock.Anything, "my-batch").Return(tc.processor, nil).Once()

			require.NoError(t, resourceProcessorRead(d, &client.BindPlane{Client: m}))
			require.Equal(t, tc.expectID, d.Id())
			require.Equal(t, tc.expectRecommend, d.Get("recommendation"))
		})
	}
}
