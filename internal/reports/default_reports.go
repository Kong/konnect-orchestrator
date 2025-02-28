package reports

import (
	"github.com/Kong/konnect-orchestrator/internal/reports/components"
	kk "github.com/Kong/sdk-konnect-go"
)

var (
	rel30d = components.CreateTimeRangeRelative(components.RelativeTimeRange{
		Type:      components.TypeRelative,
		TimeRange: *components.RelativeTimeRangeTimeRangeThirtyd.ToPointer(),
	})

	qryUsage = components.Query{
		Version:   kk.String("v6"),
		ChartType: components.ChartTypeHorizontalBar.ToPointer(),
		Explore: &components.Explore{
			TimeRange: &rel30d,
			Metrics: []components.Metrics{
				components.MetricsRequestCount,
			},
			Dimensions: []components.Dimensions{
				components.DimensionsAPIProduct,
			},
		},
	}

	qryUsageApp = components.Query{
		Version:   kk.String("v6"),
		ChartType: components.ChartTypeVerticalBar.ToPointer(),
		Explore: &components.Explore{
			TimeRange: &rel30d,
			Metrics: []components.Metrics{
				components.MetricsRequestCount,
			},
			Dimensions: []components.Dimensions{
				components.DimensionsAPIProduct,
				components.DimensionsApplication,
			},
		},
	}

	defaultReports = []*components.ReportInput{
		{
			Name:  kk.String("API Usage (last 30 days)"),
			Query: &qryUsage,
		},
		{
			Name:  kk.String("API Usage by Application (last 30 days)"),
			Query: &qryUsageApp,
		},
	}
)
