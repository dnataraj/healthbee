package pkg

import (
	"github.com/segmentio/kafka-go"
	"os"
	"testing"
)

type config struct {
	endpoint    string
	srvCertPath string
	srvKeyPath  string
	caPath      string
}

func newKafkaConfig() *config {
	cfg := &config{}
	cfg.endpoint = os.Getenv("HB_TEST_KAFKA_SERVICE")
	cfg.srvCertPath = os.Getenv("HB_TEST_KAFKA_CERT_PATH")
	cfg.srvKeyPath = os.Getenv("HB_TEST_KAFKA_CERT_KEY")
	cfg.caPath = os.Getenv("HB_TEST_CA_CERT_PATH")

	return cfg
}

func newWriter(t *testing.T) (*kafka.Writer, func(tt *testing.T)) {
	config := newKafkaConfig()
	tlsConfig, err := GetTLSConfig(config.srvCertPath, config.srvKeyPath, config.caPath)
	if err != nil {
		t.Fatal(err)
	}
	err = CreateTopic("Metrics-Test", []string{config.endpoint}, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	w := &kafka.Writer{
		Addr:  kafka.TCP(config.endpoint),
		Topic: "Metrics-Test", RequiredAcks: kafka.RequireOne,
		Transport: &kafka.Transport{TLS: tlsConfig},
	}

	return w, func(tt *testing.T) {
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
		// clean up topic
		err = DeleteTopic("Metrics-Test", []string{config.endpoint}, tlsConfig)
		if err != nil {
			t.Fatal(err)
		}
	}

}
