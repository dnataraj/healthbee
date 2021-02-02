package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dnataraj/healthbee/pkg"
	"github.com/dnataraj/healthbee/pkg/models"
	"github.com/segmentio/kafka-go"
	"net/http"
	"runtime/debug"
	"sync"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// decode is a simple response deserializer.
// If the destination interface as an OK method, this can be used for simple validation
func decode(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return err
	}
	if valid, ok := v.(interface {
		OK() error
	}); ok {
		err = valid.OK()
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	app.respond(w, "{}", http.StatusOK)
}

func (app *application) respond(w http.ResponseWriter, v interface{}, code int) {
	w.WriteHeader(code)
	if v == nil {
		return
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

// NewMonitor initializes and returns a site monitor. The monitor encapsulates
// context cancellation and is retained in a map so that it can be managed afterwards
func (app *application) NewMonitor(s *models.Site) *pkg.Monitor {
	m := pkg.NewMonitor(s, app.writer)
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	app.monitors[s.ID] = m

	return m
}

// Start resumes monitoring for the last 20 (for now) registered sites when HealthBee is started
func (app *application) resume() {
	sites, err := app.sites.GetAll()
	if err != nil {
		app.errorLog.Fatal("server: unable to resume monitoring, failed with: ", err)
	}
	app.infoLog.Printf("found sites %+v", sites)
	for _, site := range sites {
		m := app.NewMonitor(site)
		app.infoLog.Printf("server: resuming monitoring for site [%d] with address [%s]...", site.ID, site.URL)
		m.Start(app.wg)
	}
}

// read consumes messages from a specific Kafka topic and publishes this to a PostgreSQL database
// These are the site availability metrics previously published by the site monitors
// Readers (a.k.a auditors) can be cancelled via the passed in Context
// TODO: This belongs in pkg along with Monitor
func (app *application) read(ctx context.Context, id int, r *kafka.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				app.errorLog.Printf("auditor %d: unable to read message: %s", id, err.Error())
				return
			}
			app.infoLog.Printf("auditor %d: fetched result: %s", id, string(msg.Value))
			res := models.CheckResult{}
			if err := json.Unmarshal(msg.Value, &res); err != nil {
				app.errorLog.Printf("auditor %d: unable to detect valid message: %s", id, err.Error())
				return
			}
			resID, err := app.results.Insert(res.SiteID, res.At, res.ResponseCode, res.MatchedPattern)
			if err != nil {
				app.errorLog.Printf("auditor %d: unable to write metrics for site [%d], failing with: %s", id, res.SiteID, err.Error())
				return
			}
			app.infoLog.Printf("auditor %d: added metrics for site [%d], with id: %d", id, res.SiteID, resID)
		}
	}
}
