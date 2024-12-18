package role

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// RolesAPI defines the interface for the Roles SDK API.
type RoleService interface {
	GetPredefinedRoles(ctx context.Context, opts ...operations.Option) (*operations.GetPredefinedRolesResponse, error)
	ListTeamRoles(ctx context.Context, teamID string, filter *operations.ListTeamRolesQueryParamFilter, opts ...operations.Option) (*operations.ListTeamRolesResponse, error)
	TeamsAssignRole(ctx context.Context, teamID string, assignRole *components.AssignRole, opts ...operations.Option) (*operations.TeamsAssignRoleResponse, error)
}

// EnvironmentType represents the type of environment
type EnvironmentType string

const (
	DEV  EnvironmentType = "DEV"
	PROD EnvironmentType = "PROD"
)

func applyDevRoles(
	ctx context.Context,
	rolesSvc RoleService,
	teamID string,
	cpID string,
	envConfig manifest.Environment) error {
	// Check if the role is already assigned
	listResp, err := rolesSvc.ListTeamRoles(ctx, teamID, &operations.ListTeamRolesQueryParamFilter{
		RoleName:       kk.Pointer(components.CreateStringFieldEqualsFilterStr("Viewer")),
		EntityTypeName: kk.Pointer(components.CreateStringFieldEqualsFilterStr("Control Planes")),
	})
	if err != nil {
		return fmt.Errorf("failed to list team roles: %w", err)
	}

	if listResp != nil && listResp.AssignedRoleCollection != nil && len(listResp.AssignedRoleCollection.Data) > 0 {
		for _, assignedRole := range listResp.AssignedRoleCollection.Data {
			if *assignedRole.RoleName == "Viewer" && *assignedRole.EntityID == cpID {
				return nil
			}
		}
	}
	_, err = rolesSvc.TeamsAssignRole(ctx, teamID, &components.AssignRole{
		RoleName:       kk.Pointer(components.RoleName("Viewer")),
		EntityID:       kk.Pointer(cpID),
		EntityRegion:   kk.Pointer(components.AssignRoleEntityRegion(envConfig.Region)),
		EntityTypeName: kk.Pointer(components.EntityTypeName("Control Planes")),
	})
	if err != nil {
		return fmt.Errorf("failed to assign role to team: %w", err)
	}
	return nil
}

func applyProdRoles(
	ctx context.Context,
	rolesSvc RoleService,
	teamID string,
	cpID string,
	envConfig manifest.Environment) error {
	// Check if the role is already assigned
	listResp, err := rolesSvc.ListTeamRoles(ctx, teamID, &operations.ListTeamRolesQueryParamFilter{
		RoleName:       kk.Pointer(components.CreateStringFieldEqualsFilterStr("Admin")),
		EntityTypeName: kk.Pointer(components.CreateStringFieldEqualsFilterStr("Control Planes")),
	})
	if err != nil {
		return fmt.Errorf("failed to list team roles: %w", err)
	}

	if listResp != nil && listResp.AssignedRoleCollection != nil && len(listResp.AssignedRoleCollection.Data) > 0 {
		for _, assignedRole := range listResp.AssignedRoleCollection.Data {
			if *assignedRole.RoleName == "Admin" && *assignedRole.EntityID == cpID {
				return nil
			}
		}
	}
	_, err = rolesSvc.TeamsAssignRole(ctx, teamID, &components.AssignRole{
		RoleName:       kk.Pointer(components.RoleName("Admin")),
		EntityID:       kk.Pointer(cpID),
		EntityRegion:   kk.Pointer(components.AssignRoleEntityRegion(envConfig.Region)),
		EntityTypeName: kk.Pointer(components.EntityTypeName("Control Planes")),
	})
	if err != nil {
		return fmt.Errorf("failed to assign role to team: %w", err)
	}
	return nil
}

// ApplyRoles assigns the appropriate role to a team based on the environment type
func ApplyRoles(
	ctx context.Context,
	rolesSvc RoleService,
	teamID string,
	cpID string,
	envConfig manifest.Environment) error {
	envType := EnvironmentType(envConfig.Type)
	if envType == DEV {
		return applyDevRoles(ctx, rolesSvc, teamID, cpID, envConfig)
	} else if envType == PROD {
		return applyProdRoles(ctx, rolesSvc, teamID, cpID, envConfig)
	}

	return nil
}
