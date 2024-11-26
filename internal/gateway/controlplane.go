package gateway

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// ControlPlaneService handles control plane operations
type ControlPlaneService interface {
	ListControlPlanes(ctx context.Context, request operations.ListControlPlanesRequest, opts ...operations.Option) (*operations.ListControlPlanesResponse, error)
	CreateControlPlane(ctx context.Context, request components.CreateControlPlaneRequest, opts ...operations.Option) (*operations.CreateControlPlaneResponse, error)
	UpdateControlPlane(ctx context.Context, id string, request components.UpdateControlPlaneRequest, opts ...operations.Option) (*operations.UpdateControlPlaneResponse, error)
}

// ApplyControlPlanes ensures that control planes exist for each team in each environment
func ApplyControlPlanes(ctx context.Context, cpSvc ControlPlaneService, orgs map[string]internal.OrganizationManifest) error {
	for orgName, org := range orgs {
		// Process each environment in the organization
		for envName, env := range org.Environments {
			// Process each team in the environment
			for teamName := range env.Teams {
				// Control plane name follows convention: team-name-environment-name
				cpName := fmt.Sprintf("%s-%s", teamName, envName)

				// Create labels map with both env and team labels
				labels := map[string]string{
					"env":  env.Type,
					"team": teamName,
				}

				// Check if control plane exists
				cp, err := findControlPlane(ctx, cpSvc, cpName)
				if err != nil {
					return fmt.Errorf("failed to check control plane existence for %s in org %s: %w", cpName, orgName, err)
				}

				if cp == nil {
					// Create new control plane
					_, err := cpSvc.CreateControlPlane(ctx, components.CreateControlPlaneRequest{
						Name:        cpName,
						Description: kk.String(fmt.Sprintf("Control plane for team %s in environment %s", teamName, envName)),
						ClusterType: kk.Pointer(components.CreateControlPlaneRequestClusterType("CLUSTER_TYPE_CONTROL_PLANE")),
						Labels:      labels,
					})
					if err != nil {
						return fmt.Errorf("failed to create control plane %s in org %s: %w", cpName, orgName, err)
					}
				} else {
					// Update existing control plane if needed
					needsUpdate := false
					description := fmt.Sprintf("Control plane for team %s in environment %s", teamName, envName)

					if cp.Description == nil || *cp.Description != description {
						needsUpdate = true
					}

					// Check if labels need updating
					if !mapsEqual(cp.Labels, labels) {
						needsUpdate = true
					}

					if needsUpdate {
						_, err := cpSvc.UpdateControlPlane(ctx, cp.ID, components.UpdateControlPlaneRequest{
							Description: kk.String(description),
							Labels:      labels,
						})
						if err != nil {
							return fmt.Errorf("failed to update control plane %s in org %s: %w", cpName, orgName, err)
						}
					}
				}
			}
		}
	}
	return nil
}

// findControlPlane returns the control plane if it exists, nil if it doesn't
func findControlPlane(ctx context.Context, cpSvc ControlPlaneService, name string) (*components.ControlPlane, error) {
	resp, err := cpSvc.ListControlPlanes(ctx, operations.ListControlPlanesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list control planes: %w", err)
	}

	for _, cp := range resp.ListControlPlanesResponse.Data {
		if cp.Name == name {
			return &cp, nil
		}
	}

	return nil, nil
}

// mapsEqual compares two string maps for equality
func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}
