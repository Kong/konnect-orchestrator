package role

import (
	"context"
	"testing"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApplyRoles(t *testing.T) {
	tests := []struct {
		name    string
		teamID  string
		cpID    string
		envType EnvironmentType
		region  string
		setup   func(*MockRoleService)
		wantErr bool
	}{
		{
			name:    "successfully assigns admin role for DEV environment",
			envType: DEV,
			region:  "us",
			teamID:  "team-123",
			cpID:    "cp-123",
			setup: func(m *MockRoleService) {
				// Mock ListTeamRoles - return empty list (no existing roles)
				m.On("ListTeamRoles",
					mock.Anything,
					"team-123",
					&operations.ListTeamRolesQueryParamFilter{
						RoleName:       kk.Pointer(components.CreateStringFieldEqualsFilterStr("Admin")),
						EntityTypeName: kk.Pointer(components.CreateStringFieldEqualsFilterStr("Control Planes")),
					},
				).Return(&operations.ListTeamRolesResponse{
					AssignedRoleCollection: &components.AssignedRoleCollection{
						Data: []components.AssignedRole{},
					},
				}, nil)

				// Mock TeamsAssignRole - successful assignment
				m.On("TeamsAssignRole",
					mock.Anything,
					"team-123",
					&components.AssignRole{
						RoleName:       kk.Pointer(components.RoleName("Admin")),
						EntityID:       kk.Pointer("cp-123"),
						EntityRegion:   kk.Pointer(components.AssignRoleEntityRegion("us")),
						EntityTypeName: kk.Pointer(components.EntityTypeName("Control Planes")),
					},
				).Return(&operations.TeamsAssignRoleResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "skips assignment when admin role already exists for DEV",
			envType: DEV,
			region:  "us",
			teamID:  "team-123",
			cpID:    "cp-123",
			setup: func(m *MockRoleService) {
				m.On("ListTeamRoles",
					mock.Anything,
					"team-123",
					&operations.ListTeamRolesQueryParamFilter{
						RoleName:       kk.Pointer(components.CreateStringFieldEqualsFilterStr("Admin")),
						EntityTypeName: kk.Pointer(components.CreateStringFieldEqualsFilterStr("Control Planes")),
					},
				).Return(&operations.ListTeamRolesResponse{
					AssignedRoleCollection: &components.AssignedRoleCollection{
						Data: []components.AssignedRole{
							{
								RoleName: kk.String("Admin"),
								EntityID: kk.String("cp-123"),
							},
						},
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:    "successfully assigns viewer role for PROD environment",
			envType: PROD,
			region:  "us",
			teamID:  "team-123",
			cpID:    "cp-123",
			setup: func(m *MockRoleService) {
				m.On("ListTeamRoles",
					mock.Anything,
					"team-123",
					&operations.ListTeamRolesQueryParamFilter{
						RoleName:       kk.Pointer(components.CreateStringFieldEqualsFilterStr("Viewer")),
						EntityTypeName: kk.Pointer(components.CreateStringFieldEqualsFilterStr("Control Planes")),
					},
				).Return(&operations.ListTeamRolesResponse{
					AssignedRoleCollection: &components.AssignedRoleCollection{
						Data: []components.AssignedRole{},
					},
				}, nil)

				m.On("TeamsAssignRole",
					mock.Anything,
					"team-123",
					&components.AssignRole{
						RoleName:       kk.Pointer(components.RoleName("Viewer")),
						EntityID:       kk.Pointer("cp-123"),
						EntityRegion:   kk.Pointer(components.AssignRoleEntityRegion("us")),
						EntityTypeName: kk.Pointer(components.EntityTypeName("Control Planes")),
					},
				).Return(&operations.TeamsAssignRoleResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "handles ListTeamRoles error",
			envType: DEV,
			region:  "us",
			teamID:  "team-123",
			cpID:    "cp-123",
			setup: func(m *MockRoleService) {
				m.On("ListTeamRoles",
					mock.Anything,
					"team-123",
					mock.Anything,
				).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "handles TeamsAssignRole error",
			envType: DEV,
			region:  "us",
			teamID:  "team-123",
			cpID:    "cp-123",
			setup: func(m *MockRoleService) {
				m.On("ListTeamRoles",
					mock.Anything,
					"team-123",
					mock.Anything,
				).Return(&operations.ListTeamRolesResponse{
					AssignedRoleCollection: &components.AssignedRoleCollection{
						Data: []components.AssignedRole{},
					},
				}, nil)

				m.On("TeamsAssignRole",
					mock.Anything,
					"team-123",
					mock.Anything,
				).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSvc := NewMockRoleService(t)
			if tt.setup != nil {
				tt.setup(mockRoleSvc)
			}

			err := ApplyRoles(context.Background(), mockRoleSvc, tt.teamID, tt.cpID,
				manifest.Environment{
					Type:   string(tt.envType),
					Region: tt.region,
				},
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRoleSvc.AssertExpectations(t)
		})
	}
}
