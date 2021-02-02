package postgres

import (
	"github.com/dnataraj/healthbee/pkg/models"
	"reflect"
	"testing"
	"time"
)

func TestSiteModel_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	created := time.Now().UTC()
	tests := []struct {
		name       string
		id         int
		wantResult *models.Site
		wantError  error
	}{
		{
			name: "Valid site",
			id:   1,
			wantResult: &models.Site{
				ID:       1,
				URL:      "https://www.example.com",
				Interval: models.Period(time.Duration(5) * time.Second),
				Pattern:  "content",
				Created:  created,
			},
			wantError: nil,
		},
		{
			name:       "Missing site",
			id:         6,
			wantResult: nil,
			wantError:  models.ErrNoRecord,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, teardown := newTestDB(t)
			defer teardown()

			s := &SiteModel{DB: db}
			res, err := s.Get(tt.id)
			if err != tt.wantError {
				t.Errorf("want %v, got %s", tt.wantError, err)
			}
			if tt.wantError == nil && res != nil {
				res.Created = created
			}
			if !reflect.DeepEqual(res, tt.wantResult) {
				t.Errorf("want %v, got %v", tt.wantResult, res)
			}
		})
	}
}

func TestSiteModel_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	tests := []struct {
		name       string
		url        string
		interval   models.Period
		pattern    string
		wantResult int
		wantError  error
	}{
		{
			name:       "Valid insert 1",
			url:        "http://site1/test",
			interval:   models.Period(5) * models.Period(time.Second),
			pattern:    "abc{2,}",
			wantResult: 3,
			wantError:  nil,
		},
		{
			name:       "Valid insert 2",
			url:        "http://site1/test",
			interval:   models.Period(5) * models.Period(time.Second),
			pattern:    "foo?bar",
			wantResult: 3,
			wantError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, teardown := newTestDB(t)
			defer teardown()

			s := &SiteModel{DB: db}
			id, err := s.Insert(tt.url, tt.interval, tt.pattern)
			if err != tt.wantError {
				t.Errorf("want %v, got %s", tt.wantError, err)
			}

			if id != tt.wantResult {
				t.Errorf("want %d, got %d", tt.wantResult, id)
			}
		})
	}

	t.Run("Duplicate insert", func(t *testing.T) {
		db, teardown := newTestDB(t)
		defer teardown()

		s := &SiteModel{DB: db}
		_, err := s.Insert("http://site1/test", models.Period(5)*models.Period(time.Second), "test")
		if err != nil {
			t.Errorf("want nil, got %v", err)
		}
		id, err := s.Insert("http://site1/test", models.Period(5)*models.Period(time.Second), "test")

		if err != models.ErrDuplicateSite {
			t.Errorf("want %v, got %s", models.ErrDuplicateSite, err)
		}

		if id != -1 {
			t.Errorf("want %d, got %d", -1, id)
		}
	})

}

//TODO: In a similar way, exploratory tests can be added also for GetResultsForSite
