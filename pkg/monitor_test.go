package pkg

import (
	"github.com/dnataraj/healthbee/pkg/models"
	"sync"
	"testing"
	"time"
)

// A simple, practical integration test to verify publishing messages
func TestMonitor_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("kafka: skipping integration test")
	}
	// Load some sites
	// site1: monitor at 3 second intervals
	site1 := &models.Site{
		ID:       0,
		URL:      "https://www.example.com",
		Interval: models.Period(5) * models.Period(time.Second),
		Pattern:  "content",
		Created:  time.Now().UTC(),
	}
	// site2: monitor at 5 second intervals
	site2 := &models.Site{
		ID:       1,
		URL:      "https://www.example.com",
		Interval: models.Period(3) * models.Period(time.Second),
		Pattern:  "content",
		Created:  time.Now().UTC().Add(2 * time.Second),
	}

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
