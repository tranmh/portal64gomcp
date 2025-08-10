package types

import (
	"encoding/json"
	"time"
)

// Duration is a custom type that wraps time.Duration to provide proper JSON serialization
type Duration time.Duration

// MarshalJSON implements the json.Marshaler interface
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	
	*d = Duration(duration)
	return nil
}

// ToDuration converts Duration to time.Duration
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

// FromDuration converts time.Duration to Duration
func FromDuration(td time.Duration) Duration {
	return Duration(td)
}

// String returns the string representation of the duration
func (d Duration) String() string {
	return time.Duration(d).String()
}
