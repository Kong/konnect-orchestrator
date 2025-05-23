// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package components

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Kong/konnect-orchestrator/internal/reports/utils"
	"time"
)

// ChartType - Visualization type selected for this report.
type ChartType string

const (
	ChartTypeTimeseriesLine ChartType = "timeseries_line"
	ChartTypeTimeseriesBar  ChartType = "timeseries_bar"
	ChartTypeHorizontalBar  ChartType = "horizontal_bar"
	ChartTypeVerticalBar    ChartType = "vertical_bar"
)

func (e ChartType) ToPointer() *ChartType {
	return &e
}
func (e *ChartType) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "timeseries_line":
		fallthrough
	case "timeseries_bar":
		fallthrough
	case "horizontal_bar":
		fallthrough
	case "vertical_bar":
		*e = ChartType(v)
		return nil
	default:
		return fmt.Errorf("invalid value for ChartType: %v", v)
	}
}

type Metrics string

const (
	MetricsActiveServices         Metrics = "active_services"
	MetricsRequestCount           Metrics = "request_count"
	MetricsRequestPerMinute       Metrics = "request_per_minute"
	MetricsResponseLatencyAverage Metrics = "response_latency_average"
	MetricsResponseLatencyP99     Metrics = "response_latency_p99"
	MetricsResponseLatencyP95     Metrics = "response_latency_p95"
	MetricsResponseLatencyP50     Metrics = "response_latency_p50"
	MetricsUpstreamLatencyP99     Metrics = "upstream_latency_p99"
	MetricsUpstreamLatencyP95     Metrics = "upstream_latency_p95"
	MetricsUpstreamLatencyP50     Metrics = "upstream_latency_p50"
	MetricsUpstreamLatencyAverage Metrics = "upstream_latency_average"
	MetricsKongLatencyP99         Metrics = "kong_latency_p99"
	MetricsKongLatencyP95         Metrics = "kong_latency_p95"
	MetricsKongLatencyP50         Metrics = "kong_latency_p50"
	MetricsKongLatencyAverage     Metrics = "kong_latency_average"
	MetricsResponseSizeP99        Metrics = "response_size_p99"
	MetricsResponseSizeP95        Metrics = "response_size_p95"
	MetricsResponseSizeP50        Metrics = "response_size_p50"
	MetricsResponseSizeAverage    Metrics = "response_size_average"
	MetricsResponseSizeSum        Metrics = "response_size_sum"
	MetricsRequestSizeP99         Metrics = "request_size_p99"
	MetricsRequestSizeP95         Metrics = "request_size_p95"
	MetricsRequestSizeP50         Metrics = "request_size_p50"
	MetricsRequestSizeAverage     Metrics = "request_size_average"
	MetricsRequestSizeSum         Metrics = "request_size_sum"
)

func (e Metrics) ToPointer() *Metrics {
	return &e
}
func (e *Metrics) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "active_services":
		fallthrough
	case "request_count":
		fallthrough
	case "request_per_minute":
		fallthrough
	case "response_latency_average":
		fallthrough
	case "response_latency_p99":
		fallthrough
	case "response_latency_p95":
		fallthrough
	case "response_latency_p50":
		fallthrough
	case "upstream_latency_p99":
		fallthrough
	case "upstream_latency_p95":
		fallthrough
	case "upstream_latency_p50":
		fallthrough
	case "upstream_latency_average":
		fallthrough
	case "kong_latency_p99":
		fallthrough
	case "kong_latency_p95":
		fallthrough
	case "kong_latency_p50":
		fallthrough
	case "kong_latency_average":
		fallthrough
	case "response_size_p99":
		fallthrough
	case "response_size_p95":
		fallthrough
	case "response_size_p50":
		fallthrough
	case "response_size_average":
		fallthrough
	case "response_size_sum":
		fallthrough
	case "request_size_p99":
		fallthrough
	case "request_size_p95":
		fallthrough
	case "request_size_p50":
		fallthrough
	case "request_size_average":
		fallthrough
	case "request_size_sum":
		*e = Metrics(v)
		return nil
	default:
		return fmt.Errorf("invalid value for Metrics: %v", v)
	}
}

