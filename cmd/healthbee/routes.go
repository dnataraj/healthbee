package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (app *application) routes() http.Handler {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/sites", app.monitor).Methods(http.MethodPost, http.MethodGet)
	r.HandleFunc("/sites/{id}/stop", app.stop).Methods(http.MethodPost)
	r.HandleFunc("/sites/{id}", app.getMetrics).Methods(http.MethodGet)

	r.HandleFunc("/ping", app.ping).Methods(http.MethodGet)

	return r
}
