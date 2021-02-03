package pkg

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type Procedure struct {
	URL      string
	Method   string
	Response Response
}

func NewTestServer(t *testing.T, addr string, procs ...Procedure) (*httptest.Server, func()) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range procs {
			if p.URL == r.URL.String() && p.Method == r.Method {
				code := p.Response.StatusCode
				if code == 0 {
					code = http.StatusOK
				}

				w.WriteHeader(code)
				_, err := w.Write(p.Response.Body)
				if err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		return
	})

	m := http.NewServeMux()
	m.HandleFunc("/test/", handler)

	ts := httptest.NewUnstartedServer(nil)
	ts.Config = &http.Server{Handler: m}
	ts.Listener, _ = net.Listen("tcp", addr)

	return ts, func() {
		ts.Close()
	}
}
