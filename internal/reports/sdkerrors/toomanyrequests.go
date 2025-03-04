// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package sdkerrors

import (
	"encoding/json"
)

// TooManyRequests - The error response object.
type TooManyRequests struct {
	// Documentation for this error.
	Type *string `json:"type,omitempty"`
	// The HTTP status code.
	Status *int64 `json:"status,omitempty"`
	// The error response code.
	Title *string `json:"title,omitempty"`
	// The Konnect traceback code
	Instance *string `json:"instance,omitempty"`
	// Details about the error.
	Detail *string `json:"detail,omitempty"`
}

var _ error = &TooManyRequests{}

func (e *TooManyRequests) Error() string {
	data, _ := json.Marshal(e)
	return string(data)
}
