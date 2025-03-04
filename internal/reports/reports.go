package reports

import (
	"context"

	"github.com/Kong/konnect-orchestrator/internal/reports/components"
	"github.com/Kong/konnect-orchestrator/internal/reports/operations"
	kk "github.com/Kong/sdk-konnect-go"
)

type CustomReportsService interface {
	GetReports(ctx context.Context,
		pageSize *int64,
		pageNumber *int64,
		opts ...operations.Option) (*operations.GetReportsResponse, error)
	CreateReport(ctx context.Context,
		request *components.ReportInput,
		opts ...operations.Option) (*operations.CreateReportResponse, error)
}

// If you change the name of a report, a new one will be created and the old one remains
func ApplyReports(
	ctx context.Context,
	reportsService CustomReportsService,
) error {
	reports, err := reportsService.GetReports(ctx, kk.Int64(50), kk.Int64(1))
	if err != nil {
		return err
	}

	for _, defaultReport := range defaultReports {
		if !reportExists(defaultReport.Name, reports.GetReportCollection().GetData()) {
			_, err = reportsService.CreateReport(ctx, defaultReport)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func reportExists(reportName *string, reports []components.ReportCollectionReport) bool {
	for _, report := range reports {
		if *report.Name == *reportName {
			return true
		}
	}
	return false
}
