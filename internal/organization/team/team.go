package team

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal/organization/user"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// https://docs.konghq.com/konnect/api/identity-management/latest/#/Teams
type TeamService interface {
	ListTeams(ctx context.Context, request operations.ListTeamsRequest, opts ...operations.Option) (*operations.ListTeamsResponse, error)
	CreateTeam(ctx context.Context, request *components.CreateTeam, opts ...operations.Option) (*operations.CreateTeamResponse, error)
	GetTeam(ctx context.Context, teamID string, opts ...operations.Option) (*operations.GetTeamResponse, error)
	UpdateTeam(ctx context.Context, teamID string, updateTeam *components.UpdateTeam, opts ...operations.Option) (*operations.UpdateTeamResponse, error)
}

type TeamMembershipService interface {
	ListTeamUsers(ctx context.Context, request operations.ListTeamUsersRequest, opts ...operations.Option) (*operations.ListTeamUsersResponse, error)
	AddUserToTeam(ctx context.Context, teamID string, addUserToTeam *components.AddUserToTeam, opts ...operations.Option) (*operations.AddUserToTeamResponse, error)
}

// Team represents a team in the organization, with a name and description.
type Team struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Users       []string `json:"users" yaml:"users"`
}

func ApplyTeam(ctx context.Context,
	teamSvc TeamService,
	teamMembershipSvc TeamMembershipService,
	userSvc user.UserService,
	inviteSvc user.InviteService,
	teamConfig Team) error {
	// Step 1: Check if team already exists
	equalsFilter := components.CreateStringFieldEqualsFilterStr(teamConfig.Name)

	listTeamsResponse, err := teamSvc.ListTeams(ctx,
		operations.ListTeamsRequest{
			Filter: &operations.ListTeamsQueryParamFilter{
				Name: &components.StringFieldFilter{
					StringFieldEqualsFilter: &equalsFilter,
					Type:                    components.StringFieldFilterTypeStringFieldEqualsFilter,
				},
			},
		})
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	var teamID string
	// TODO: This ignores pagination
	for _, team := range listTeamsResponse.TeamCollection.Data {
		if *team.Name == teamConfig.Name {
			teamID = *team.ID
			break
		}
	}

	// Step 2: Create or Update based on existence
	if teamID == "" {
		// Team does not exist, create it
		createTeamResponse, err := teamSvc.CreateTeam(ctx, &components.CreateTeam{
			Name:        teamConfig.Name,
			Description: kk.String(teamConfig.Description),
		})
		if err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}
		// update the existingTeamID
		teamID = *createTeamResponse.Team.ID
	} else {
		// Team exists, update it
		_, err = teamSvc.UpdateTeam(ctx, teamID, &components.UpdateTeam{
			Name:        kk.String(teamConfig.Name),
			Description: kk.String(teamConfig.Description),
		})
		if err != nil {
			return fmt.Errorf("failed to update team: %w", err)
		}
	}

	// Step 3: Apply users if any are specified
	if len(teamConfig.Users) > 0 {
		if err := user.ApplyUsers(ctx, userSvc, inviteSvc, teamConfig.Users); err != nil {
			return fmt.Errorf("failed to apply users: %w", err)
		}

		// list the current team users
		// TODO: This ignores pagination
		listTeamUsersResponse, err := teamMembershipSvc.ListTeamUsers(ctx, operations.ListTeamUsersRequest{
			TeamID: teamID,
		})
		if err != nil {
			return fmt.Errorf("failed to list team users: %w", err)
		}

		// Get all users from user service to map emails to user IDs
		for _, userEmail := range teamConfig.Users {
			// Look up user ID
			equalsFilter := components.CreateStringFieldEqualsFilterStr(userEmail)
			listUsersResponse, err := userSvc.ListUsers(ctx, operations.ListUsersRequest{
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

			// Get the user ID
			var userID string
			for _, user := range listUsersResponse.UserCollection.Data {
				if *user.Email == userEmail {
					userID = *user.ID
					break
				}
			}

			// Check if user is already in team
			var userInTeam bool
			for _, existingUser := range listTeamUsersResponse.UserCollection.Data {
				if *existingUser.Email == userEmail {
					userInTeam = true
					break
				}
			}

			// Add user to team if not already present
			if !userInTeam {
				_, err = teamMembershipSvc.AddUserToTeam(ctx, teamID, &components.AddUserToTeam{
					UserID: userID,
				})
				if err != nil {
					return fmt.Errorf("failed to add user %s to team: %w", userEmail, err)
				}
			}
		}
	}

	return nil
}
