package gateway

import (
	"context"

	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
)

// ControlPlaneService handles control plane operations
type ControlPlaneService interface {
	ListControlPlanes(ctx context.Context, request operations.ListControlPlanesRequest, opts ...operations.Option) (*operations.ListControlPlanesResponse, error)
	CreateControlPlane(ctx context.Context, request components.CreateControlPlaneRequest, opts ...operations.Option) (*operations.CreateControlPlaneResponse, error)
}