type Dimensions string

const (
	DimensionsAPIProduct                Dimensions = "api_product"
	DimensionsAPIProductVersion         Dimensions = "api_product_version"
	DimensionsControlPlane              Dimensions = "control_plane"
	DimensionsControlPlaneGroup         Dimensions = "control_plane_group"
	DimensionsDataPlaneNode             Dimensions = "data_plane_node"
	DimensionsGatewayService            Dimensions = "gateway_service"
	DimensionsPortal                    Dimensions = "portal"
	DimensionsRoute                     Dimensions = "route"
	DimensionsStatusCode                Dimensions = "status_code"
	DimensionsStatusCodeGrouped         Dimensions = "status_code_grouped"
	DimensionsTime                      Dimensions = "time"
	DimensionsApplication               Dimensions = "application"
	DimensionsConsumer                  Dimensions = "consumer"
	DimensionsCountryCode               Dimensions = "country_code"
	DimensionsIsoCode                   Dimensions = "iso_code"
	DimensionsUpstreamStatusCode        Dimensions = "upstream_status_code"
	DimensionsUpstreamStatusCodeGrouped Dimensions = "upstream_status_code_grouped"
	DimensionsResponseSource            Dimensions = "response_source"
	DimensionsDataPlaneNodeVersion      Dimensions = "data_plane_node_version"
)

func (e Dimensions) ToPointer() *Dimensions {
	return &e
}
func (e *Dimensions) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "api_product":
		fallthrough
	case "api_product_version":
		fallthrough
	case "control_plane":
		fallthrough
	case "control_plane_group":
		fallthrough
	case "data_plane_node":
		fallthrough
	case "gateway_service":
		fallthrough
	case "portal":
		fallthrough
	case "route":
		fallthrough
	case "status_code":
		fallthrough
	case "status_code_grouped":
		fallthrough
	case "time":
		fallthrough
	case "application":
		fallthrough
	case "consumer":
		fallthrough
	case "country_code":
		fallthrough
	case "iso_code":
		fallthrough
	case "upstream_status_code":
		fallthrough
	case "upstream_status_code_grouped":
		fallthrough
	case "response_source":
		fallthrough
	case "data_plane_node_version":
		*e = Dimensions(v)
		return nil
	default:
		return fmt.Errorf("invalid value for Dimensions: %v", v)
	}
}

// Granularity - `granularity` is only valid for queries that include a time dimension, and it specifies the time buckets for the returned data.  For example, `MINUTELY` granularity will return datapoints for every minute.  Not all granularities are available for all time ranges: for example, custom timeframes only have `DAILY` granularity.
//
// If unspecified, a default value for the given time range will be chosen according to the following table:
//
// - `FIFTEEN_MIN`: `MINUTELY`
// - `ONE_HOUR`: `MINUTELY`
// - `SIX_HOUR`: `HOURLY`
// - `TWELVE_HOUR`: `HOURLY`
// - `ONE_DAY`: `HOURLY`
// - `SEVEN_DAY`: `DAILY`
// - `THIRTY_DAY`: `DAILY`
// - `CURRENT_WEEK`: `DAILY`
// - `CURRENT_MONTH`: `DAILY`
// - `PREVIOUS_WEEK`: `DAILY`
// - `PREVIOUS_MONTH`: `DAILY`
type Granularity string

const (
	GranularityFiveMinutely   Granularity = "fiveMinutely"
	GranularityTenMinutely    Granularity = "tenMinutely"
	GranularityThirtyMinutely Granularity = "thirtyMinutely"
	GranularityHourly         Granularity = "hourly"
)

