// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package operations

import (
	"github.com/Kong/konnect-orchestrator/internal/reports/components"
	"net/http"
)

type GetReportRequest struct {
	// The report's ID
	ReportID string `pathParam:"style=simple,explode=false,name=reportId"`
}

func (o *GetReportRequest) GetReportID() string {
	if o == nil {
		return ""
	}
	return o.ReportID
}

type GetReportResponse struct {
	// HTTP response content type for this operation
	ContentType string
	// HTTP response status code for this operation
	StatusCode int
	// Raw HTTP response; suitable for custom response parsing
	RawResponse *http.Response
	// A response including a single report.
	Report *components.Report
}

func (o *GetReportResponse) GetContentType() string {
	if o == nil {
		return ""
	}
	return o.ContentType
}

func (o *GetReportResponse) GetStatusCode() int {
	if o == nil {
		return 0
	}
	return o.StatusCode
}

func (o *GetReportResponse) GetRawResponse() *http.Response {
	if o == nil {
		return nil
	}
	return o.RawResponse
}

func (o *GetReportResponse) GetReport() *components.Report {
	if o == nil {
		return nil
	}
	return o.Report
}