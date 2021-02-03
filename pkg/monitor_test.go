package pkg

import (
	"fmt"
	"github.com/dnataraj/healthbee/pkg/models"
	"sync"
	"testing"
	"time"
)

const (
	addr           = "localhost:4444"
	TestHTTPServer = "http://" + addr
)

var site1 = &models.Site{
	ID:       0,
	URL:      "https://www.example.com",
	Interval: models.Period(5) * models.Period(time.Second),
	Pattern:  "found",
	Created:  time.Now().UTC(),
}

// site2: monitor at 5 second intervals
var site2 = &models.Site{
	ID:       1,
	URL:      "https://www.example.com",
	Interval: models.Period(3) * models.Period(time.Second),
	Pattern:  "content",
	Created:  time.Now().UTC().Add(2 * time.Second),
}

// A simple, practical integration test to verify publishing messages
func TestMonitor_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("kafka: skipping integration test")
	}
	// Load some sites
	// site1: monitor at 3 second intervals

	// set up the environment
	w, teardown := newWriter(t)
	defer teardown(t)

	wg := sync.WaitGroup{}
	m1 := NewMonitor(site1, w)
	m1.Start(&wg)

	m2 := NewMonitor(site2, w)
	m2.Start(&wg)

	//TODO: Create a consumer and read a couple of messages for verification

	// sleep for 6 seconds, at least 2 messages should be published
	time.Sleep(6 * time.Second)
	// halt the monitors
	m1.Cancel()
	m2.Cancel()

}

// Test that sites are being monitored, and results match
func TestMonitor_getResult(t *testing.T) {
	site1.URL = fmt.Sprintf("%s/test/site1", TestHTTPServer)
	site2.URL = fmt.Sprintf("%s/test/site2", TestHTTPServer)

	tests := []struct {
		name      string
		site      *models.Site
		checkedAt time.Time
		path      string
		respCode  int
		respBody  string
		wantMatch bool
		wantError error
	}{
		{
			name:      "Pattern found",
			site:      site1,
			checkedAt: time.Now().UTC(),
			path:      "/test/site1",
			respCode:  200,
			respBody:  `<html>found</html>`,
			wantMatch: true,
			wantError: nil,
		},
		{
			name:      "Pattern not found",
			site:      site2,
			checkedAt: time.Now().UTC().Add(5 * time.Second),
			path:      "/test/site2",
			respCode:  200,
			respBody:  `<html>not found</html>`,
			wantMatch: false,
			wantError: nil,
		},
		{
			name:      "Site not found",
			site:      site2,
			checkedAt: time.Now().UTC().Add(5 * time.Second),
			path:      "/test/notthere",
			respCode:  404,
			respBody:  `<html>not found</html>`,
			wantMatch: false,
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, teardown := NewTestServer(t, addr, Procedure{
				URL:    tt.path,
				Method: "GET",
				Response: Response{
					StatusCode: tt.respCode,
					Body:       []byte(tt.respBody),
				},
			})
			ts.Start()
			defer teardown()

			// Start monitoring
			m := NewMonitor(tt.site, nil)
			defer m.Cancel()
			res, err := m.getResult(tt.checkedAt)
			if err != tt.wantError {
				t.Errorf("want nil, got %s", err)
			}

			if tt.respCode != res.ResponseCode {
				t.Errorf("want %d, got %d", tt.respCode, res.ResponseCode)
			}
			if tt.wantMatch != res.MatchedPattern {
				t.Errorf("want %v, got %v", tt.wantMatch, res.MatchedPattern)
			}
			if tt.checkedAt != res.At {
				t.Errorf("want %s, got %s", tt.checkedAt, res.At)
			}
		})
	}
}
