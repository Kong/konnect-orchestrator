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
		setup   func(*MockUserService, *MockInviteService)
		wantErr bool
	}{
		{
			name:   "creates invite for new user",
			emails: []string{"new@example.com"},
			setup: func(us *MockUserService, is *MockInviteService) {
				// Setup ListUsers to return empty result
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
			name:   "skips existing user",
			emails: []string{"existing@example.com"},
			setup: func(us *MockUserService, is *MockInviteService) {
				us.On("ListUsers", mock.Anything, mock.MatchedBy(func(req operations.ListUsersRequest) bool {
					return req.Filter != nil &&
						req.Filter.Email != nil &&
						*req.Filter.Email.StringFieldEqualsFilter.Str == "existing@example.com"
				})).Return(&operations.ListUsersResponse{
					UserCollection: &components.UserCollection{
						Data: []components.User{
							{
								Email: kk.String("existing@example.com"),
							},
						},
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "handles list error",
			emails: []string{"error@example.com"},
			setup: func(us *MockUserService, is *MockInviteService) {
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
			tt.setup(mockUserSvc, mockInviteSvc)

			err := ApplyUsers(context.Background(), mockUserSvc, mockInviteSvc, tt.emails)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockUserSvc, mockInviteSvc)
		})
	}
}
