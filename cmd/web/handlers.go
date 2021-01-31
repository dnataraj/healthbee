// Package main provides the implementation for the HealthBee API
package main

import (
	"errors"
	"github.com/dnataraj/healthbee/pkg/models"
	"net/http"
)

// createSite is a POST HTTP handler that accepts a JSON payload and creates a site entry,
// and initiates the monitoring for this site
// The handler expects the request body to have the following schema
// { "url": <string>, "period": <int>, "pattern": <string> }
func (app *application) monitor(w http.ResponseWriter, r *http.Request) {
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
	//mon := app.NewMonitor(&site)
	app.infoLog.Printf("starting HealthBee for site: %d", site.ID)
	//mon.Start(app.wg)

	app.respond(w, site, http.StatusOK)
}
