package models

import (
	"encoding/json"
	"errors"
	"time"
)

// Period represents a time interval or period as a time.Duration
// Simple marshalling/unmarshalling as demonstrated here:
// https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations
type Period time.Duration

func (p Period) Duration() time.Duration {
	return time.Duration(p)
}

func (p Period) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Duration().String())
}

func (p *Period) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*p = Period(time.Duration(value))
		return nil
	case string:
		var err error
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*p = Period(d)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

var ErrDuplicateSite = errors.New("sites: duplicate site registration")
var ErrNoRecord = errors.New("sites: no record found")

type Site struct {
	ID         int          `json:"id,omitempty"`
	URL        string       `json:"url"`
	Interval   Period       `json:"interval"`
	Pattern    string       `json:"pattern"`
	LastResult *CheckResult `json:",omitempty"`
}

//TODO: in retrospect this is a not a good name for the struct, it should be HealthCheckResult or just Result
type CheckResult struct {
	ID             int       `json:"id"`
	SiteID         int       `json:"site_id"`
	At             time.Time `json:"at"`
	ResponseCode   int       `json:"response_code"`
	MatchedPattern bool      `json:"matched"`
}
