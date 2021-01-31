package pkg

import (
	"context"
	"fmt"
	"github.com/dnataraj/healthbee/pkg/models"
	"sync"
	"time"
)

// Monitor represents the availability check for each site
type Monitor struct {
	Site    *models.Site
	Context context.Context
	Cancel  context.CancelFunc
}

func (m *Monitor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	t := time.NewTicker(m.Site.Interval.Duration())
	defer t.Stop()

	for {
		select {
		case now := <-t.C:
			fmt.Printf("site %d> monitoring at %s\n", m.Site.ID, now.UTC().Format("Wed Feb 25 11:06:39.1234 PST 2015"))
			// fetch the site
			// process result -> push to topic?
		case <-m.Context.Done():
			fmt.Printf("site %d> monitoring stopped.\n", m.Site.ID)
			return
		}
	}
}
