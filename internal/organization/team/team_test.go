package team

import (
	"context"
	"fmt"
	"testing"

	kk "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApplyTeam(t *testing.T) {
	tests := []struct {
		name    string
		config  Team
		setup   func(*MockTeamService)
		wantErr bool
	}{
		{
			name: "creates new team when it doesn't exist",
			config: Team{
				Name:        "new-team",
				Description: "A new team",
			},
			setup: func(m *MockTeamService) {
				// Setup ListTeams to return empty result
				m.On("ListTeams", mock.Anything, mock.MatchedBy(func(req operations.ListTeamsRequest) bool {
					return req.Filter != nil &&
						req.Filter.Name != nil &&
						*req.Filter.Name.StringFieldEqualsFilter.Str == "new-team"
				})).Return(&operations.ListTeamsResponse{
					TeamCollection: &components.TeamCollection{
						Data: []components.Team{},
					},
				}, nil)

				// Setup CreateTeam expectation
				m.On("CreateTeam", mock.Anything, &components.CreateTeam{
					Name:        "new-team",
					Description: kk.String("A new team"),
				}).Return(&operations.CreateTeamResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "updates existing team",
			config: Team{
				Name:        "existing-team",
				Description: "Updated description",
			},
			setup: func(m *MockTeamService) {
				// Setup ListTeams to return existing team
				m.On("ListTeams", mock.Anything, mock.MatchedBy(func(req operations.ListTeamsRequest) bool {
					return req.Filter != nil &&
						req.Filter.Name != nil &&
						*req.Filter.Name.StringFieldEqualsFilter.Str == "existing-team"
				})).Return(&operations.ListTeamsResponse{
					TeamCollection: &components.TeamCollection{
						Data: []components.Team{
							{
								ID:          kk.String("team-123"),
								Name:        kk.String("existing-team"),
								Description: kk.String("Old description"),
							},
						},
					},
				}, nil)

				// Setup UpdateTeam expectation
				m.On("UpdateTeam", mock.Anything, "team-123", &components.UpdateTeam{
					Name:        kk.String("existing-team"),
					Description: kk.String("Updated description"),
				}).Return(&operations.UpdateTeamResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "handles list error",
			config: Team{
				Name:        "error-team",
				Description: "Error team",
			},
			setup: func(m *MockTeamService) {
				m.On("ListTeams", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("list error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockTeamService{}
			tt.setup(mockSvc)

			err := ApplyTeam(context.Background(), mockSvc, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockSvc)
		})
	}
}
