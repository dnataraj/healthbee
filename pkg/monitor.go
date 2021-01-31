package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dnataraj/healthbee/pkg/models"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var client = &http.Client{}
var infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
var warnLog = log.New(os.Stderr, "WARN\t", log.Ldate|log.Ltime)

// Monitor represents the availability check for each site
type Monitor struct {
	Site    *models.Site
	Context context.Context
	Cancel  context.CancelFunc
	writer  *kafka.Writer
}

func NewMonitor(s *models.Site, w *kafka.Writer) *Monitor {
	m := &Monitor{}
	m.Site = s
	m.Context, m.Cancel = context.WithCancel(context.Background())
	m.writer = w
	return m
}

// Start starts the monitor for the site in a goroutine. The wait group operand is incremented and the monitoring
// results are published to a Kafka topic.
func (m *Monitor) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		t := time.NewTicker(m.Site.Interval.Duration())
		defer t.Stop()

		for {
			select {
			case now := <-t.C:
				at := now.UTC()
				infoLog.Printf("monitor: site [%d] checked at %s", m.Site.ID, at.Format(time.Stamp))
				// process the site
				res, err := m.getResult(at)
				if err != nil {
					warnLog.Printf("monitor: site[%d] check failed at %s, with: %s", m.Site.ID, at.Format(time.Stamp), err.Error())
				}
				// publish the metrics to kafka
				infoLog.Printf("monitor: site[%d] publishing metrics to kafka: %+v", m.Site.ID, res)
				err = m.publishResult(res)
				if err != nil {
					warnLog.Printf("monitor: site[%d] check failed at %s, with: %s", m.Site.ID, at.Format(time.Stamp), err.Error())
				}
			case <-m.Context.Done():
				fmt.Printf("site %d> monitoring stopped.\n", m.Site.ID)
				return
			}
		}
	}(wg)
}

func (m *Monitor) getResult(at time.Time) (*models.CheckResult, error) {
	req, err := http.NewRequest("GET", m.Site.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed with: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &models.CheckResult{
			SiteID:         m.Site.ID,
			At:             at,
			ResponseCode:   -1,
			MatchedPattern: false,
		}, fmt.Errorf("fetch failed with: %s", err)
	}
	defer resp.Body.Close()
	// read the body and check if pattern exists
	// assumption that content search is required even for non 200 responses
	data, err := ioutil.ReadAll(resp.Body)
	matcher := regexp.MustCompile(m.Site.Pattern)
	found := matcher.MatchString(string(data))

	return &models.CheckResult{
		SiteID:         m.Site.ID,
		At:             at,
		ResponseCode:   resp.StatusCode,
		MatchedPattern: found,
	}, nil
}

func (m *Monitor) publishResult(res *models.CheckResult) error {
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("publish failed with: %s", err)
	}
	err = m.writer.WriteMessages(m.Context, kafka.Message{
		Key:   []byte(strconv.Itoa(m.Site.ID)),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("publish failed with: %s", err)
	}
	return nil
}
