package models

import "time"

type Site struct {
	ID         int
	URL        string
	Period     time.Time
	Pattern    string
	LastResult *CheckResult
}

type CheckResult struct {
	At           time.Time
	ResponseCode int
	FoundPattern bool
}
