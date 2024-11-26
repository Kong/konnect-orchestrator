package internal

import "github.com/Kong/konnect-orchestrator/internal/organization/team"

// OrchestratorManifest represents the root configuration structure
type OrchestratorManifest struct {
	Teams         map[string]TeamManifest         `json:"teams,omitempty" yaml:"teams,omitempty"`
	Organizations map[string]OrganizationManifest `json:"organizations,omitempty" yaml:"organizations,omitempty"`
}

// TeamManifest represents a team's configuration
type TeamManifest struct {
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Users       []string                   `json:"users,omitempty" yaml:"users,omitempty"`
	Services    map[string]ServiceManifest `json:"services,omitempty" yaml:"services,omitempty"`
}

// ServiceManifest represents a service's configuration
type ServiceManifest struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	VCS         string `json:"vcs,omitempty" yaml:"vcs,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	SpecPath    string `json:"spec-path,omitempty" yaml:"spec-path,omitempty"`
	KongPath    string `json:"kong-path,omitempty" yaml:"kong-path,omitempty"`
}

// OrganizationManifest represents an organization's configuration
type OrganizationManifest struct {
	AccessToken  AccessToken                    `json:"access-token" yaml:"access-token"`
	Environments map[string]EnvironmentManifest `json:"environments,omitempty" yaml:"environments,omitempty"`
}

// EnvironmentManifest represents an environment's configuration
type EnvironmentManifest struct {
	Type  string                             `json:"type" yaml:"type"`
	Teams map[string]EnvironmentTeamManifest `json:"teams,omitempty" yaml:"teams,omitempty"`
}

// EnvironmentTeamManifest represents a team's configuration within an environment
type EnvironmentTeamManifest struct {
	Services []string `json:"services,omitempty" yaml:"services,omitempty"`
}

// AccessToken represents the configuration for organization access tokens
type AccessToken struct {
	Type  string `json:"type" yaml:"type"`
	Value string `json:"value" yaml:"value"`
}

// ParseTeams converts the TeamManifest map into a slice of Team objects
func ParseTeams(teams map[string]TeamManifest) []team.Team {
	result := make([]team.Team, 0, len(teams))
	for name, config := range teams {
		services := make(map[string]team.ServiceConfig)
		for svcKey, svc := range config.Services {
			services[svcKey] = team.ServiceConfig{
				Name:        svc.Name,
				VCS:         svc.VCS,
				Description: svc.Description,
				SpecPath:    svc.SpecPath,
				KongPath:    svc.KongPath,
			}
		}

		result = append(result, team.Team{
			Name:        name,
			Description: config.Description,
			Users:       config.Users,
			Services:    services,
		})
	}
	return result
}
