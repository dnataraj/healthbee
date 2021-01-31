package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dnataraj/healthbee/pkg"
	"github.com/dnataraj/healthbee/pkg/models/postgres"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
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

	sites    *postgres.SiteModel
	monitors map[int]*pkg.Monitor
	writer   *kafka.Writer
	wg       *sync.WaitGroup
	sync.Mutex
}

func main() {
	addr := flag.String("addr", ":8000", "HTTP network address")
	brokerList := flag.String("brokers", "localhost:9092", "Comma separated distributed cache peers")
	dbpass := flag.String("password", "", "Password for the sites database")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	connstr := fmt.Sprintf("postgres://postgres:%s@localhost/sites?sslmode=require", *dbpass)
	db, err := pkg.OpenDB(connstr)
	if err != nil {
		errorLog.Fatal("unable to connect to sites database: ", err.Error())
	}
	defer db.Close()

	brokers := strings.Split(*brokerList, ",")
	err = pkg.CreateTopic("Metrics", brokers)
	if err != nil {
		errorLog.Fatal("error creating metrics topic on cluster: ", err.Error())
	}
	w := &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: "Metrics", RequiredAcks: kafka.RequireAll,
	}

	wg := sync.WaitGroup{}
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		sites:    &postgres.SiteModel{DB: db},
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

	infoLog.Printf("starting server on %s", *addr)
	wg.Add(1)
	go webServer(ctx, srv, &wg)

	// trap signals for clean shutdown and wait for all monitors to wrap up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	infoLog.Println("server: shutting down HTTP services...")
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
