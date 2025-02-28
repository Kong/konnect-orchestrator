// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package sdkerrors

import (
	"encoding/json"
)

// PermissionDenied - The error object.
type PermissionDenied struct {
	// Documentation for this error.
	Type *string `json:"type,omitempty"`
	// HTTP status code.
	Status *int64 `json:"status,omitempty"`
	// HTTP status code
	Title *string `json:"title,omitempty"`
	// Konnect traceback error code.
	Instance *string `json:"instance,omitempty"`
	// Information about the error response.
	Detail *string `json:"detail,omitempty"`
}

var _ error = &PermissionDenied{}

func (e *PermissionDenied) Error() string {
	data, _ := json.Marshal(e)
	return string(data)
}