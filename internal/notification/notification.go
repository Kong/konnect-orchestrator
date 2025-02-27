package notification

import (
	"context"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/sdk-konnect-go-internal/models/components"
	"github.com/Kong/sdk-konnect-go-internal/models/operations"
)

type NotificationsService interface {
	CreateEventSubscription(
		ctx context.Context,
		eventID string,
		eventSubscription *components.EventSubscription,
		opts ...operations.Option) (*operations.CreateEventSubscriptionResponse, error)
	ListUserConfigurations(
		ctx context.Context,
		filter *components.ConfigurationFilterParameters,
		opts ...operations.Option) (*operations.ListUserConfigurationsResponse, error)
	ListEventSubscriptions(
		ctx context.Context,
		eventID string,
		opts ...operations.Option) (*operations.ListEventSubscriptionsResponse, error)
	UpdateEventSubscription(
		ctx context.Context,
		request operations.UpdateEventSubscriptionRequest,
		opts ...operations.Option) (*operations.UpdateEventSubscriptionResponse, error)
	DeleteEventSubscription(
		ctx context.Context,
		eventID string,
		subscriptionID string,
		opts ...operations.Option) (*operations.DeleteEventSubscriptionResponse, error)
}

// If you change the name of a portal, a new one will be created an the old one remains
func ApplyNotificationsConfig(
	ctx context.Context,
	notificationsService NotificationsService,
	notificationsConfig *manifest.Notifications) error {

	var (
		emailEnabled = true
		inAppEnabled = true
	)

	if notificationsConfig != nil {
		if notificationsConfig.InApp != nil {
			inAppEnabled = *notificationsConfig.InApp
		}
		if notificationsConfig.Email != nil {
			emailEnabled = *notificationsConfig.Email
		}
	}

	userConfigurations, err := notificationsService.ListUserConfigurations(
		ctx,
		nil,
	)
	if err != nil {
		return err
	}

	for _, configuration := range userConfigurations.GetUserConfigurationListResponse().GetData() {
		eventSubscriptions, err := notificationsService.ListEventSubscriptions(ctx, configuration.EventID)
		if err != nil {
			return err
		}

		if emailEnabled && inAppEnabled {
			// We desire the default state, delete all existing subscriptions (could be 0) to return there
			for _, eventSub := range eventSubscriptions.GetEventSubscriptionListResponse().GetData() {
				_, err := notificationsService.DeleteEventSubscription(ctx, configuration.EventID, eventSub.ID)
				if err != nil {
					return err
				}
			}
		} else {
			if len(eventSubscriptions.GetEventSubscriptionListResponse().GetData()) > 0 {
				// If we have subscriptions, check them against the desired state and update as necessary
				for _, eventSub := range eventSubscriptions.GetEventSubscriptionListResponse().GetData() {
					if !desiredStateIsMatching(eventSub.Channels, emailEnabled, inAppEnabled) {
						_, err := notificationsService.UpdateEventSubscription(ctx, operations.UpdateEventSubscriptionRequest{
							EventID:        configuration.EventID,
							SubscriptionID: eventSub.ID,
							EventSubscription: &components.EventSubscription{
								Regions:  eventSub.Regions,
								Entities: eventSub.Entities,
								Enabled:  eventSub.Enabled,
								Channels: []components.NotificationChannel{
									{
										Type:    components.NotificationChannelTypeEmail,
										Enabled: emailEnabled,
									},
									{
										Type:    components.NotificationChannelTypeInApp,
										Enabled: inAppEnabled,
									},
								},
							},
						})
						if err != nil {
							return err
						}
					}
				}
			} else {
				// If we _don't_ have subscriptions (and we don't desire the default state), create a new subscription
				_, err := notificationsService.CreateEventSubscription(ctx, configuration.EventID, (&components.EventSubscription{
					Regions:  []components.NotificationRegion{"*"},
					Entities: []string{"*"},
					Enabled:  true,
					Channels: []components.NotificationChannel{
						{
							Type:    components.NotificationChannelTypeEmail,
							Enabled: emailEnabled,
						},
						{
							Type:    components.NotificationChannelTypeInApp,
							Enabled: inAppEnabled,
						},
					},
				}))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func desiredStateIsMatching(channels []components.NotificationChannel, emailEnabled, inAppEnabled bool) bool {
	var desiredIsMatching = true

	for _, c := range channels {
		if c.Type == components.NotificationChannelTypeEmail && c.Enabled != emailEnabled {
			desiredIsMatching = false
		}
		if c.Type == components.NotificationChannelTypeInApp && c.Enabled != inAppEnabled {
			desiredIsMatching = false
		}
	}

	return desiredIsMatching
}
