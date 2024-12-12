package manifest

// Orchestrator represents the root configuration structure
type Orchestrator struct {
	Platform      Platform                `json:"platform,omitempty" yaml:"platform,omitempty"`
	Teams         map[string]Team         `json:"teams,omitempty" yaml:"teams,omitempty"`
	Organizations map[string]Organization `json:"organizations,omitempty" yaml:"organizations,omitempty"`
}

// Team represents a team's configuration
type Team struct {
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Users       []string           `json:"users,omitempty" yaml:"users,omitempty"`
	Services    map[string]Service `json:"services,omitempty" yaml:"services,omitempty"`
}

// Service represents a service's configuration
type Service struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Git         Git    `json:"git,omitempty" yaml:"git,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	SpecPath    string `json:"spec-path,omitempty" yaml:"spec-path,omitempty"`
}

// Organization represents an organization's configuration
type Organization struct {
	AccessToken  AccessToken            `json:"access-token" yaml:"access-token"`
	Environments map[string]Environment `json:"environments,omitempty" yaml:"environments,omitempty"`
}

// Environment represents an environment's configuration
type Environment struct {
	Type   string                      `json:"type" yaml:"type"`
	Region string                      `json:"region" yaml:"region"`
	Teams  map[string]TeamEnvironments `json:"teams,omitempty" yaml:"teams,omitempty"`
}

// TeamEnvironments represents a team's configuration within an environment
type TeamEnvironments struct {
	Services map[string]TeamEnvironmentServices `json:"services,omitempty" yaml:"services,omitempty"`
}

// TeamEnvironmentServices represents a service's configuration within an environment
type TeamEnvironmentServices struct {
	Branch string `json:"branch,omitempty" yaml:"branch,omitempty"`
}

// AccessToken represents the configuration for organization access tokens
type AccessToken struct {
	Type  string `json:"type" yaml:"type"`
	Value string `json:"value" yaml:"value"`
}

// Platform represents the platform team configuration
type Platform struct {
	Git Git `json:"git,omitempty" yaml:"git,omitempty"`
}

// Git represents git repository configuration
type Git struct {
	Remote string  `json:"remote,omitempty" yaml:"remote,omitempty"`
	Auth   GitAuth `json:"auth,omitempty" yaml:"auth,omitempty"`
}

// GitAuth represents git authentication configuration
type GitAuth struct {
	Type string     `json:"type,omitempty" yaml:"type,omitempty"`
	SSH  GitAuthSSH `json:"ssh,omitempty" yaml:"ssh,omitempty"`
}

// GitAuthSSH represents SSH key configuration
type GitAuthSSH struct {
	Key GitAuthSSHKey `json:"key,omitempty" yaml:"key,omitempty"`
}

// GitAuthSSHKey represents a generic key configuration that can be loaded from different sources
type GitAuthSSHKey struct {
	Type  string `json:"type,omitempty" yaml:"type,omitempty"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}
