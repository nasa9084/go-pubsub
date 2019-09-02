package pubsub

import (
	"context"
	"fmt"
	"sync"
)

type PubSub struct {
	mu sync.Mutex

	c chan Message

	q   []Message
	qmu sync.Mutex

	s []chan Message

	logger Logger
}

type Message struct {
	Topic string
	Value interface{}
}

func (m Message) String() string {
	return fmt.Sprintf("[%s]\t%+v", m.Topic, m.Value)
}

type Logger interface {
	Print(...interface{})
}

type nopLogger struct{}

func (nopLogger) Print(...interface{}) {}

func New() *PubSub {
	ps := &PubSub{
		c:      make(chan Message),
		logger: nopLogger{},
	}
	go func() {
		for m := range ps.c {
			for _, c := range ps.s {
				go func(m Message) {
					c <- m
				}(m)
			}
		}
	}()
	return ps
}

func (ps *PubSub) SetLogger(logger Logger) {
	ps.mu.Lock()
	ps.logger = logger
	ps.mu.Unlock()
}

func (ps *PubSub) Close() {
	ps.logger.Print("pubsub closed")
	close(ps.c)
}

func (ps *PubSub) Pub(topic string, v interface{}) {
	m := Message{
		Topic: topic,
		Value: v,
	}
	ps.logger.Print(m.String())
	ps.c <- m
}

func (ps *PubSub) Sub(topics ...string) (c <-chan Message, unsub func()) {
	send := make(chan Message)
	recv := make(chan Message)
	ps.mu.Lock()
	ps.s = append(ps.s, recv)
	idx := len(ps.s)
	ps.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	go filter(ctx, recv, send, topics)

	return send, func() {
		var s []chan Message
		if len(ps.s) == 1 {
			s = []chan Message{}
		} else {
			n := append(ps.s[:idx], ps.s[idx+1:]...)
			s = make([]chan Message, len(n))
		}
		ps.mu.Lock()
		ps.s = s
		ps.mu.Unlock()
		close(send)
		cancel()
	}
}

func filter(ctx context.Context, from, to chan Message, topics []string) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-from:
			for _, topic := range topics {
				if topic == m.Topic {
					to <- m
				}
			}
		}
	}
}
