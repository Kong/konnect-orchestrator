package internal

import "github.com/Kong/konnect-orchestrator/internal/organization/team"

// OrchestratorManifest represents the root configuration structure
type OrchestratorManifest struct {
	Teams         map[string]TeamManifest         `json:"teams,omitempty" yaml:"teams,omitempty"`
	Services      map[string]ServiceManifest      `json:"services,omitempty" yaml:"services,omitempty"`
	Organizations map[string]OrganizationManifest `json:"organizations,omitempty" yaml:"organizations,omitempty"`
}

// OrganizationManifest represents an organization's configuration
type OrganizationManifest struct {
	Name        string      `json:"name,omitempty" yaml:"name,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Teams       []string    `json:"teams,omitempty" yaml:"teams,omitempty"`
	AccessToken AccessToken `json:"access-token" yaml:"access-token"`
}

// AccessToken represents the configuration for organization access tokens
type AccessToken struct {
	Type  string `json:"type" yaml:"type"`
	Value string `json:"value" yaml:"value"`
}

// TeamManifest represents a team's configuration
type TeamManifest struct {
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Users       []string `json:"users,omitempty" yaml:"users,omitempty"`
}

// ServiceManifest represents a service's configuration
type ServiceManifest struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	VCS         string `json:"vcs,omitempty" yaml:"vcs,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Team        string `json:"team,omitempty" yaml:"team,omitempty"`
	SpecPath    string `json:"spec-path,omitempty" yaml:"spec-path,omitempty"`
	KongPath    string `json:"kong-path,omitempty" yaml:"kong-path,omitempty"`
}

// ParseTeams converts the TeamManifest map into a slice of Team objects
func ParseTeams(teams map[string]TeamManifest) []team.Team {
	result := make([]team.Team, 0, len(teams))
	for name, config := range teams {
		result = append(result, team.Team{
			Name:        name,
			Description: config.Description,
			Users:       config.Users,
		})
	}
	return result
}
