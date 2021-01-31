package pkg

import (
	"database/sql"
	"github.com/segmentio/kafka-go"
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

func CreateTopic(name string, brokers []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	topicConfigs := []kafka.TopicConfig{
		kafka.TopicConfig{
			Topic:             "Metrics",
			NumPartitions:     4,
			ReplicationFactor: 2,
		},
	}
	return conn.CreateTopics(topicConfigs...)
}
