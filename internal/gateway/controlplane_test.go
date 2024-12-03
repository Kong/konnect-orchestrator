package gateway

import (
	"context"
	"fmt"
	"testing"

	"github.com/Kong/konnect-orchestrator/internal"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApplyControlPlane(t *testing.T) {
	tests := []struct {
		name       string
		envName    string
		env        internal.EnvironmentManifest
		teamName   string
		setup      func(*MockControlPlaneService)
		wantErr    bool
		expectedID string
	}{
		{
			name:    "creates new control plane with labels",
			envName: "DEV",
			env: internal.EnvironmentManifest{
				Type:   "DEV",
				Region: "us",
			},
			teamName: "team1",
			setup: func(m *MockControlPlaneService) {
				m.On("ListControlPlanes",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&operations.ListControlPlanesResponse{
					ListControlPlanesResponse: &components.ListControlPlanesResponse{
						Data: []components.ControlPlane{},
					},
				}, nil)

				m.On("CreateControlPlane",
					mock.Anything,
					mock.MatchedBy(func(req components.CreateControlPlaneRequest) bool {
						return req.Name == "team1-DEV" &&
							*req.Description == "Control plane for team team1 in environment DEV" &&
							*req.ClusterType == components.CreateControlPlaneRequestClusterType("CLUSTER_TYPE_CONTROL_PLANE") &&
							req.Labels["env"] == "DEV" &&
							req.Labels["team"] == "team1"
					}),
					mock.Anything,
				).Return(
					&operations.CreateControlPlaneResponse{
						ControlPlane: &components.ControlPlane{
							ID: "new-cp-123",
						},
					}, nil)
			},
			wantErr:    false,
			expectedID: "new-cp-123",
		},
		{
			name:    "updates existing control plane with new labels",
			envName: "PROD",
			env: internal.EnvironmentManifest{
				Type:   "PROD",
				Region: "us",
			},
			teamName: "team1",
			setup: func(m *MockControlPlaneService) {
				m.On("ListControlPlanes",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&operations.ListControlPlanesResponse{
					ListControlPlanesResponse: &components.ListControlPlanesResponse{
						Data: []components.ControlPlane{
							{
								ID:          "existing-cp-123",
								Name:        "team1-PROD",
								Labels:      map[string]string{},
								Description: kk.String("Old description"),
							},
						},
					},
				}, nil)

				m.On("UpdateControlPlane",
					mock.Anything,
					"existing-cp-123",
					mock.MatchedBy(func(req components.UpdateControlPlaneRequest) bool {
						return *req.Description == "Control plane for team team1 in environment PROD" &&
							req.Labels["env"] == "PROD" &&
							req.Labels["team"] == "team1"
					}),
					mock.Anything,
				).Return(&operations.UpdateControlPlaneResponse{}, nil)
			},
			wantErr:    false,
			expectedID: "existing-cp-123",
		},
		{
			name:    "handles list error",
			envName: "DEV",
			env: internal.EnvironmentManifest{
				Type:   "DEV",
				Region: "us",
			},
			teamName: "team1",
			setup: func(m *MockControlPlaneService) {
				m.On("ListControlPlanes", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("list error"))
			},
			wantErr:    true,
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCPSvc := &MockControlPlaneService{}
			tt.setup(mockCPSvc)

			id, err := ApplyControlPlane(context.Background(), mockCPSvc, tt.envName, tt.env, tt.teamName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
			mock.AssertExpectationsForObjects(t, mockCPSvc)
		})
	}
}
