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

			s := SiteModel{DB: db}
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