func (e Granularity) ToPointer() *Granularity {
	return &e
}
func (e *Granularity) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "fiveMinutely":
		fallthrough
	case "tenMinutely":
		fallthrough
	case "thirtyMinutely":
		fallthrough
	case "hourly":
		*e = Granularity(v)
		return nil
	default:
		return fmt.Errorf("invalid value for Granularity: %v", v)
	}
}

type FiltersType string

const (
	FiltersTypeAPIProduct        FiltersType = "API_PRODUCT"
	FiltersTypeAPIProductVersion FiltersType = "API_PRODUCT_VERSION"
	FiltersTypeRoute             FiltersType = "ROUTE"
	FiltersTypeApplication       FiltersType = "APPLICATION"
	FiltersTypeStatusCode        FiltersType = "STATUS_CODE"
	FiltersTypeStatusCodeGrouped FiltersType = "STATUS_CODE_GROUPED"
	FiltersTypeGatewayService    FiltersType = "GATEWAY_SERVICE"
	FiltersTypeControlPlane      FiltersType = "CONTROL_PLANE"
)

type Filters struct {
	APIProductFilter        *APIProductFilter        `queryParam:"inline"`
	APIProductVersionFilter *APIProductVersionFilter `queryParam:"inline"`
	RouteFilter             *RouteFilter             `queryParam:"inline"`
	ApplicationFilter       *ApplicationFilter       `queryParam:"inline"`
	StatusCodeFilter        *StatusCodeFilter        `queryParam:"inline"`
	StatusCodeGroupedFilter *StatusCodeGroupedFilter `queryParam:"inline"`
	GatewayServiceFilter    *GatewayServiceFilter    `queryParam:"inline"`
	ControlPlaneFilter      *ControlPlaneFilter      `queryParam:"inline"`

	Type FiltersType
}

func CreateFiltersAPIProduct(apiProduct APIProductFilter) Filters {
	typ := FiltersTypeAPIProduct

	typStr := Dimension(typ)
	apiProduct.Dimension = typStr

	return Filters{
		APIProductFilter: &apiProduct,
		Type:             typ,
	}
}

func CreateFiltersAPIProductVersion(apiProductVersion APIProductVersionFilter) Filters {
	typ := FiltersTypeAPIProductVersion

	typStr := APIProductVersionFilterDimension(typ)
	apiProductVersion.Dimension = typStr

	return Filters{
		APIProductVersionFilter: &apiProductVersion,
		Type:                    typ,
	}
}

func CreateFiltersRoute(route RouteFilter) Filters {
	typ := FiltersTypeRoute

	typStr := RouteFilterDimension(typ)
	route.Dimension = typStr

	return Filters{
		RouteFilter: &route,
		Type:        typ,
	}
}

func CreateFiltersApplication(application ApplicationFilter) Filters {
	typ := FiltersTypeApplication

	typStr := ApplicationFilterDimension(typ)
	application.Dimension = typStr

	return Filters{
		ApplicationFilter: &application,
		Type:              typ,
	}
}

func CreateFiltersStatusCode(statusCode StatusCodeFilter) Filters {
	typ := FiltersTypeStatusCode

	typStr := StatusCodeFilterDimension(typ)
	statusCode.Dimension = typStr

	return Filters{
		StatusCodeFilter: &statusCode,
		Type:             typ,
	}
}

func CreateFiltersStatusCodeGrouped(statusCodeGrouped StatusCodeGroupedFilter) Filters {
	typ := FiltersTypeStatusCodeGrouped

	typStr := StatusCodeGroupedFilterDimension(typ)
	statusCodeGrouped.Dimension = typStr

	return Filters{
		StatusCodeGroupedFilter: &statusCodeGrouped,
		Type:                    typ,
	}
}

func CreateFiltersGatewayService(gatewayService GatewayServiceFilter) Filters {
	typ := FiltersTypeGatewayService

	typStr := GatewayServiceFilterDimension(typ)
	gatewayService.Dimension = typStr

	return Filters{
		GatewayServiceFilter: &gatewayService,
		Type:                 typ,
	}
}

