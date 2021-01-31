package models

import (
	"encoding/json"
	"errors"
	"time"
)

// Period represents a time interval or period as a time.Duration
// Simple marshalling/unmarshalling as demonstrated here:
// https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations
type Period struct {
	time.Duration
}

func (p Period) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())

}

func (p *Period) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		p.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		p.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type Site struct {
	ID         int          `json:"id,omitempty"`
	URL        string       `json:"url"`
	Interval   Period       `json:"interval"`
	Pattern    string       `json:"pattern"`
	LastResult *CheckResult `json:",omitempty"`
}

type CheckResult struct {
	At           time.Time
	ResponseCode int
	FoundPattern bool
}
