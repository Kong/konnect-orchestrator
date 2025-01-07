package user

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

func TestApplyUsers(t *testing.T) {
	tests := []struct {
		name    string
		emails  []string
		setup   func(*MockUserService, *MockInviteService, *MockTeamMembershipService)
		wantErr bool
	}{
		{
			name:   "creates invite for new user",
			emails: []string{"new@example.com"},
			setup: func(us *MockUserService, is *MockInviteService, tms *MockTeamMembershipService) {
				// First ListUsers call - returns empty
				firstListCall := us.On("ListUsers", mock.Anything, mock.MatchedBy(func(req operations.ListUsersRequest) bool {
					return req.Filter != nil &&
						req.Filter.Email != nil &&
						*req.Filter.Email.StringFieldEqualsFilter.Str == "new@example.com"
				})).Return(&operations.ListUsersResponse{
					UserCollection: &components.UserCollection{
						Data: []components.User{},
					},
				}, nil).Once()

				// InviteUser call - must happen after first ListUsers
				inviteCall := is.On("InviteUser", mock.Anything, &components.InviteUser{
					Email: "new@example.com",
				}).Return(&operations.InviteUserResponse{}, nil).Once().NotBefore(firstListCall)

				// Second ListUsers call - returns the new user, must happen after invite
				secondListCall := us.On("ListUsers", mock.Anything, mock.MatchedBy(func(req operations.ListUsersRequest) bool {
					return req.Filter != nil &&
						req.Filter.Email != nil &&
						*req.Filter.Email.StringFieldEqualsFilter.Str == "new@example.com"
				})).Return(&operations.ListUsersResponse{
					UserCollection: &components.UserCollection{
						Data: []components.User{
							{
								Email: kk.String("new@example.com"),
								ID:    kk.String("user-123"),
							},
						},
					},
				}, nil).Once().NotBefore(inviteCall)

				// AddUserToTeam call - must happen after second ListUsers
				tms.On("AddUserToTeam", mock.Anything, "team-123", &components.AddUserToTeam{
					UserID: "user-123",
				}).Return(&operations.AddUserToTeamResponse{}, nil).Once().NotBefore(secondListCall)
			},
			wantErr: false,
		},
		{
			name:   "skips existing user",
			emails: []string{"existing@example.com"},
			setup: func(us *MockUserService, is *MockInviteService, tms *MockTeamMembershipService) {
				us.On("ListUsers", mock.Anything, mock.MatchedBy(func(req operations.ListUsersRequest) bool {
					return req.Filter != nil &&
						req.Filter.Email != nil &&
						*req.Filter.Email.StringFieldEqualsFilter.Str == "existing@example.com"
				})).Return(&operations.ListUsersResponse{
					UserCollection: &components.UserCollection{
						Data: []components.User{
							{
								Email: kk.String("existing@example.com"),
								ID:    kk.String("user-123"),
							},
						},
					},
				}, nil)

				tms.On("AddUserToTeam", mock.Anything, "team-123", &components.AddUserToTeam{
					UserID: "user-123",
				}).Return(&operations.AddUserToTeamResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "handles list error",
			emails: []string{"error@example.com"},
			setup: func(us *MockUserService, is *MockInviteService, tms *MockTeamMembershipService) {
				us.On("ListUsers", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("list error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserSvc := &MockUserService{}
			mockInviteSvc := &MockInviteService{}
			mockTeamMembershipSvc := &MockTeamMembershipService{}
			tt.setup(mockUserSvc, mockInviteSvc, mockTeamMembershipSvc)

			err := ApplyUsers(context.Background(), mockUserSvc, mockInviteSvc, "team-123", mockTeamMembershipSvc, tt.emails)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockUserSvc, mockInviteSvc)
		})
	}
}
