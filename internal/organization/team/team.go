package team

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/organization/user"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// https://docs.konghq.com/konnect/api/identity-management/latest/#/Teams
type Service interface {
	ListTeams(ctx context.Context,
		request operations.ListTeamsRequest,
		opts ...operations.Option) (*operations.ListTeamsResponse, error)
	CreateTeam(ctx context.Context,
		request *components.CreateTeam,
		opts ...operations.Option) (*operations.CreateTeamResponse, error)
	GetTeam(ctx context.Context,
		teamID string,
		opts ...operations.Option) (*operations.GetTeamResponse, error)
	UpdateTeam(ctx context.Context,
		teamID string,
		updateTeam *components.UpdateTeam,
		opts ...operations.Option) (*operations.UpdateTeamResponse, error)
}

func ApplyTeam(ctx context.Context,
	teamSvc Service,
	teamMembershipSvc user.TeamMembershipService,
	userSvc user.Service,
	inviteSvc user.InviteService,
	teamName string,
	teamConfig manifest.Team,
) (string, error) {
	// Step 1: Check if team exists
	var teamID string
	team, err := findTeamByName(ctx, teamSvc, teamName)
	if err != nil {
		return "", fmt.Errorf("failed to find team: %w", err)
	}

	// Step 2: Create or Update based on existence
	if team == nil {
		// Create new team
		resp, err := teamSvc.CreateTeam(ctx, &components.CreateTeam{
			Name:        teamName,
			Description: teamConfig.Description,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create team: %w", err)
		}
		teamID = *resp.Team.ID
	} else {
		// Update existing team
		teamID = *team.ID
		// Only update if there are differences
		needsUpdate := false
		if *team.Name != teamName {
			needsUpdate = true
		}
		if (team.Description == nil && teamConfig.Description != nil) ||
			(team.Description != nil && *team.Description != *teamConfig.Description) {
			needsUpdate = true
		}
		if needsUpdate {
			_, err = teamSvc.UpdateTeam(ctx, teamID, &components.UpdateTeam{
				Name:        kk.String(teamName),
				Description: teamConfig.Description,
			})
			if err != nil {
				return "", fmt.Errorf("failed to update team: %w", err)
			}
		}
	}

	// Step 3: Apply users
	if len(teamConfig.Users) > 0 {
		if err := user.ApplyUsers(ctx, userSvc, inviteSvc, teamID, teamMembershipSvc, teamConfig.Users); err != nil {
			return "", fmt.Errorf("failed to apply users: %w", err)
		}
	}

	return teamID, nil
}

// findTeamByName searches for a team by name and returns its ID if found, empty string if not found
func findTeamByName(ctx context.Context, teamSvc Service, teamName string) (*components.Team, error) {
	// List all teams
	resp, err := teamSvc.ListTeams(ctx, operations.ListTeamsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	// Search for team with matching name
	for _, team := range resp.TeamCollection.Data {
		if team.Name != nil && *team.Name == teamName {
			return &team, nil
		}
	}

	// Team not found
	return nil, nil
}
