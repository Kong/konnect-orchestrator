package auth

import (
	"context"
	"errors"
	"fmt"

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

type IdentitProviderTeamMappingService interface {
	UpdateIdpTeamMappings(ctx context.Context,
		opts ...operations.Option) (*operations.UpdateIdpTeamMappingsResponse, error)
}

func ApplyAuthSettings(
	ctx context.Context,
	idpSvc IdentityProviderConfigService,
	authSvc AuthenticationSettingsService,
	authSettings manifest.Authorization) error {

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
		samlResp, err := idpSvc.CreateIdentityProvider(ctx, components.CreateIdentityProvider{
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
		samlProviderID = *samlResp.IdentityProvider.ID
	}
	// ***********************************************************************************************

	// ***********************************************************************************************
	// Finally, update the authentication settings to enable the desired combination of auth methods.

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

	return nil
}
