package postgres

import (
	"github.com/dnataraj/healthbee/pkg/models"
	"testing"
	"time"
)

func TestResultModel_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	tests := []struct {
		name       string
		id         int
		wantResult *models.CheckResult
		wantError  error
	}{
		{
			name: "Valid result",
			id:   1,
			wantResult: &models.CheckResult{
				ID:             1,
				SiteID:         1,
				At:             time.Now().UTC(),
				ResponseTime:   models.Period(time.Duration(600) * time.Millisecond),
				ResponseCode:   200,
				MatchedPattern: true,
			},
			wantError: nil,
		},
		{
			name:       "Invalid result",
			id:         10,
			wantResult: nil,
			wantError:  models.ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, teardown := newTestDB(t)
			defer teardown()

			r := ResultModel{DB: db}
			res, err := r.Get(tt.id)
			if err != tt.wantError {
				t.Errorf("want %v, got %s", tt.wantError, err)
			}
			// Not validating the time field for this exercise, otherise reflect.DeepEqual() gets this done
			if !equals(res, tt.wantResult) {
				t.Errorf("want %v, got %v", tt.wantResult, res)
			}
		})
	}
}

func TestResultModel_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	at := time.Now().UTC()
	tests := []struct {
		name         string
		siteID       int
		at           time.Time
		responseTime models.Period
		responseCode int
		matched      bool
		wantResult   int
		wantError    error
	}{
		{
			name:         "Valid insert",
			siteID:       1,
			at:           at,
			responseTime: models.Period(300 * time.Millisecond),
			responseCode: 200,
			matched:      true,
			wantResult:   4,
			wantError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, teardown := newTestDB(t)
			defer teardown()

			r := &ResultModel{DB: db}
			id, err := r.Insert(tt.siteID, tt.at, tt.responseTime, tt.responseCode, tt.matched)
			if err != tt.wantError {
				t.Errorf("want %v, got %s", tt.wantError, err)
			}
			if id != tt.wantResult {
				t.Errorf("want %d, got %d", tt.wantResult, id)
			}
		})
	}
}

func TestResultModel_GetResultsForSite(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	t.Run("Get all metrics", func(t *testing.T) {
		db, teardown := newTestDB(t)
		defer teardown()

		r := ResultModel{DB: db}
		results, err := r.GetResultsForSite(2)
		if err != nil {
			t.Fatal(err)
		}

		// check number of metrics returned
		if len(results) != 2 {
			t.Errorf("want 2 metrics, got %d", len(results))
		}
		// test validity of ordering
		res := results[0]
		if res.ResponseTime.Duration().Milliseconds() != 1200 {
			t.Errorf("want response time 200, got %d", res.ResponseTime)
		}

	})
}

func equals(r1, r2 *models.CheckResult) bool {
	if r1 == nil || r2 == nil {
		return r1 == r2
	}
	if r1.ID != r2.ID ||
		r1.ResponseTime != r2.ResponseTime ||
		r1.ResponseCode != r2.ResponseCode ||
		r1.SiteID != r2.SiteID ||
		r1.MatchedPattern != r2.MatchedPattern {
		return false
	}
	return true
}
