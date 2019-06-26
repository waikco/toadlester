package model

import (
	"encoding/json"
	"strings"
	"time"
)

// custom type for destructing string representation of time durations from json
type CustomDuration struct {
	time.Duration
}

func (d *CustomDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *CustomDuration) UnmarshalJSON(b []byte) (err error) {
	d.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return nil
}
