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

type MockControlPlaneService struct {
	mock.Mock
}

func (m *MockControlPlaneService) ListControlPlanes(ctx context.Context, request operations.ListControlPlanesRequest, opts ...operations.Option) (*operations.ListControlPlanesResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*operations.ListControlPlanesResponse), args.Error(1)
}

func (m *MockControlPlaneService) CreateControlPlane(ctx context.Context, request components.CreateControlPlaneRequest, opts ...operations.Option) (*operations.CreateControlPlaneResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*operations.CreateControlPlaneResponse), args.Error(1)
}

func (m *MockControlPlaneService) UpdateControlPlane(ctx context.Context, id string, request components.UpdateControlPlaneRequest, opts ...operations.Option) (*operations.UpdateControlPlaneResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*operations.UpdateControlPlaneResponse), args.Error(1)
}

func TestApplyControlPlanes(t *testing.T) {
	tests := []struct {
		name string
		orgs map[string]internal.OrganizationManifest

		setup   func(*MockControlPlaneService)
		wantErr bool
	}{
		{
			name: "creates new control plane with labels",
			orgs: map[string]internal.OrganizationManifest{
				"test-org": {
					Environments: map[string]internal.EnvironmentManifest{
						"dev": {
							Type: "DEV",
							Teams: map[string]internal.EnvironmentTeamManifest{
								"team1": {},
							},
						},
					},
				},
			},
			setup: func(m *MockControlPlaneService) {
				m.On("ListControlPlanes", mock.Anything, mock.Anything).Return(&operations.ListControlPlanesResponse{
					ListControlPlanesResponse: &components.ListControlPlanesResponse{
						Data: []components.ControlPlane{},
					},
				}, nil)

				m.On("CreateControlPlane", mock.Anything, mock.MatchedBy(func(req components.CreateControlPlaneRequest) bool {
					return req.Name == "team1-dev" &&
						*req.Description == "Control plane for team team1 in environment dev" &&
						*req.ClusterType == components.CreateControlPlaneRequestClusterType("CLUSTER_TYPE_CONTROL_PLANE") &&
						req.Labels["env"] == "DEV" &&
						req.Labels["team"] == "team1"
				})).Return(&operations.CreateControlPlaneResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "updates existing control plane with new labels",
			orgs: map[string]internal.OrganizationManifest{
				"test-org": {
					Environments: map[string]internal.EnvironmentManifest{
						"prod": {
							Type: "PROD",
							Teams: map[string]internal.EnvironmentTeamManifest{
								"team1": {},
							},
						},
					},
				},
			},
			setup: func(m *MockControlPlaneService) {
				m.On("ListControlPlanes", mock.Anything, mock.Anything).Return(&operations.ListControlPlanesResponse{
					ListControlPlanesResponse: &components.ListControlPlanesResponse{
						Data: []components.ControlPlane{
							{
								ID:          "cp-123",
								Name:        "team1-prod",
								Labels:      map[string]string{},
								Description: kk.String("Old description"),
							},
						},
					},
				}, nil)

				m.On("UpdateControlPlane", mock.Anything, "cp-123", mock.MatchedBy(func(req components.UpdateControlPlaneRequest) bool {
					return *req.Description == "Control plane for team team1 in environment prod" &&
						req.Labels["env"] == "PROD" &&
						req.Labels["team"] == "team1"
				})).Return(&operations.UpdateControlPlaneResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "handles list error",
			orgs: map[string]internal.OrganizationManifest{
				"test-org": {
					Environments: map[string]internal.EnvironmentManifest{
						"dev": {
							Teams: map[string]internal.EnvironmentTeamManifest{
								"team1": {},
							},
						},
					},
				},
			},
			setup: func(m *MockControlPlaneService) {
				m.On("ListControlPlanes", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("list error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCPSvc := &MockControlPlaneService{}
			tt.setup(mockCPSvc)

			err := ApplyControlPlanes(context.Background(), mockCPSvc, tt.orgs)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockCPSvc)
		})
	}
}
