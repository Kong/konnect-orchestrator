// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package components

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Kong/konnect-orchestrator/internal/reports/utils"
)

type TimeRangeType string

const (
	TimeRangeTypeRelative TimeRangeType = "relative"
	TimeRangeTypeAbsolute TimeRangeType = "absolute"
)

// TimeRange - The period of time to return data.  Relative time ranges are relative to the current moment.  Absolute time ranges specify an unchanging period of time.  If not specified, a default relative timeframe of last 24 hours will be chosen.
type TimeRange struct {
	RelativeTimeRange *RelativeTimeRange `queryParam:"inline"`
	AbsoluteTimeRange *AbsoluteTimeRange `queryParam:"inline"`

	Type TimeRangeType
}

func CreateTimeRangeRelative(relative RelativeTimeRange) TimeRange {
	typ := TimeRangeTypeRelative

	typStr := Type(typ)
	relative.Type = typStr

	return TimeRange{
		RelativeTimeRange: &relative,
		Type:              typ,
	}
}

func CreateTimeRangeAbsolute(absolute AbsoluteTimeRange) TimeRange {
	typ := TimeRangeTypeAbsolute

	typStr := AbsoluteTimeRangeType(typ)
	absolute.Type = typStr

	return TimeRange{
		AbsoluteTimeRange: &absolute,
		Type:              typ,
	}
}

func (u *TimeRange) UnmarshalJSON(data []byte) error {

	type discriminator struct {
		Type string `json:"type"`
	}

	dis := new(discriminator)
	if err := json.Unmarshal(data, &dis); err != nil {
		return fmt.Errorf("could not unmarshal discriminator: %w", err)
	}

	switch dis.Type {
	case "relative":
		relativeTimeRange := new(RelativeTimeRange)
		if err := utils.UnmarshalJSON(data, &relativeTimeRange, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == relative) type RelativeTimeRange within TimeRange: %w", string(data), err)
		}

		u.RelativeTimeRange = relativeTimeRange
		u.Type = TimeRangeTypeRelative
		return nil
	case "absolute":
		absoluteTimeRange := new(AbsoluteTimeRange)
		if err := utils.UnmarshalJSON(data, &absoluteTimeRange, "", true, false); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == absolute) type AbsoluteTimeRange within TimeRange: %w", string(data), err)
		}

		u.AbsoluteTimeRange = absoluteTimeRange
		u.Type = TimeRangeTypeAbsolute
		return nil
	}

	return fmt.Errorf("could not unmarshal `%s` into any supported union types for TimeRange", string(data))
}

func (u TimeRange) MarshalJSON() ([]byte, error) {
	if u.RelativeTimeRange != nil {
		return utils.MarshalJSON(u.RelativeTimeRange, "", true)
	}

	if u.AbsoluteTimeRange != nil {
		return utils.MarshalJSON(u.AbsoluteTimeRange, "", true)
	}

	return nil, errors.New("could not marshal union type TimeRange: all fields are null")
}
