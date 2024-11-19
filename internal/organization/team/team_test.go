package team

import (
	"context"
	"fmt"
	"testing"

	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Add mock interfaces for user services
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) ListUsers(ctx context.Context, request operations.ListUsersRequest, opts ...operations.Option) (*operations.ListUsersResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*operations.ListUsersResponse), args.Error(1)
}

type MockInviteService struct {
	mock.Mock
}

func (m *MockInviteService) InviteUser(ctx context.Context, request *components.InviteUser, opts ...operations.Option) (*operations.InviteUserResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*operations.InviteUserResponse), args.Error(1)
}

func TestApplyTeam(t *testing.T) {
	tests := []struct {
		name    string
		config  Team
		setup   func(*MockTeamService, *MockUserService, *MockInviteService)
		wantErr bool
	}{
		{
			name: "creates new team when it doesn't exist",
			config: Team{
				Name:        "new-team",
				Description: "A new team",
			},
			setup: func(m *MockTeamService, us *MockUserService, is *MockInviteService) {
				// Setup ListTeams to return empty result
				m.On("ListTeams", mock.Anything, mock.MatchedBy(func(req operations.ListTeamsRequest) bool {
					return req.Filter != nil &&
						req.Filter.Name != nil &&
						*req.Filter.Name.StringFieldEqualsFilter.Str == "new-team"
				})).Return(&operations.ListTeamsResponse{
					TeamCollection: &components.TeamCollection{
						Data: []components.Team{},
					},
				}, nil)

				// Setup CreateTeam expectation
				m.On("CreateTeam", mock.Anything, &components.CreateTeam{
					Name:        "new-team",
					Description: kk.String("A new team"),
				}).Return(&operations.CreateTeamResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "updates existing team",
			config: Team{
				Name:        "existing-team",
				Description: "Updated description",
			},
			setup: func(m *MockTeamService, us *MockUserService, is *MockInviteService) {
				// Setup ListTeams to return existing team
				m.On("ListTeams", mock.Anything, mock.MatchedBy(func(req operations.ListTeamsRequest) bool {
					return req.Filter != nil &&
						req.Filter.Name != nil &&
						*req.Filter.Name.StringFieldEqualsFilter.Str == "existing-team"
				})).Return(&operations.ListTeamsResponse{
					TeamCollection: &components.TeamCollection{
						Data: []components.Team{
							{
								ID:          kk.String("team-123"),
								Name:        kk.String("existing-team"),
								Description: kk.String("Old description"),
							},
						},
					},
				}, nil)

				// Setup UpdateTeam expectation
				m.On("UpdateTeam", mock.Anything, "team-123", &components.UpdateTeam{
					Name:        kk.String("existing-team"),
					Description: kk.String("Updated description"),
				}).Return(&operations.UpdateTeamResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "creates team and handles users",
			config: Team{
				Name:        "team-with-users",
				Description: "Team with users",
				Users:       []string{"new@example.com"},
			},
			setup: func(m *MockTeamService, us *MockUserService, is *MockInviteService) {
				// Setup ListTeams
				m.On("ListTeams", mock.Anything, mock.MatchedBy(func(req operations.ListTeamsRequest) bool {
					return req.Filter != nil &&
						req.Filter.Name != nil &&
						*req.Filter.Name.StringFieldEqualsFilter.Str == "team-with-users"
				})).Return(&operations.ListTeamsResponse{
					TeamCollection: &components.TeamCollection{
						Data: []components.Team{},
					},
				}, nil)

				// Setup CreateTeam
				m.On("CreateTeam", mock.Anything, &components.CreateTeam{
					Name:        "team-with-users",
					Description: kk.String("Team with users"),
				}).Return(&operations.CreateTeamResponse{}, nil)

				// Setup ListUsers for the new user
				us.On("ListUsers", mock.Anything, mock.MatchedBy(func(req operations.ListUsersRequest) bool {
					return req.Filter != nil &&
						req.Filter.Email != nil &&
						*req.Filter.Email.StringFieldEqualsFilter.Str == "new@example.com"
				})).Return(&operations.ListUsersResponse{
					UserCollection: &components.UserCollection{
						Data: []components.User{},
					},
				}, nil)

				is.On("InviteUser", mock.Anything, &components.InviteUser{
					Email: "new@example.com",
				}).Return(&operations.InviteUserResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "handles list error",
			config: Team{
				Name:        "error-team",
				Description: "Error team",
			},
			setup: func(m *MockTeamService, us *MockUserService, is *MockInviteService) {
				m.On("ListTeams", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("list error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTeamSvc := &MockTeamService{}
			mockUserSvc := &MockUserService{}
			mockInviteSvc := &MockInviteService{}
			tt.setup(mockTeamSvc, mockUserSvc, mockInviteSvc)

			err := ApplyTeam(context.Background(), mockTeamSvc, mockUserSvc, mockInviteSvc, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockTeamSvc, mockUserSvc, mockInviteSvc)
		})
	}
}