func CreateFiltersControlPlane(controlPlane ControlPlaneFilter) Filters {
	typ := FiltersTypeControlPlane

	typStr := ControlPlaneFilterDimension(typ)
	controlPlane.Dimension = typStr

	return Filters{
		ControlPlaneFilter: &controlPlane,
		Type:               typ,
	}
}

func (u *Filters) UnmarshalJSON(data []byte) error {

	type discriminator struct {
		Dimension string `json:"dimension"`
	}

	dis := new(discriminator)
	if err := json.Unmarshal(data, &dis); err != nil {
		return fmt.Errorf("could not unmarshal discriminator: %w", err)
	}

	switch dis.Dimension {
	case "API_PRODUCT":
		apiProductFilter := new(APIProductFilter)
		if err := utils.UnmarshalJSON(data, &apiProductFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == API_PRODUCT) type APIProductFilter within Filters: %w", string(data), err)
		}

		u.APIProductFilter = apiProductFilter
		u.Type = FiltersTypeAPIProduct
		return nil
	case "API_PRODUCT_VERSION":
		apiProductVersionFilter := new(APIProductVersionFilter)
		if err := utils.UnmarshalJSON(data, &apiProductVersionFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == API_PRODUCT_VERSION) type APIProductVersionFilter within Filters: %w", string(data), err)
		}

		u.APIProductVersionFilter = apiProductVersionFilter
		u.Type = FiltersTypeAPIProductVersion
		return nil
	case "ROUTE":
		routeFilter := new(RouteFilter)
		if err := utils.UnmarshalJSON(data, &routeFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == ROUTE) type RouteFilter within Filters: %w", string(data), err)
		}

		u.RouteFilter = routeFilter
		u.Type = FiltersTypeRoute
		return nil
	case "APPLICATION":
		applicationFilter := new(ApplicationFilter)
		if err := utils.UnmarshalJSON(data, &applicationFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == APPLICATION) type ApplicationFilter within Filters: %w", string(data), err)
		}

		u.ApplicationFilter = applicationFilter
		u.Type = FiltersTypeApplication
		return nil
	case "STATUS_CODE":
		statusCodeFilter := new(StatusCodeFilter)
		if err := utils.UnmarshalJSON(data, &statusCodeFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == STATUS_CODE) type StatusCodeFilter within Filters: %w", string(data), err)
		}

		u.StatusCodeFilter = statusCodeFilter
		u.Type = FiltersTypeStatusCode
		return nil
	case "STATUS_CODE_GROUPED":
		statusCodeGroupedFilter := new(StatusCodeGroupedFilter)
		if err := utils.UnmarshalJSON(data, &statusCodeGroupedFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == STATUS_CODE_GROUPED) type StatusCodeGroupedFilter within Filters: %w", string(data), err)
		}

		u.StatusCodeGroupedFilter = statusCodeGroupedFilter
		u.Type = FiltersTypeStatusCodeGrouped
		return nil
	case "GATEWAY_SERVICE":
		gatewayServiceFilter := new(GatewayServiceFilter)
		if err := utils.UnmarshalJSON(data, &gatewayServiceFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == GATEWAY_SERVICE) type GatewayServiceFilter within Filters: %w", string(data), err)
		}

		u.GatewayServiceFilter = gatewayServiceFilter
		u.Type = FiltersTypeGatewayService
		return nil
	case "CONTROL_PLANE":
		controlPlaneFilter := new(ControlPlaneFilter)
		if err := utils.UnmarshalJSON(data, &controlPlaneFilter, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Dimension == CONTROL_PLANE) type ControlPlaneFilter within Filters: %w", string(data), err)
		}

		u.ControlPlaneFilter = controlPlaneFilter
		u.Type = FiltersTypeControlPlane
		return nil
	}

	return fmt.Errorf("could not unmarshal `%s` into any supported union types for Filters", string(data))
}

