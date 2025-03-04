package portal

import (
	"context"

	"sigs.k8s.io/yaml"

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
	UpdateAPI(ctx context.Context,
		apiID string,
		updateAPIRequest components.UpdateAPIRequest,
		opts ...operations.Option) (*operations.UpdateAPIResponse, error)
}

type APISpecsConfigService interface {
	CreateAPISpec(ctx context.Context,
		apiID string,
		createAPISpecRequest components.CreateAPISpecRequest,
		opts ...operations.Option) (*operations.CreateAPISpecResponse, error)
	UpdateAPISpec(ctx context.Context,
		request operations.UpdateAPISpecRequest,
		opts ...operations.Option) (*operations.UpdateAPISpecResponse, error)
	ListAPISpecs(ctx context.Context,
		request operations.ListAPISpecsRequest,
		opts ...operations.Option) (*operations.ListAPISpecsResponse, error)
}

type APIPublicationConfigService interface {
	PublishAPIToPortal(ctx context.Context,
		request operations.PublishAPIToPortalRequest,
		opts ...operations.Option) (*operations.PublishAPIToPortalResponse, error)
}

type APIImplementationConfigService interface {
	CreateAPIImplementation(ctx context.Context,
		apiID string,
		apiImplementation components.APIImplementation,
		opts ...operations.Option) (*operations.CreateAPIImplementationResponse, error)
	ListAPIImplementations(ctx context.Context,
		request operations.ListAPIImplementationsRequest,
		opts ...operations.Option) (*operations.ListAPIImplementationsResponse, error)
}

