package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	return p.conn.Publish(subject, data)
}
