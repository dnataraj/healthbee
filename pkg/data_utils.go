package pkg

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTopic(name string, brokers []string, config *tls.Config) error {
	ctlrConn, err := getController(brokers, config)
	if err != nil {
		return err
	}
	defer ctlrConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             name,
			NumPartitions:     4,
			ReplicationFactor: 2,
		},
	}
	return ctlrConn.CreateTopics(topicConfigs...)
}

func DeleteTopic(name string, brokers []string, config *tls.Config) error {
	ctlrConn, err := getController(brokers, config)
	if err != nil {
		return err
	}
	defer ctlrConn.Close()

	return ctlrConn.DeleteTopics(name)
}

func getController(brokers []string, config *tls.Config) (*kafka.Conn, error) {
	dialer := &kafka.Dialer{Timeout: 10 * time.Second, TLS: config}
	conn, err := dialer.Dial("tcp", brokers[0])
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ctlr, err := conn.Controller()
	if err != nil {
		return nil, err
	}
	return dialer.Dial("tcp", net.JoinHostPort(ctlr.Host, strconv.Itoa(ctlr.Port)))
}

func GetTLSConfig(srvCertPath, keyPath, caPath string) (*tls.Config, error) {
	srvCert, err := tls.LoadX509KeyPair(srvCertPath, keyPath)
	if err != nil {
		return nil, err
	}
	cac, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(cac)

	return &tls.Config{
			Certificates: []tls.Certificate{srvCert},
			RootCAs:      caPool,
		},
		nil
}

func NewReader(brokers []string, dialer *kafka.Dialer) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "message-reader-group",
		Topic:   "Metrics",
		Dialer:  dialer,
	})
}
