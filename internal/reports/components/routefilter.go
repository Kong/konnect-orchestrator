// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package components

import (
	"encoding/json"
	"fmt"
)

// RouteFilterDimension - The dimension to filter.
type RouteFilterDimension string

const (
	RouteFilterDimensionRoute RouteFilterDimension = "ROUTE"
)

func (e RouteFilterDimension) ToPointer() *RouteFilterDimension {
	return &e
}
func (e *RouteFilterDimension) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "ROUTE":
		*e = RouteFilterDimension(v)
		return nil
	default:
		return fmt.Errorf("invalid value for RouteFilterDimension: %v", v)
	}
}

type RouteFilter struct {
	// The dimension to filter.
	Dimension RouteFilterDimension `json:"dimension"`
	// The type of filter to apply.  `IN` filters will limit results to only the specified values, while `NOT_IN` filters will exclude the specified values.
	Type FilterType `json:"type"`
	// The routes to include in the results.  Because route UUIDs are only unique within a given control plane, the filter values must be of the form "[control plane UUID]:[route UUID]".
	//
	Values []string `json:"values"`
}

func (o *RouteFilter) GetDimension() RouteFilterDimension {
	if o == nil {
		return RouteFilterDimension("")
	}
	return o.Dimension
}

func (o *RouteFilter) GetType() FilterType {
	if o == nil {
		return FilterType("")
	}
	return o.Type
}

func (o *RouteFilter) GetValues() []string {
	if o == nil {
		return []string{}
	}
	return o.Values
}
