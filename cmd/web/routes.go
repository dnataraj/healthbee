package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (app *application) routes() http.Handler {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/monitor", app.createSite).Methods(http.MethodPost)
	r.HandleFunc("/ping", app.ping).Methods(http.MethodGet)

	return r
}
