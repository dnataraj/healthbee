// Package main provides the implementation for the Healthbee API
package main

import (
	"github.com/dnataraj/healthbee/pkg/models"
	"net/http"
)

// createSite is a POST HTTP handler that accepts a JSON payload and creates a site entry,
// and initiates the Healthbee monitor for this site
func (app *application) createSite(w http.ResponseWriter, r *http.Request) {
	site := models.Site{}
	err := decode(r, &site)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// generate an entry for site in the database

	// if successful, initiate checks
	app.respond(w, "{}", http.StatusOK)
}
