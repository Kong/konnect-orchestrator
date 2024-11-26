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

// Team represents a team in the organization
type Team struct {
	Name        string                   `json:"name" yaml:"name"`
	Description string                   `json:"description" yaml:"description"`
	Users       []string                 `json:"users" yaml:"users"`
	Services    map[string]ServiceConfig `json:"services" yaml:"services"`
}

// ServiceConfig represents a service configuration
type ServiceConfig struct {
	Name        string `json:"name" yaml:"name"`
	VCS         string `json:"vcs" yaml:"vcs"`
	Description string `json:"description" yaml:"description"`
	SpecPath    string `json:"spec-path" yaml:"spec-path"`
	KongPath    string `json:"kong-path" yaml:"kong-path"`
}

func ApplyTeam(ctx context.Context,
	teamSvc TeamService,
	teamMembershipSvc TeamMembershipService,
	userSvc user.UserService,
	inviteSvc user.InviteService,
	teamConfig Team) error {

	// Step 1: Check if team exists
	teamID, err := findTeamByName(ctx, teamSvc, teamConfig.Name)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// Step 2: Create or Update based on existence
	if teamID == "" {
		// Create new team
		createResp, err := teamSvc.CreateTeam(ctx, &components.CreateTeam{
			Name:        teamConfig.Name,
			Description: kk.String(teamConfig.Description),
		})
		if err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}
		teamID = *createResp.Team.ID
	} else {
		// Update existing team
		_, err = teamSvc.UpdateTeam(ctx, teamID, &components.UpdateTeam{
			Name:        kk.String(teamConfig.Name),
			Description: kk.String(teamConfig.Description),
		})
		if err != nil {
			return fmt.Errorf("failed to update team: %w", err)
		}
	}

	// Step 3: Apply users
	if len(teamConfig.Users) > 0 {
		if err := user.ApplyUsers(ctx, userSvc, inviteSvc, teamConfig.Users); err != nil {
			return fmt.Errorf("failed to apply users: %w", err)
		}
	}

	return nil
}

// findTeamByName searches for a team by name and returns its ID if found, empty string if not found
func findTeamByName(ctx context.Context, teamSvc TeamService, teamName string) (string, error) {
	// List all teams
	resp, err := teamSvc.ListTeams(ctx, operations.ListTeamsRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to list teams: %w", err)
	}

	// Search for team with matching name
	for _, team := range resp.TeamCollection.Data {
		if team.Name != nil && *team.Name == teamName {
			return *team.ID, nil
		}
	}

	// Team not found
	return "", nil
}
