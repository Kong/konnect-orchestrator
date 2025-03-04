package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/Kong/sdk-konnect-go/models/sdkerrors"
)

// From the Konnect SDK:
//   * AuthenticationSettings is the response from the /authentication-settings endpoint

type AuthenticationSettingsService interface {
	GetAuthenticationSettings(ctx context.Context,
		opts ...operations.Option) (*operations.GetAuthenticationSettingsResponse, error)
	UpdateAuthenticationSettings(ctx context.Context,
		request *components.UpdateAuthenticationSettings,
		opts ...operations.Option) (*operations.UpdateAuthenticationSettingsResponse, error)
	UpdateIdpConfiguration(ctx context.Context,
		request *components.UpdateIDPConfiguration,
		opts ...operations.Option) (*operations.UpdateIdpConfigurationResponse, error)
}

type IdentityProviderConfigService interface {
	GetIdentityProviders(ctx context.Context,
		filter *operations.Filter,
		opts ...operations.Option) (*operations.GetIdentityProvidersResponse, error)
	CreateIdentityProvider(ctx context.Context,
		request components.CreateIdentityProvider,
		opts ...operations.Option) (*operations.CreateIdentityProviderResponse, error)
	UpdateIdentityProvider(ctx context.Context,
		id string,
		updateIdentityProvider components.UpdateIdentityProvider,
		opts ...operations.Option) (*operations.UpdateIdentityProviderResponse, error)
}

type TeamsService interface {
	ListTeams(ctx context.Context,
		request operations.ListTeamsRequest,
		opts ...operations.Option) (*operations.ListTeamsResponse, error)
}

type IdentityProviderTeamMappingService interface {
	PatchTeamGroupMappings(ctx context.Context,
		request *components.PatchTeamGroupMappings,
		opts ...operations.Option) (*operations.PatchTeamGroupMappingsResponse, error)
}

// getTitleCase converts a slug (with hyphens) into Title Case.
func getTitleCase(str string) string {
	// Replace hyphens with spaces
	str = strings.ReplaceAll(str, "-", " ")

	// Regex to match words
	re := regexp.MustCompile(`\b\w\S*\b`)

	// Capitalize each word
	return re.ReplaceAllStringFunc(str, func(txt string) string {
		return strings.ToUpper(txt[:1]) + strings.ToLower(txt[1:])
	})
}

