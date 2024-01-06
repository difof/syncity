package event

import (
	"context"
	"testing"
	"time"

	"github.com/difof/errors"
)

type testEvent struct{}

func (e testEvent) Topic() string {
	return "test"
}

func TestBus(t *testing.T) {
	b := NewBus()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	b.Subscribe(ctx, "test", func(e Event) (err error) {
		defer errors.Recover(&err)

		t.Logf("1 received event %s", e.Topic())

		return
	})

	b.Subscribe(ctx, "test", func(e Event) (err error) {
		defer errors.Recover(&err)

		t.Logf("2 received event %s", e.Topic())

		return
	})

	if err := b.Publish("test", &testEvent{}); err != nil {
		t.Fatal(err)
	}

	<-ctx.Done()
}
