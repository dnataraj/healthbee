package main

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/dnataraj/healthbee/pkg"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
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
	local := flag.Bool("local", false, "Set for local development mode")
	brokerList := flag.String("brokers", "localhost:9092", "Comma separated distributed cache peers")
	dsn := flag.String("dsn", "", "DSN/Connection string for the PostgreSQL database")
	srvCertPath := flag.String("service-cert", "./certs/kafka/service.cert", "Path to the service public certificate")
	srvKeyPath := flag.String("service-key", "./certs/kafka/service.key", "Path to the private key")
	caPath := flag.String("ca-cert", "./certs/kafka/ca.pem", "Path to the CA certificate")
	flag.Parse()

	db, err := pkg.OpenDB(*dsn)
	if err != nil {
		errorLog.Fatal("unable to connect to sites database: ", err.Error())
	}
	defer db.Close()

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
		errorLog.Fatal("error creating metrics topic on cluster: ", err.Error())
	}

	dialer := &kafka.Dialer{Timeout: 10 * time.Second, TLS: tlsConfig}
	infoLog.Println("auditor: creating readers for incoming metrics...")
	r1 := newReader(brokers, dialer)
	r2 := newReader(brokers, dialer)

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	infoLog.Println("auditor: starting 2 readers for incoming metrics...")
	wg.Add(1)
	go read(ctx, 1, r1, &wg)
	wg.Add(1)
	go read(ctx, 2, r2, &wg)

	// trap signals for clean shutdown and wait for all monitors to wrap up
	infoLog.Println("auditor: services up and running...")
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

func newReader(brokers []string, dialer *kafka.Dialer) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "message-reader-group",
		Topic:   "Metrics",
		Dialer:  dialer,
	})
}
