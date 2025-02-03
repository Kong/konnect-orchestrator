package auth

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	kk "github.com/Kong/sdk-konnect-go-internal"
	"github.com/Kong/sdk-konnect-go-internal/models/components"
	"github.com/Kong/sdk-konnect-go-internal/models/operations"
)

type AuthenticationSettingsService interface {
	GetAuthenticationSettings(ctx context.Context,
		opts ...operations.Option) (*operations.GetAuthenticationSettingsResponse, error)
	UpdateAuthenticationSettings(ctx context.Context,
		request *components.UpdateAuthenticationSettings,
		opts ...operations.Option) (*operations.UpdateAuthenticationSettingsResponse, error)
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

func applyAuthSettings(
	ctx context.Context,
	idpSvc IdentityProviderConfigService,
	authSvc AuthenticationSettingsService,
	authSettings manifest.Authorization) error {

	// Auth settings have to be configured in a certain order.
	// Use the /identity-providers endpoint to manage OIDC and SAML configurations.
	// 1. If we have an OIDC configuration, search for an OIDC provider config
	// 2. If we have an OIDC provider config, patch via the found ID
	// 3. If we do not have an OIDC provider config, POST a new one
	// 4. If we have a SAML configuration, search for a SAML provider config
	// 5. If we have a SAML provider config, patch vis the found ID
	// 6. If we do not have a SAML provider config, POST a new one
	// 7. PATCH the authentication settings via the /authentication-settings endpoint
	//    * OIDC and SAMl enabled fields are XOR
	// 8. PATCH identity-providers enabled flag for specific the specific provider
	// 9. Next mappings...

	idps, err := idpSvc.GetIdentityProviders(ctx, &operations.Filter{})
	if err != nil {
		return fmt.Errorf("failed to get identity providers: %w", err)
	}

	// Search for both the OIDC and SAML Identity Providers
	var oidcProvider, samlProvider *components.IdentityProvider

	if idps.IdentityProviders != nil {
		for _, idp := range idps.IdentityProviders {
			if *idp.Type == components.IdentityProviderTypeOidc {
				oidcProvider = &idp
			} else if *idp.Type == components.IdentityProviderTypeOidc {
				samlProvider = &idp
			}
		}
	}

	if authSettings.OIDC.Enabled {
		if oidcProvider == nil {
			secret, err := util.ResolveSecretValue(authSettings.OIDC.ClientSecret)
			if err != nil {
				return fmt.Errorf("failed to resolve OIDC client secret: %w", err)
			}

			// create OIDC provider
			resp, err := idpSvc.CreateIdentityProvider(ctx, components.CreateIdentityProvider{
				Type:      components.IdentityProviderTypeOidc.ToPointer(),
				LoginPath: kk.String(authSettings.OIDC.LoginPath),
				Config: &components.CreateIdentityProviderConfig{
					OIDCIdentityProviderConfig: &components.OIDCIdentityProviderConfig{
						IssuerURL:     authSettings.OIDC.Issuer,
						ClientID:      authSettings.OIDC.ClientID,
						ClientSecret:  kk.String(secret),
						Scopes:        authSettings.OIDC.Scopes,
						ClaimMappings: nil,
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create OIDC provider: %w", err)
			}
			oidcProvider = resp.IdentityProvider
		} else {
			// update OIDC provider
			resp, err := idpSvc.UpdateIdentityProvider(ctx, *oidcProvider.ID, components.UpdateIdentityProvider{})
			if err != nil {
				return fmt.Errorf("failed to update OIDC provider: %w", err)
			}
			oidcProvider = resp.IdentityProvider
		}
	} else {
		if oidcProvider != nil {
			// disable OIDC provider
		}
	}

	if authSettings.SAML.Enabled {
		if samlProvider == nil {
		} else {
		}
	} else {
		if samlProvider != nil {
		}
	}

	_, err = authSvc.UpdateAuthenticationSettings(ctx,
		&components.UpdateAuthenticationSettings{
			BasicAuthEnabled: kk.Bool(authSettings.BuiltIn.Enabled),
			OidcAuthEnabled:  kk.Bool(authSettings.OIDC.Enabled),
			SamlAuthEnabled:  kk.Bool(authSettings.SAML.Enabled),
		})

	if err != nil {
		return fmt.Errorf("failed to update authentication settings: %w", err)
	}

	return nil
}
