package events

import (
	"context"

	"github.com/IBM/sarama"
)

func NewKafkaProducer(brokers []string, clientID string) (Producer, func() error, error) {
	cfg := sarama.NewConfig()
	cfg.ClientID = clientID
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 3

	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, nil, err
	}
	return &producer{
			Sync: p,
		}, func() error {
			return p.Close()
		}, nil
}

type producer struct {
	Sync sarama.SyncProducer
}

func (p *producer) Publish(ctx context.Context, topic string, payload []byte) error {
	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(payload)}
	_, _, err := p.Sync.SendMessage(msg)
	return err
}
