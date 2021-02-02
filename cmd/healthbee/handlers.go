// Package main provides the implementation for the HealthBee API
package main

import (
	"errors"
	"fmt"
	"github.com/dnataraj/healthbee/pkg/models"
	"net/http"
)

// monitor is a POST HTTP handler that accepts a JSON payload and creates a site entry,
// and initiates the monitoring for this site
// The handler expects the request body to have the following schema
// { "url": <string>, "period": <int>, "pattern": <string> }
// Duplicate site registrations are not allowed and results in a HTTP 409
func (app *application) monitor(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		app.list(w, r)
		return
	}
	site := models.Site{}
	err := decode(r, &site)
	if err != nil {
		app.errorLog.Print("error processing request: ", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// generate an entry for site in the database
	site.ID, err = app.sites.Insert(site.URL, site.Interval.Duration(), site.Pattern)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateSite) {
			app.clientError(w, http.StatusConflict)
		} else {
			app.serverError(w, err)
		}
		return
	}
	// if successful, initiate checks
	mon := app.NewMonitor(&site)
	app.infoLog.Printf("starting HealthBee for site: %d", site.ID)
	mon.Start(app.wg)

	w.Header().Add("Location", fmt.Sprintf("/monitor/%d", site.ID))
	app.respond(w, site, http.StatusCreated)
}

// list is a GET HTTP handler that returns a list of registered sites
func (app *application) list(w http.ResponseWriter, r *http.Request) {
	app.respond(w, "{}", http.StatusNotImplemented)
}

// stop is a POST HTTP handler that stops a monitor for a given site
func (app *application) stop(w http.ResponseWriter, r *http.Request) {
	app.respond(w, "{}", http.StatusNotImplemented)
}

// getMetrics returns a list of the last 20 metrics for the given site
func (app *application) getMetrics(w http.ResponseWriter, r *http.Request) {
	app.respond(w, "{}", http.StatusNotImplemented)
}
