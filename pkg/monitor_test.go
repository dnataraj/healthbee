package pkg

import (
	"github.com/dnataraj/healthbee/pkg/models"
	"sync"
	"testing"
	"time"
)

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

	// sleep for 6 seconds
	time.Sleep(6 * time.Second)

}
