package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dnataraj/healthbee/pkg"
	"github.com/dnataraj/healthbee/pkg/models"
	"net/http"
	"runtime/debug"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

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

func (app *application) NewMonitor(s *models.Site) *pkg.Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	m := &pkg.Monitor{
		Site:    s,
		Context: ctx,
		Cancel:  cancel,
	}
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	app.monitors[s.ID] = m

	return m
}