func (u Filters) MarshalJSON() ([]byte, error) {
	if u.APIProductFilter != nil {
		return utils.MarshalJSON(u.APIProductFilter, "", true)
	}

	if u.APIProductVersionFilter != nil {
		return utils.MarshalJSON(u.APIProductVersionFilter, "", true)
	}

	if u.RouteFilter != nil {
		return utils.MarshalJSON(u.RouteFilter, "", true)
	}

	if u.ApplicationFilter != nil {
		return utils.MarshalJSON(u.ApplicationFilter, "", true)
	}

	if u.StatusCodeFilter != nil {
		return utils.MarshalJSON(u.StatusCodeFilter, "", true)
	}

	if u.StatusCodeGroupedFilter != nil {
		return utils.MarshalJSON(u.StatusCodeGroupedFilter, "", true)
	}

	if u.GatewayServiceFilter != nil {
		return utils.MarshalJSON(u.GatewayServiceFilter, "", true)
	}

	if u.ControlPlaneFilter != nil {
		return utils.MarshalJSON(u.ControlPlaneFilter, "", true)
	}

	return nil, errors.New("could not marshal union type Filters: all fields are null")
}

type Explore struct {
	// The period of time to return data.  Relative time ranges are relative to the current moment.  Absolute time ranges specify an unchanging period of time.  If not specified, a default relative timeframe of last 24 hours will be chosen.
	//
	TimeRange *TimeRange `json:"time_range,omitempty"`
	// A property of your API (such as request count or latency) that you wish to report on.
	// Your chosen metric is aggregated within the specified dimensions, meaning that if you query 'request count by service', you'll receive the total number of requests each service received within the given time frame.  Some metrics, such as latency and response size, have more complicated aggregations: selecting P99 will result in the 99th percentile of the chosen metric.
	//
	Metrics []Metrics `json:"metrics,omitempty"`
	// The dimensions for the report.  A report may have up to 2 dimensions, including time.
	// If the report has a timeseries graph, the time dimension will be added automatically if not provided.
	// If no dimensions are provided, the report will simply return the provided metric aggregated across
	// all available data.
	//
	Dimensions []Dimensions `json:"dimensions,omitempty"`
	// `granularity` is only valid for queries that include a time dimension, and it specifies the time buckets for the returned data.  For example, `MINUTELY` granularity will return datapoints for every minute.  Not all granularities are available for all time ranges: for example, custom timeframes only have `DAILY` granularity.
	//
	// If unspecified, a default value for the given time range will be chosen according to the following table:
	//
	// - `FIFTEEN_MIN`: `MINUTELY`
	// - `ONE_HOUR`: `MINUTELY`
	// - `SIX_HOUR`: `HOURLY`
	// - `TWELVE_HOUR`: `HOURLY`
	// - `ONE_DAY`: `HOURLY`
	// - `SEVEN_DAY`: `DAILY`
	// - `THIRTY_DAY`: `DAILY`
	// - `CURRENT_WEEK`: `DAILY`
	// - `CURRENT_MONTH`: `DAILY`
	// - `PREVIOUS_WEEK`: `DAILY`
	// - `PREVIOUS_MONTH`: `DAILY`
	//
	Granularity *Granularity `json:"granularity,omitempty"`
	Filters     []Filters    `json:"filters,omitempty"`
}

func (o *Explore) GetTimeRange() *TimeRange {
	if o == nil {
		return nil
	}
	return o.TimeRange
}

func (o *Explore) GetTimeRangeRelative() *RelativeTimeRange {
	if v := o.GetTimeRange(); v != nil {
		return v.RelativeTimeRange
	}
	return nil
}

func (o *Explore) GetTimeRangeAbsolute() *AbsoluteTimeRange {
	if v := o.GetTimeRange(); v != nil {
		return v.AbsoluteTimeRange
	}
	return nil
}

