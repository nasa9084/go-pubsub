package pubsub_test

import (
	"testing"

	pubsub "github.com/nasa9084/go-pubsub"
)

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Print(v ...interface{}) {
	l.t.Log(v...)
}

func TestPubSub(t *testing.T) {
	ps := pubsub.New()
	ps.SetLogger(&testLogger{t})

	c, unsub := ps.Sub("topic")
	defer unsub()

	go func() {
		ps.Pub("notSub", "hi, world")
		ps.Pub("topic", "hello, world")
	}()

	msg := <-c
	if msg.Topic != "topic" {
		t.Errorf("")
		return
	}
	if s, ok := msg.Value.(string); !ok || s != "hello, world" {
		t.Errorf("")
		return
	}
}
