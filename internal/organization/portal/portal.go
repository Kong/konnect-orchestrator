package portal

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	kk "github.com/Kong/sdk-konnect-go-internal"
	"github.com/Kong/sdk-konnect-go-internal/models/components"
	"github.com/Kong/sdk-konnect-go-internal/models/operations"
)

type PortalsConfigService interface {
	ListPortals(ctx context.Context,
		request operations.ListPortalsRequest,
		opts ...operations.Option) (*operations.ListPortalsResponse, error)
	CreatePortal(ctx context.Context,
		request components.CreatePortalV3,
		opts ...operations.Option) (*operations.CreatePortalResponse, error)
	UpdatePortal(ctx context.Context,
		portalID string,
		updatePortalV3 components.UpdatePortalV3,
		opts ...operations.Option) (*operations.UpdatePortalResponse, error)
}

type ApisConfigService interface {
	ListApis(ctx context.Context,
		request operations.ListApisRequest,
		opts ...operations.Option) (*operations.ListApisResponse, error)
	CreateAPI(ctx context.Context,
		request components.CreateAPIRequest,
		opts ...operations.Option) (*operations.CreateAPIResponse, error)
}

// If you change the name of a portal, a new one will be created an the old one remains
func ApplyPortalConfig(
	ctx context.Context,
	portalName string,
	portalsConfigService PortalsConfigService,
	apisConfigService ApisConfigService) (string, error) {

	var portalId string

	portals, err := portalsConfigService.ListPortals(ctx, operations.ListPortalsRequest{
		Filter: &components.PortalFilterParameters{
			Name: &components.StringFieldFilter{
				StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
					Str: kk.String(portalName),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	if len(portals.ListPortalsResponseV3.Data) < 1 {
		newPortal, err := portalsConfigService.CreatePortal(ctx, components.CreatePortalV3{
			Name:        portalName,
			DisplayName: kk.String(portalName),
		})
		if err != nil {
			return "", err
		}
		portalId = newPortal.PortalResponseV3.ID
	} else {
		portalId = portals.ListPortalsResponseV3.Data[0].ID
		_, err = portalsConfigService.UpdatePortal(ctx, portalId, components.UpdatePortalV3{
			DisplayName: kk.String(portalName),
		})
		if err != nil {
			return "", err
		}
	}

	return portalId, nil
}

func ApplyApiConfig(ctx context.Context,
	apisConfigService ApisConfigService,
	apiName string,
	serviceConfig manifest.Service,
	rawSpec []byte,
	portalId string) error {

	var spec map[string]interface{}
	err := yaml.Unmarshal(rawSpec, &spec)
	if err != nil {
		return err
	}

	// Extract the version
	var version string
	if info, ok := spec["info"].(map[string]interface{}); ok {
		version = info["version"].(string)
	}

	//var api *components.APIResponseSchema

	// **************************************************************************
	// Search for existing API by name and version
	resp, err := apisConfigService.ListApis(ctx,
		operations.ListApisRequest{
			Filter: &components.APIFilterParameters{
				Name: &components.StringFieldFilter{
					StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
						Str: kk.String(apiName),
					},
				},
				Version: &components.StringFieldFilter{
					StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
						Str: kk.String(version),
					},
				},
			},
		})
	if err != nil {
		return err
	}
	// **************************************************************************

	// **************************************************************************
	// Create a new or use the existing API
	if len(resp.ListAPIResponse.Data) < 1 {
		_, err := apisConfigService.CreateAPI(ctx,
			components.CreateAPIRequest{
				Name:        apiName,
				Version:     kk.String(version),
				Description: kk.String(serviceConfig.Description),
			})
		if err != nil {
			return err
		}
		//api = createResponse.APIResponseSchema
	} else {
		//api = &resp.ListAPIResponse.Data[0]
	}
	// **************************************************************************

	// **************************************************************************
	// Update the API Spec
	// **************************************************************************

	// **************************************************************************
	// Publish the API to the portal

	// **************************************************************************

	return err
}
