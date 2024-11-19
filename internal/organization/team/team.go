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

// Team represents a team in the organization, with a name and description.
type Team struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Users       []string `json:"users" yaml:"users"`
}

func ApplyTeam(ctx context.Context, svc TeamService, userSvc user.UserService, inviteSvc user.InviteService, config Team) error {
	// Step 1: Check if team already exists
	equalsFilter := components.CreateStringFieldEqualsFilterStr(config.Name)

	listResp, err := svc.ListTeams(ctx,
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

	var existingTeamID string
	// TODO: This ignores pagination
	for _, team := range listResp.TeamCollection.Data {
		if *team.Name == config.Name {
			existingTeamID = *team.ID
			break
		}
	}

	// Step 2: Create or Update based on existence
	if existingTeamID == "" {
		// Team does not exist, create it
		_, err = svc.CreateTeam(ctx, &components.CreateTeam{
			Name:        config.Name,
			Description: kk.String(config.Description),
		})
		if err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}
	} else {
		// Team exists, update it
		_, err = svc.UpdateTeam(ctx, existingTeamID, &components.UpdateTeam{
			Name:        kk.String(config.Name),
			Description: kk.String(config.Description),
		})
		if err != nil {
			return fmt.Errorf("failed to update team: %w", err)
		}
	}

	// Step 3: Apply users if any are specified
	if len(config.Users) > 0 {
		if err := user.ApplyUsers(ctx, userSvc, inviteSvc, config.Users); err != nil {
			return fmt.Errorf("failed to apply users: %w", err)
		}
	}

	return nil
}
