package gateway

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// ControlPlaneService handles control plane operations
type ControlPlaneService interface {
	ListControlPlanes(ctx context.Context,
		request operations.ListControlPlanesRequest,
		opts ...operations.Option) (*operations.ListControlPlanesResponse, error)
	CreateControlPlane(ctx context.Context,
		request components.CreateControlPlaneRequest,
		opts ...operations.Option) (*operations.CreateControlPlaneResponse, error)
	UpdateControlPlane(ctx context.Context,
		id string,
		request components.UpdateControlPlaneRequest,
		opts ...operations.Option) (*operations.UpdateControlPlaneResponse, error)
}

// ApplyControlPlane ensures that a control plane exists for a team in a specific environment
func ApplyControlPlane(
	ctx context.Context,
	cpSvc ControlPlaneService,
	envName string,
	env manifest.Environment,
	teamName string,
) (string, error) {
	var cpName string
	if env.Teams == nil || env.Teams[teamName].ControlPlaneName == nil {
		// Control plane name follows convention: team-name-environment-name
		cpName = fmt.Sprintf("%s-%s", teamName, envName)
	}

	// Create labels map with both env and team labels
	labels := map[string]string{
		"env":  env.Type,
		"team": teamName,
	}

	// Check if control plane exists
	cp, err := findControlPlane(ctx, cpSvc, cpName)
	if err != nil {
		return "", fmt.Errorf("failed to check control plane existence for %s: %w", cpName, err)
	}

	if cp == nil {
		var clusterType components.CreateControlPlaneRequestClusterType
		if env.Type == "DEV" {
			clusterType = components.CreateControlPlaneRequestClusterType("CLUSTER_TYPE_SERVERLESS")
		} else {
			clusterType = components.CreateControlPlaneRequestClusterType("CLUSTER_TYPE_CONTROL_PLANE")
		}

		// Create new control plane
		resp, err := cpSvc.CreateControlPlane(ctx, components.CreateControlPlaneRequest{
			Name:        cpName,
			Description: kk.String(fmt.Sprintf("Control plane for team %s in environment %s", teamName, envName)),
			ClusterType: kk.Pointer(clusterType),
			Labels:      labels,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create control plane %s: %w", cpName, err)
		}
		return resp.ControlPlane.ID, nil
	}

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
			return "", fmt.Errorf("failed to update control plane %s: %w", cpName, err)
		}
	}
	return cp.ID, nil
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
