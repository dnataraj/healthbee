package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dnataraj/healthbee/pkg"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

var infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

func read(ctx context.Context, id int, r *kafka.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				errorLog.Printf("auditor %d: unable to read message from cluster: %s", id, err.Error())
				return
			}
			infoLog.Printf("auditor %d: fetched result: %s", id, string(msg.Value))
		}
	}
}

func main() {
	brokerList := flag.String("brokers", "localhost:9092", "Comma separated distributed cache peers")
	dbpass := flag.String("password", "", "Password for the sites database")
	flag.Parse()

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

	r1 := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "message-reader-group",
		Topic:   "Metrics",
	})
	r2 := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "message-reader-group",
		Topic:   "Metrics",
	})

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go read(ctx, 1, r1, &wg)
	wg.Add(1)
	go read(ctx, 2, r2, &wg)

	// trap signals for clean shutdown and wait for all monitors to wrap up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	infoLog.Println("auditor: shutting down services...")
	cancel()

	// cancel all monitors
	infoLog.Println("auditor: waiting for metrics processors...")
	wg.Wait()

	infoLog.Println("auditor: exiting.")
}
