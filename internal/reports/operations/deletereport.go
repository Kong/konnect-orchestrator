// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package operations

import (
	"net/http"
)

type DeleteReportRequest struct {
	// The report's ID
	ReportID string `pathParam:"style=simple,explode=false,name=reportId"`
}

func (o *DeleteReportRequest) GetReportID() string {
	if o == nil {
		return ""
	}
	return o.ReportID
}

type DeleteReportResponse struct {
	// HTTP response content type for this operation
	ContentType string
	// HTTP response status code for this operation
	StatusCode int
	// Raw HTTP response; suitable for custom response parsing
	RawResponse *http.Response
}

func (o *DeleteReportResponse) GetContentType() string {
	if o == nil {
		return ""
	}
	return o.ContentType
}

func (o *DeleteReportResponse) GetStatusCode() int {
	if o == nil {
		return 0
	}
	return o.StatusCode
}

func (o *DeleteReportResponse) GetRawResponse() *http.Response {
	if o == nil {
		return nil
	}
	return o.RawResponse
}