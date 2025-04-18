// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package components

// StringFieldOContainsFilter - Returns entities that fuzzy-match any of the comma-delimited phrases in the filter string.
type StringFieldOContainsFilter struct {
	Ocontains string `queryParam:"name=ocontains"`
}

func (o *StringFieldOContainsFilter) GetOcontains() string {
	if o == nil {
		return ""
	}
	return o.Ocontains
}
