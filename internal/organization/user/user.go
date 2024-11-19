package user

import (
	"context"
	"fmt"

	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// UserService defines the interface for user-related operations
type UserService interface {
	ListUsers(ctx context.Context, request operations.ListUsersRequest, opts ...operations.Option) (*operations.ListUsersResponse, error)
}

// InviteService defines the interface for invite-related operations
type InviteService interface {
	InviteUser(ctx context.Context, request *components.InviteUser, opts ...operations.Option) (*operations.InviteUserResponse, error)
}

// User represents a user in the organization
type User struct {
	Email string `json:"email" yaml:"email"`
}

func ApplyUsers(ctx context.Context, userSvc UserService, inviteSvc InviteService, emails []string) error {
	for _, email := range emails {
		// Check if user already exists
		equalsFilter := components.CreateStringFieldEqualsFilterStr(email)
		listResp, err := userSvc.ListUsers(ctx, operations.ListUsersRequest{
			Filter: &operations.ListUsersQueryParamFilter{
				Email: &components.StringFieldFilter{
					StringFieldEqualsFilter: &equalsFilter,
					Type:                    components.StringFieldFilterTypeStringFieldEqualsFilter,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		var userExists bool
		for _, user := range listResp.UserCollection.Data {
			if *user.Email == email {
				userExists = true
				break
			}
		}

		if !userExists {
			_, err = inviteSvc.InviteUser(ctx, &components.InviteUser{
				Email: email,
			})
			if err != nil {
				return fmt.Errorf("failed to create invite for %s: %w", email, err)
			}
		}
	}

	return nil
}
