package types

import (
	"encoding/json"
	"strconv"
	"time"
)

type DurationInSeconds time.Duration

func (d DurationInSeconds) Value() time.Duration {
	return time.Duration(d)
}

func (d *DurationInSeconds) SetValue(s string) error {
	if s == "" {
		return nil
	}

	sec, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	*d = DurationInSeconds(time.Duration(sec) * time.Second)

	return nil

}

// UnmarshalJSON implements json.Unmarshaler to support both "1s" format and integer seconds
func (d *DurationInSeconds) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// Try parsing as duration string (e.g., "1s", "300s")
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*d = DurationInSeconds(duration)
		return nil
	}

	// Try parsing as integer seconds
	var sec int
	if err := json.Unmarshal(data, &sec); err != nil {
		return err
	}
	*d = DurationInSeconds(time.Duration(sec) * time.Second)
	return nil
}
