package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/dnataraj/healthbee/pkg"
	"github.com/dnataraj/healthbee/pkg/models/postgres"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger

	sites   *postgres.SiteModel
	results *postgres.ResultModel

	monitors map[int]*pkg.Monitor
	writer   *kafka.Writer
	wg       *sync.WaitGroup
	sync.Mutex
}

func main() {
	local := flag.Bool("local", false, "Set for local development mode")
	reset := flag.Bool("reset", false, "Clear all site registrations and metrics")
	addr := flag.String("addr", ":8000", "HTTP network address")
	brokerList := flag.String("brokers", "localhost:9092", "Comma separated distributed cache peers")
	dsn := flag.String("dsn", "", "DSN/Connection string for the PostgreSQL database")
	srvCertPath := flag.String("service-cert", "./certs/kafka/service.cert", "Path to the service public certificate")
	srvKeyPath := flag.String("service-key", "./certs/kafka/service.key", "Path to the private key")
	caPath := flag.String("ca-cert", "./certs/kafka/ca.pem", "Path to the CA certificate")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := pkg.OpenDB(*dsn)
	if err != nil {
		errorLog.Fatal("server: unable to connect to sites database: ", err.Error())
	}
	defer db.Close()

	if *reset {
		// re-initialize the database
		script, err := ioutil.ReadFile("pkg/models/postgres/testdata/setup.sql")
		if err != nil {
			errorLog.Fatal("server: unable to reset HealthBee database: ", err.Error())
		}
		infoLog.Println("server: HealthBee database reset requested, performing...")
		_, err = db.Exec(string(script))
		if err != nil {
			errorLog.Fatal("server: unable to reset HealthBee database: ", err.Error())
		}
	}

	// initialize TLS config for non-local services
	var tlsConfig *tls.Config
	if !*local {
		infoLog.Println("server: configuring TLS for kafka service...")
		tlsConfig, err = pkg.GetTLSConfig(*srvCertPath, *srvKeyPath, *caPath)
		if err != nil {
			errorLog.Fatal("server: error initializing kafka dialer: ", err.Error())
		}
	}

	brokers := strings.Split(*brokerList, ",")
	err = pkg.CreateTopic("Metrics", brokers, tlsConfig)
	if err != nil {
		errorLog.Fatal("server: error creating metrics topic on cluster: ", err.Error())
	}

	w := &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: "Metrics", RequiredAcks: kafka.RequireAll,
		Transport: &kafka.Transport{TLS: tlsConfig},
	}

	wg := sync.WaitGroup{}
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		sites:    &postgres.SiteModel{DB: db},
		results:  &postgres.ResultModel{DB: db},
		monitors: make(map[int]*pkg.Monitor),
		writer:   w,
		wg:       &wg,
	}

	srv := &http.Server{
		Addr:     ":8000",
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start the auditors
	infoLog.Println("server: creating readers for incoming metrics...")
	dialer := &kafka.Dialer{Timeout: 10 * time.Second, TLS: tlsConfig}
	r1 := newReader(brokers, dialer)
	r2 := newReader(brokers, dialer)
	infoLog.Println("auditor: starting 2 readers for incoming metrics...")
	wg.Add(1)
	go app.read(ctx, 1, r1, &wg)
	wg.Add(1)
	go app.read(ctx, 2, r2, &wg)

	infoLog.Printf("starting HealthBee API server on %s", *addr)
	wg.Add(1)
	go webServer(ctx, srv, &wg)

	// trap signals for clean shutdown and wait for all monitors to wrap up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	infoLog.Println("server: shutting down HTTP and auditor services...")
	cancel()

	// cancel all monitors
	infoLog.Println("server: shutting down HealthBee monitors...")
	for _, monitors := range app.monitors {
		monitors.Cancel()
	}

	wg.Wait()
	infoLog.Println("server: all monitors stopped, exiting.")
}

// webServer starts the main HTTP service using the provided http.Server configuration.
// Based on the article about connection draining found here: https://tylerchr.blog/golang-18-whats-coming/
// A context and wait group is provided to handle graceful shutdown and notification
func webServer(ctx context.Context, srv *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()
	// graceful shutdown, if stop is signalled from main routine
	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shCtx)
}

func newReader(brokers []string, dialer *kafka.Dialer) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "message-reader-group",
		Topic:   "Metrics",
		Dialer:  dialer,
	})
}
