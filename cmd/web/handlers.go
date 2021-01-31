// Package main provides the implementation for the Healthbee API
package main

import (
	"github.com/dnataraj/healthbee/pkg/models"
	"net/http"
)

// createSite is a POST HTTP handler that accepts a JSON payload and creates a site entry,
// and initiates the Healthbee monitor for this site
// The handler expects the request body to have the following schema
// { "url": <string>, "period": <int>, "pattern": <string> }
func (app *application) createSite(w http.ResponseWriter, r *http.Request) {
	site := models.Site{}
	err := decode(r, &site)
	if err != nil {
		app.errorLog.Print("error processing request: ", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// generate an entry for site in the database
	site.ID, err = app.sites.Insert(site.URL, site.Interval, site.Pattern)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// if successful, initiate checks
	app.respond(w, site, http.StatusOK)
}
