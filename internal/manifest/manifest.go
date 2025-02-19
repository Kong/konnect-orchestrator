package manifest

type Orchestrator struct {
	Platform      *Platform                `json:"platform,omitempty" yaml:"platform,omitempty"`
	Teams         map[string]*Team         `json:"teams,omitempty" yaml:"teams,omitempty"`
	Organizations map[string]*Organization `json:"organizations,omitempty" yaml:"organizations,omitempty"`
}

type Team struct {
	Description *string             `json:"description,omitempty" yaml:"description,omitempty"`
	Users       []string            `json:"users,omitempty" yaml:"users,omitempty"`
	Services    map[string]*Service `json:"services,omitempty" yaml:"services,omitempty"`
}

type Service struct {
	Name        *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Git         *GitConfig `json:"git,omitempty" yaml:"git,omitempty"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	SpecPath    *string    `json:"spec-path,omitempty" yaml:"spec-path,omitempty"`
}

type Organization struct {
	AccessToken   Secret                  `json:"access-token" yaml:"access-token"`
	Environments  map[string]*Environment `json:"environments,omitempty" yaml:"environments,omitempty"`
	Authorization *Authorization          `json:"authorization,omitempty" yaml:"authorization,omitempty"`
}

type Environment struct {
	Type   string                      `json:"type" yaml:"type"`
	Region string                      `json:"region" yaml:"region"`
	Teams  map[string]*TeamEnvironment `json:"teams,omitempty" yaml:"teams,omitempty"`
}

type TeamEnvironment struct {
	ControlPlaneName *string                        `json:"control-plane-name,omitempty" yaml:"control-plane-name,omitempty"`
	Services         map[string]*EnvironmentService `json:"services,omitempty" yaml:"services,omitempty"`
}

type EnvironmentService struct {
	Branch *string `json:"branch,omitempty" yaml:"branch,omitempty"`
}

type Secret struct {
	// Type is the storage type of secret, e.g. file, env, literal
	Type string `json:"type" yaml:"type"`
	// Value is the value of the secret, e.g. the file path, env var name, or literal value
	Value string `json:"value" yaml:"value"`
}

// Platform represents the platform team configuration
type Platform struct {
	Git *GitConfig `json:"git,omitempty" yaml:"git,omitempty"`
}

// GitConfig represents git repository configuration
type GitConfig struct {
	Remote *string       `json:"remote,omitempty" yaml:"remote,omitempty"`
	Author *Author       `json:"author,omitempty" yaml:"author,omitempty"`
	Auth   *AuthConfig   `json:"auth,omitempty" yaml:"auth,omitempty"`
	GitHub *GitHubConfig `json:"github,omitempty" yaml:"github,omitempty"`
}

type GitHubConfig struct {
	Token *Secret `json:"token,omitempty" yaml:"token,omitempty"`
}

// AuthConfig represents git authentication configuration
type AuthConfig struct {
	Type  *string    `json:"type,omitempty" yaml:"type,omitempty"`
	SSH   *SSHConfig `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	Token *Secret    `json:"token,omitempty" yaml:"token,omitempty"`
}

// SSHConfig represents SSH key configuration
type SSHConfig struct {
	Key *Secret `json:"key,omitempty" yaml:"key,omitempty"`
}

// Author represents git commit author configuration
type Author struct {
	Name  *string `json:"name,omitempty" yaml:"name,omitempty"`
	Email *string `json:"email,omitempty" yaml:"email,omitempty"`
}

// Authorization represents the organization's authentication configuration
type Authorization struct {
	BuiltIn      *BuiltInAuth `json:"built-in,omitempty" yaml:"built-in,omitempty"`
	OIDC         *OIDCAuth    `json:"oidc,omitempty" yaml:"oidc,omitempty"`
	SAML         *SAMLAuth    `json:"saml,omitempty" yaml:"saml,omitempty"`
	TeamMappings TeamMappings `json:"team-mappings" yaml:"team-mappings"`
}

// BuiltInAuth represents built-in authentication configuration
type BuiltInAuth struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// OIDCAuth represents OIDC authentication configuration
type OIDCAuth struct {
	Enabled       bool              `json:"enabled" yaml:"enabled"`
	LoginPath     string            `json:"login-path" yaml:"login-path"`
	Issuer        string            `json:"issuer" yaml:"issuer"`
	ClientID      string            `json:"client-id" yaml:"client-id"`
	ClientSecret  Secret            `json:"client-secret" yaml:"client-secret"`
	ClaimMappings map[string]string `json:"claim-mappings" yaml:"claim-mappings"`
	Scopes        []string          `json:"scopes" yaml:"scopes"`
}

// SAMLAuth represents SAML authentication configuration
type SAMLAuth struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	LoginPath      string `json:"login-path" yaml:"login-path"`
	IDPMetadataURL string `json:"idp-metadata-url" yaml:"idp-metadata-url"`
}

// TeamMappings represents team mapping configuration for SAML
type TeamMappings struct {
	BuiltIn BuiltInTeamMapping `json:"built-in" yaml:"built-in"`
	IDP     IDPTeamMapping     `json:"idp" yaml:"idp"`
}

// BuiltInTeamMapping represents built-in team mapping configuration
type BuiltInTeamMapping struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// IDPTeamMapping represents IDP team mapping configuration
type IDPTeamMapping struct {
	Enabled  bool                `json:"enabled" yaml:"enabled"`
	Mappings map[string][]string `json:"mappings" yaml:"mappings"`
}