func (o *Explore) GetMetrics() []Metrics {
	if o == nil {
		return nil
	}
	return o.Metrics
}

func (o *Explore) GetDimensions() []Dimensions {
	if o == nil {
		return nil
	}
	return o.Dimensions
}

func (o *Explore) GetGranularity() *Granularity {
	if o == nil {
		return nil
	}
	return o.Granularity
}

func (o *Explore) GetFilters() []Filters {
	if o == nil {
		return nil
	}
	return o.Filters
}

type Query struct {
	Version *string `json:"version,omitempty"`
	// Visualization type selected for this report.
	ChartType  *ChartType `json:"chartType,omitempty"`
	Datasource *string    `json:"datasource,omitempty"`
	Explore    *Explore   `json:"explore,omitempty"`
}

func (o *Query) GetVersion() *string {
	if o == nil {
		return nil
	}
	return o.Version
}

func (o *Query) GetChartType() *ChartType {
	if o == nil {
		return nil
	}
	return o.ChartType
}

func (o *Query) GetDatasource() *string {
	if o == nil {
		return nil
	}
	return o.Datasource
}

func (o *Query) GetExplore() *Explore {
	if o == nil {
		return nil
	}
	return o.Explore
}

type Report struct {
	// The ID of the report.
	ID *string `json:"id,omitempty"`
	// An ISO-8601 timestamp representing when the report was created.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// An ISO-8601 timestamp representing when the report was last updated.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// The UUID of the Konnect user that created the report.
	CreatedBy *string `json:"created_by,omitempty"`
	// The user-provided name for the report.
	// If not provided, the report will be named "Untitled Report" with a timestamp suffix.
	//
	Name *string `json:"name,omitempty"`
	// An optional extended description for the report.
	Description *string `json:"description,omitempty"`
	// The org id owning the report.
	OrgID *string `json:"org_id,omitempty"`
	Query *Query  `json:"query,omitempty"`
}

func (r Report) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(r, "", false)
}

func (r *Report) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &r, "", false, false); err != nil {
		return err
	}
	return nil
}

func (o *Report) GetID() *string {
	if o == nil {
		return nil
	}
	return o.ID
}

func (o *Report) GetCreatedAt() *time.Time {
	if o == nil {
		return nil
	}
	return o.CreatedAt
}

func (o *Report) GetUpdatedAt() *time.Time {
	if o == nil {
		return nil
	}
	return o.UpdatedAt
}

func (o *Report) GetCreatedBy() *string {
	if o == nil {
		return nil
	}
	return o.CreatedBy
}

func (o *Report) GetName() *string {
	if o == nil {
		return nil
	}
	return o.Name
}

func (o *Report) GetDescription() *string {
	if o == nil {
		return nil
	}
	return o.Description
}

func (o *Report) GetOrgID() *string {
	if o == nil {
		return nil
	}
	return o.OrgID
}

func (o *Report) GetQuery() *Query {
	if o == nil {
		return nil
	}
	return o.Query
}

// ReportInput - The request schema for the create report request.
//
// If you pass the same `name` and `description` of an existing report in the request, a report with the same `name` and `description` will be created. The two reports will have different `id` values to differentiate them.
//
// Note that all fields are optional: if you pass an empty JSON object as the request (`{}`), a new report will be created with a default configuration.
type ReportInput struct {
	// The user-provided name for the report.
	// If not provided, the report will be named "Untitled Report" with a timestamp suffix.
	//
	Name *string `json:"name,omitempty"`
	// An optional extended description for the report.
	Description *string `json:"description,omitempty"`
	Query       *Query  `json:"query,omitempty"`
}

func (o *ReportInput) GetName() *string {
	if o == nil {
		return nil
	}
	return o.Name
}

func (o *ReportInput) GetDescription() *string {
	if o == nil {
		return nil
	}
	return o.Description
}

func (o *ReportInput) GetQuery() *Query {
	if o == nil {
		return nil
	}
	return o.Query
}