// If you change the name of a portal, a new one will be created an the old one remains
func ApplyPortalConfig(
	ctx context.Context,
	portalDisplayName string,
	envName string,
	envType string,
	portalsConfigService PortalsConfigService,
	_ ApisConfigService,
	labels map[string]string,
) (string, error) {
	var portalID string

	portals, err := portalsConfigService.ListPortals(ctx, operations.ListPortalsRequest{
		Filter: &components.PortalFilterParameters{
			Name: &components.StringFieldFilter{
				StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
					Str: kk.String(envName),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// PROD portals are open by default, dev portals are secured
	authEnabled := envType != "PROD"
	visibility := "public"
	if envType != "PROD" {
		visibility = "private"
		portalDisplayName = portalDisplayName + " (" + envName + ")"
	}

	if len(portals.ListPortalsResponseV3.Data) < 1 {
		newPortal, err := portalsConfigService.CreatePortal(ctx, components.CreatePortalV3{
			Name:                             envName,
			DisplayName:                      kk.String(portalDisplayName),
			AuthenticationEnabled:            kk.Bool(authEnabled),
			DefaultAPIVisibility:             components.DefaultAPIVisibility(visibility).ToPointer(),
			DefaultPageVisibility:            components.DefaultPageVisibility(visibility).ToPointer(),
			DefaultApplicationAuthStrategyID: nil,
			Labels:                           toPortalLabels(labels),
		})
		if err != nil {
			return "", err
		}
		portalID = newPortal.PortalResponseV3.ID
	} else {
		portalID = portals.ListPortalsResponseV3.Data[0].ID
		_, err = portalsConfigService.UpdatePortal(ctx, portalID, components.UpdatePortalV3{
			DisplayName:                      kk.String(envName),
			AuthenticationEnabled:            kk.Bool(authEnabled),
			DefaultAPIVisibility:             components.UpdatePortalV3DefaultAPIVisibility(visibility).ToPointer(),
			DefaultPageVisibility:            components.UpdatePortalV3DefaultPageVisibility(visibility).ToPointer(),
			DefaultApplicationAuthStrategyID: nil,
			Labels:                           toPortalLabels(labels),
		})
		if err != nil {
			return "", err
		}
	}

	return portalID, nil
}

func ApplyAPIConfig(ctx context.Context,
	apisConfigService ApisConfigService,
	apiSpecsConfigService APISpecsConfigService,
	apiPubConfigService APIPublicationConfigService,
	apiImplementationConfigService APIImplementationConfigService,
	apiName string,
	serviceConfig manifest.Service,
	rawSpec []byte,
	portalID string,
	cpID string,
	gwSvcID string,
	labels map[string]string,
) (string, error) {
	var spec map[string]interface{}
	err := yaml.Unmarshal(rawSpec, &spec)
	if err != nil {
		return "", err
	}

	// Extract the version
	var version string
	if info, ok := spec["info"].(map[string]interface{}); ok {
		version = info["version"].(string)
	}

	var api *components.APIResponseSchema

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
		return "", err
	}
	// **************************************************************************

	// **************************************************************************
	// Create a new or use the existing API
	if len(resp.ListAPIResponse.Data) < 1 {
		createResponse, err := apisConfigService.CreateAPI(ctx,
			components.CreateAPIRequest{
				Name:        apiName,
				Version:     kk.String(version),
				Description: kk.String(*serviceConfig.Description),
				Labels:      labels,
			})
		if err != nil {
			return "", err
		}
		api = createResponse.APIResponseSchema
	} else {
		api = &resp.ListAPIResponse.Data[0]
		_, err = apisConfigService.UpdateAPI(ctx,
			resp.ListAPIResponse.Data[0].ID,
			components.UpdateAPIRequest{
				Name:        kk.String(apiName),
				Version:     kk.String(version),
				Description: kk.String(*serviceConfig.Description),
				Labels:      toPortalLabels(labels),
			})
		if err != nil {
			return "", err
		}
	}
	// **************************************************************************

	// **************************************************************************
	// Update the API Spec
	listSpecResponse, err := apiSpecsConfigService.ListAPISpecs(ctx, operations.ListAPISpecsRequest{
		APIID: api.ID,
	})
	if err != nil {
		return "", err
	}
	if len(listSpecResponse.ListAPISpecResponse.Data) < 1 {
		_, err = apiSpecsConfigService.CreateAPISpec(ctx, api.ID, components.CreateAPISpecRequest{
			Content: string(rawSpec),
			Type:    components.APISpecTypeOas3.ToPointer(),
		})
		if err != nil {
			return "", err
		}
	} else {
		_, err = apiSpecsConfigService.UpdateAPISpec(ctx, operations.UpdateAPISpecRequest{
			APIID:  api.ID,
			SpecID: listSpecResponse.ListAPISpecResponse.Data[0].ID,
			APISpec: components.APISpec{
				Content: kk.String(string(rawSpec)),
				Type:    components.APISpecTypeOas3.ToPointer(),
			},
		})
		if err != nil {
			return "", err
		}
	}
	// **************************************************************************

	// **************************************************************************
	// Publish the API to the portal
	_, err = apiPubConfigService.PublishAPIToPortal(ctx,
		operations.PublishAPIToPortalRequest{
			APIID:    api.ID,
			PortalID: portalID,
		})
	if err != nil {
		return "", err
	}
	// **************************************************************************

	if gwSvcID == "" {
		return api.ID, nil
	}

	// **************************************************************************
	// Create an API Implementation
	apiImpls, err := apiImplementationConfigService.ListAPIImplementations(ctx, operations.ListAPIImplementationsRequest{
		Filter: &components.APIImplementationFilterParameters{
			APIID: &components.UUIDFieldFilter{
				StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
					Str: kk.String(api.ID),
				},
			},
			ControlPlaneID: &components.UUIDFieldFilter{
				StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
					Str: kk.String(cpID),
				},
			},
			ServiceID: &components.UUIDFieldFilter{
				StringFieldEqualsFilter: &components.StringFieldEqualsFilter{
					Str: kk.String(gwSvcID),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	if len(apiImpls.ListAPIImplementationsResponse.Data) < 1 {
		_, err = apiImplementationConfigService.CreateAPIImplementation(ctx, api.ID, components.APIImplementation{
			Service: components.APIImplementationService{
				ControlPlaneID: cpID,
				ID:             gwSvcID,
			},
		})
		if err != nil {
			return "", err
		}
	}

	return api.ID, nil
}

func toPortalLabels(labels map[string]string) map[string]*string {
	o := map[string]*string{}
	for k, v := range labels {
		o[k] = kk.String(v)
	}
	return o
}
