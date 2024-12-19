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

type TeamMembershipService interface {
	ListTeamUsers(ctx context.Context, request operations.ListTeamUsersRequest, opts ...operations.Option) (*operations.ListTeamUsersResponse, error)
	AddUserToTeam(ctx context.Context, teamID string, addUserToTeam *components.AddUserToTeam, opts ...operations.Option) (*operations.AddUserToTeamResponse, error)
}

// User represents a user in the organization
type User struct {
	Email string `json:"email" yaml:"email"`
}

func lookupUserByEmail(ctx context.Context, userSvc UserService, email string) (exists bool, userID string, err error) {
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
		return false, "", fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range listResp.UserCollection.Data {
		if *user.Email == email {
			return true, *user.ID, nil
		}
	}
	return false, "", nil
}

func ApplyUsers(
	ctx context.Context,
	userSvc UserService,
	inviteSvc InviteService,
	teamID string,
	teamMembershipSvc TeamMembershipService,
	emails []string,
) error {
	for _, email := range emails {
		// Check if user already exists
		userExists, userID, err := lookupUserByEmail(ctx, userSvc, email)
		if err != nil {
			return err
		}

		if !userExists {
			_, err := inviteSvc.InviteUser(ctx, &components.InviteUser{
				Email: email,
			})
			if err != nil {
				return fmt.Errorf("failed to create invite for %s: %w", email, err)
			}
			// Lookup the newly invited user
			exists, newUserID, lookupErr := lookupUserByEmail(ctx, userSvc, email)
			if lookupErr != nil {
				return fmt.Errorf("failed to lookup invited user %s: %w", email, lookupErr)
			}
			if !exists {
				return fmt.Errorf("failed to find newly invited user %s", email)
			}
			userID = newUserID
		}

		// TODO: Check for membership first
		_, err = teamMembershipSvc.AddUserToTeam(ctx, teamID, &components.AddUserToTeam{
			UserID: userID,
		})
		// for now ignore errors
		//if err != nil {
		//	return fmt.Errorf("failed to add user to team: %w", err)
		//}
	}

	return nil
}
