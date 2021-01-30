package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/dnataraj/healthbee/pkg/models/postgres"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger

	sites *postgres.SiteModel
}

func produce(ctx context.Context, w *kafka.Writer, log *log.Logger) {
	i := 0
	for {
		err := w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(fmt.Sprintf("message-%d", i)),
			Value: []byte(fmt.Sprintf("This is message %d", i)),
		})
		if err != nil {
			log.Fatal("unable to write message to kafka cluster: ", err.Error())
		}
		i++
		time.Sleep(time.Second)
	}

}

func consume(ctx context.Context, r *kafka.Reader, log *log.Logger) {
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Fatal("unable to read message from cluster: ", err.Error())
		}

		log.Println("fetched message: ", string(msg.Value))
	}
}

func main() {
	addr := flag.String("addr", ":8000", "HTTP network address")
	brokerList := flag.String("brokers", "localhost:9092", "Comma separated distributed cache peers")
	dbpass := flag.String("password", "", "Password for the sites database")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	connstr := fmt.Sprintf("postgres://postgres:%s@localhost/sites?sslmode=require", *dbpass)
	db, err := openDB(connstr)
	if err != nil {
		errorLog.Fatal("unable to connect to sites database: ", err.Error())
	}
	defer db.Close()

	brokers := strings.Split(*brokerList, ",")
	//ctx := context.Background()
	//conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], "Metrics", 0)
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		errorLog.Fatal("error connecting to cluster: ", err.Error())
	}
	defer conn.Close()
	topicConfigs := []kafka.TopicConfig{
		kafka.TopicConfig{
			Topic:             "Metrics",
			NumPartitions:     4,
			ReplicationFactor: 2,
		},
	}
	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		errorLog.Fatal("error creating Metrics topic on cluster: ", err.Error())
	}
	/*
		w := &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: "Metrics", RequiredAcks: kafka.RequireAll,
		}

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: "message-reader-group",
			Topic:   "Metrics",
		})

	*/

	//go produce(ctx, w, errorLog)
	//consume(ctx, r, infoLog)
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		sites:    &postgres.SiteModel{DB: db},
	}

	srv := &http.Server{
		Addr:     ":8000",
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
