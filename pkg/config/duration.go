package config

import (
	"encoding/json"
	"time"
)

// Custom wrapper around time.Duration to implement json marshalling
type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	res, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(res)

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d Duration) String() string {
	return time.Duration(d).String()
}