func ApplyAuthSettings(
	ctx context.Context,
	idpSvc IdentityProviderConfigService,
	authSvc AuthenticationSettingsService,
	teamSvc TeamsService,
	teamMappingSvc IdentityProviderTeamMappingService,
	authSettings manifest.Authorization,
) error {
	oidcProviderID := ""
	samlProviderID := ""

	// ***********************************************************************************************
	// First, apply the OIDC configuration which uses the 'legacy' /identity-provider API for now
	secret, err := util.ResolveSecretValue(authSettings.OIDC.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to resolve OIDC client secret: %w", err)
	}
	_, err = authSvc.UpdateIdpConfiguration(ctx, &components.UpdateIDPConfiguration{
		Issuer:       kk.String(authSettings.OIDC.Issuer),
		LoginPath:    kk.String(authSettings.OIDC.LoginPath),
		ClientID:     kk.String(authSettings.OIDC.ClientID),
		ClientSecret: kk.String(secret),
		Scopes:       authSettings.OIDC.Scopes,
		ClaimMappings: &components.UpdateIDPConfigurationClaimMappings{
			Email:  kk.String(authSettings.OIDC.ClaimMappings["email"]),
			Name:   kk.String(authSettings.OIDC.ClaimMappings["name"]),
			Groups: kk.String(authSettings.OIDC.ClaimMappings["groups"]),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	oidcQueryResponse, err := idpSvc.GetIdentityProviders(ctx, &operations.Filter{
		Type: &components.StringFieldEqualsFilter{Str: kk.String("oidc")},
	})
	if err != nil {
		return fmt.Errorf("failed to get saml identity providers: %w", err)
	}
	if len(oidcQueryResponse.IdentityProviders) > 0 {
		oidcProviderID = *oidcQueryResponse.IdentityProviders[0].ID
	}
	// ***********************************************************************************************

	// ***********************************************************************************************
	// Now apply the SAML configuration.
	// Now query for an existing SAML provider specifically, since we'll manage it
	//	differently then OIDC due to Konnect API limitations
	samlQueryResponse, err := idpSvc.GetIdentityProviders(ctx, &operations.Filter{
		Type: &components.StringFieldEqualsFilter{Str: kk.String("saml")},
	})
	if err != nil {
		return fmt.Errorf("failed to get saml identity providers: %w", err)
	}
	if len(samlQueryResponse.IdentityProviders) > 0 {
		// update saml provider
		samlProvider := samlQueryResponse.IdentityProviders[0]
		samlProviderID = *samlProvider.ID
		update := components.UpdateIdentityProvider{}
		doUpdate := false
		if authSettings.SAML.LoginPath != *samlProvider.LoginPath {
			update.LoginPath = kk.String(authSettings.SAML.LoginPath)
			doUpdate = true
		}
		if authSettings.SAML.IDPMetadataURL != *samlProvider.Config.SAMLIdentityProviderConfig.IdpMetadataURL {
			doUpdate = true
			update.Config = &components.UpdateIdentityProviderConfig{
				SAMLIdentityProviderConfigInput: &components.SAMLIdentityProviderConfigInput{
					IdpMetadataURL: kk.String(authSettings.SAML.IDPMetadataURL),
				},
			}
		}
		if doUpdate {
			_, err := idpSvc.UpdateIdentityProvider(ctx,
				*samlQueryResponse.IdentityProviders[0].ID,
				update)
			if err != nil {
				return fmt.Errorf("failed to update saml provider: %w", err)
			}
		}
	} else {
		// create new saml provider
		_, err := idpSvc.CreateIdentityProvider(ctx, components.CreateIdentityProvider{
			Type:      components.IdentityProviderTypeSaml.ToPointer(),
			LoginPath: kk.String(authSettings.SAML.LoginPath),
			Config: &components.CreateIdentityProviderConfig{
				SAMLIdentityProviderConfigInput: &components.SAMLIdentityProviderConfigInput{
					IdpMetadataURL: kk.String(authSettings.SAML.IDPMetadataURL),
				},
			},
		})
		if err != nil {
			var sdkErr *sdkerrors.SDKError
			// Right now the create endpoint will return a 200 instead of a 201 as specified causing the
			//  sdk to barf. This is to ignore that one case.
			if errors.As(err, &sdkErr) && sdkErr.StatusCode != 200 {
				return fmt.Errorf("failed to create saml provider: %w", err)
			}
		}
		samlQueryResponse, err := idpSvc.GetIdentityProviders(ctx, &operations.Filter{
			Type: &components.StringFieldEqualsFilter{Str: kk.String("saml")},
		})
		if err != nil {
			return fmt.Errorf("failed to get saml identity providers: %w", err)
		}
		if len(samlQueryResponse.IdentityProviders) > 0 {
			samlProviderID = *samlQueryResponse.IdentityProviders[0].ID
		}
	}
	// ***********************************************************************************************

	// ***********************************************************************************************
	// Then, update the authentication settings and idps to enable the desired combination of auth methods.

	_, err = authSvc.UpdateAuthenticationSettings(ctx, &components.UpdateAuthenticationSettings{
		BasicAuthEnabled: kk.Bool(authSettings.BuiltIn.Enabled),
		OidcAuthEnabled:  kk.Bool(false),
		SamlAuthEnabled:  kk.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to update authentication settings: %w", err)
	}

	_, err = idpSvc.UpdateIdentityProvider(ctx, oidcProviderID, components.UpdateIdentityProvider{
		Enabled: kk.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to disable OIDC provider: %w", err)
	}
	_, err = idpSvc.UpdateIdentityProvider(ctx, samlProviderID, components.UpdateIdentityProvider{
		Enabled: kk.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to disable SAML provider: %w", err)
	}

	_, err = authSvc.UpdateAuthenticationSettings(ctx, &components.UpdateAuthenticationSettings{
		OidcAuthEnabled: kk.Bool(authSettings.OIDC.Enabled),
	})
	if err != nil {
		return fmt.Errorf("failed to update authentication settings: %w", err)
	}
	_, err = idpSvc.UpdateIdentityProvider(ctx, oidcProviderID, components.UpdateIdentityProvider{
		Enabled: kk.Bool(authSettings.OIDC.Enabled),
	})
	if err != nil {
		return fmt.Errorf("failed to enable OIDC provider: %w", err)
	}

	_, err = authSvc.UpdateAuthenticationSettings(ctx, &components.UpdateAuthenticationSettings{
		SamlAuthEnabled: kk.Bool(authSettings.SAML.Enabled),
	})
	if err != nil {
		return fmt.Errorf("failed to update authentication settings: %w", err)
	}
	_, err = idpSvc.UpdateIdentityProvider(ctx, samlProviderID, components.UpdateIdentityProvider{
		Enabled: kk.Bool(authSettings.SAML.Enabled),
	})
	if err != nil {
		return fmt.Errorf("failed to enable SAML provider: %w", err)
	}
	// ***********************************************************************************************

	// ***********************************************************************************************
	// Now setup the team mappings
	_, err = authSvc.UpdateAuthenticationSettings(ctx, &components.UpdateAuthenticationSettings{
		KonnectMappingEnabled: kk.Bool(authSettings.TeamMappings.BuiltIn.Enabled),
		IdpMappingEnabled:     kk.Bool(authSettings.TeamMappings.IDP.Enabled),
	})
	if err != nil {
		return fmt.Errorf("failed to update authentication settings: %w", err)
	}

	teamSvcResp, err := teamSvc.ListTeams(ctx, operations.ListTeamsRequest{
		PageSize: kk.Int64(100),
	})
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	teams := teamSvcResp.TeamCollection.Data
	var mappings []components.Data

	for cfgTeamName, groups := range authSettings.TeamMappings.IDP.Mappings {
		teamID := ""
		for _, team := range teams {
			konnectTeamName := *team.Name
			if *team.SystemTeam {
				konnectTeamName = getTitleCase(*team.Name)
			}
			if cfgTeamName == konnectTeamName {
				teamID = *team.ID
				break
			}
		}
		mappings = append(mappings, components.Data{
			TeamID: kk.String(teamID),
			Groups: groups,
		})
	}

	_, err = teamMappingSvc.PatchTeamGroupMappings(ctx, &components.PatchTeamGroupMappings{
		Data: mappings,
	})
	if err != nil {
		return fmt.Errorf("failed to update team group mappings: %w", err)
	}

	return nil
}
