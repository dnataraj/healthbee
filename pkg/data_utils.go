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
	dialer := &kafka.Dialer{Timeout: 10 * time.Second, TLS: config}
	conn, err := dialer.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	ctlr, err := conn.Controller()
	if err != nil {
		return err
	}
	ctlrConn, err := dialer.Dial("tcp", net.JoinHostPort(ctlr.Host, strconv.Itoa(ctlr.Port)))
	if err != nil {
		return err
	}
	defer ctlrConn.Close()

	topicConfigs := []kafka.TopicConfig{
		kafka.TopicConfig{
			Topic:             "Metrics",
			NumPartitions:     4,
			ReplicationFactor: 2,
		},
	}
	return ctlrConn.CreateTopics(topicConfigs...)
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
